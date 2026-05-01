package infra

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"strconv"
	"strings"
	"time"

	"falzo-be/internal/post"
	pkgcache "falzo-be/pkg/cache"
	"falzo-be/pkg/config"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	engagementStream        = "post:engagement_events"
	engagementDeadStream    = "post:engagement_events:dead"
	engagementRetryHash     = "post:engagement_event_retries"
	engagementConsumerGroup = "post-engagement-db-writers"
	engagementActionLike    = "like"
	engagementActionSave    = "save"
)

var defaultEngagementConfig = config.EngagementConfig{
	ClaimMinIdle: 30 * time.Second,
	MaxRetries:   10,
}

type EngagementStreamRepository struct {
	next   post.Repository
	client *goredis.Client
}

func NewEngagementStreamRepository(next post.Repository, cache pkgcache.Client) post.Repository {
	if next == nil || cache == nil || cache.Client() == nil {
		return next
	}

	return &EngagementStreamRepository{next: next, client: cache.Client()}
}

func (r *EngagementStreamRepository) Create(ctx context.Context, item *post.Post) error {
	return r.next.Create(ctx, item)
}

func (r *EngagementStreamRepository) Like(ctx context.Context, postID uint64, userID uint64) error {
	return r.publishEngagement(ctx, engagementActionLike, postID, userID)
}

func (r *EngagementStreamRepository) Save(ctx context.Context, postID uint64, userID uint64) error {
	return r.publishEngagement(ctx, engagementActionSave, postID, userID)
}

func (r *EngagementStreamRepository) GetPosts(ctx context.Context, page int, limit int) ([]post.Post, error) {
	return r.next.GetPosts(ctx, page, limit)
}

func (r *EngagementStreamRepository) GetPostDetail(ctx context.Context, postID uint64) (*post.Post, error) {
	return r.next.GetPostDetail(ctx, postID)
}

func (r *EngagementStreamRepository) GetPostsByLocation(ctx context.Context, locationName post.LocationName) ([]post.Post, error) {
	return r.next.GetPostsByLocation(ctx, locationName)
}

func (r *EngagementStreamRepository) publishEngagement(ctx context.Context, action string, postID uint64, userID uint64) error {
	pipe := r.client.TxPipeline()
	switch action {
	case engagementActionLike:
		pipe.SAdd(ctx, postLikesKey(postID), strconv.FormatUint(userID, 10))
	case engagementActionSave:
		pipe.SAdd(ctx, postSavesKey(postID), strconv.FormatUint(userID, 10))
	}
	pipe.XAdd(ctx, &goredis.XAddArgs{
		Stream: engagementStream,
		Values: map[string]any{
			"action":  action,
			"post_id": postID,
			"user_id": userID,
		},
	})

	if _, err := pipe.Exec(ctx); err != nil {
		return post.ErrDependencyUnavailable
	}

	return nil
}

func RunEngagementStreamWorker(ctx context.Context, sink post.Repository, cache pkgcache.Client, cfg config.EngagementConfig) {
	if sink == nil || cache == nil || cache.Client() == nil {
		return
	}

	worker := engagementStreamWorker{
		sink:     sink,
		client:   cache.Client(),
		config:   normalizeEngagementConfig(cfg),
		consumer: engagementConsumerName(),
	}
	worker.run(ctx)
}

type engagementStreamWorker struct {
	sink     post.Repository
	client   *goredis.Client
	config   config.EngagementConfig
	consumer string
}

func (w engagementStreamWorker) run(ctx context.Context) {
	if w.consumer == "" {
		w.consumer = engagementConsumerName()
	}
	if err := w.ensureGroup(ctx); err != nil {
		log.Error().Err(err).Msg("post engagement stream group setup failed")
		return
	}

	for {
		if ctx.Err() != nil {
			return
		}

		if err := w.claimAndProcess(ctx); err != nil {
			w.handleReadError(ctx, err)
			continue
		}

		processedPending, err := w.readAndProcess(ctx, "0", 100*time.Millisecond)
		if err != nil {
			w.handleReadError(ctx, err)
			continue
		}
		if processedPending {
			continue
		}

		if _, err := w.readAndProcess(ctx, ">", time.Second); err != nil {
			w.handleReadError(ctx, err)
		}
	}
}

func (w engagementStreamWorker) ensureGroup(ctx context.Context) error {
	err := w.client.XGroupCreateMkStream(ctx, engagementStream, engagementConsumerGroup, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}

	return nil
}

func (w engagementStreamWorker) readAndProcess(ctx context.Context, streamID string, block time.Duration) (bool, error) {
	streams, err := w.client.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    engagementConsumerGroup,
		Consumer: w.consumer,
		Streams:  []string{engagementStream, streamID},
		Count:    50,
		Block:    block,
	}).Result()
	if err != nil {
		return false, err
	}

	processed := false
	for _, stream := range streams {
		for _, message := range stream.Messages {
			processed = true
			if err := w.processMessage(ctx, message); err != nil {
				w.handleProcessingFailure(ctx, message, err)
				continue
			}

			w.ackProcessed(ctx, message.ID)
		}
	}

	return processed, nil
}

func (w engagementStreamWorker) claimAndProcess(ctx context.Context) error {
	messages, _, err := w.client.XAutoClaim(ctx, &goredis.XAutoClaimArgs{
		Stream:   engagementStream,
		Group:    engagementConsumerGroup,
		Consumer: w.consumer,
		MinIdle:  w.config.ClaimMinIdle,
		Start:    "0-0",
		Count:    50,
	}).Result()
	if err != nil {
		return err
	}

	for _, message := range messages {
		if err := w.processMessage(ctx, message); err != nil {
			w.handleProcessingFailure(ctx, message, err)
			continue
		}

		w.ackProcessed(ctx, message.ID)
	}

	return nil
}

func (w engagementStreamWorker) processMessage(ctx context.Context, message goredis.XMessage) error {
	action := fmt.Sprint(message.Values["action"])
	postID, err := parseUintField(message.Values["post_id"])
	if err != nil {
		return err
	}
	userID, err := parseUintField(message.Values["user_id"])
	if err != nil {
		return err
	}

	switch action {
	case engagementActionLike:
		return w.sink.Like(ctx, postID, userID)
	case engagementActionSave:
		return w.sink.Save(ctx, postID, userID)
	default:
		return fmt.Errorf("unknown post engagement action %q", action)
	}
}

func (w engagementStreamWorker) handleReadError(ctx context.Context, err error) {
	if ctx.Err() != nil || errors.Is(err, goredis.Nil) {
		return
	}

	log.Error().Err(err).Msg("post engagement stream read failed")
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
	}
}

func (w engagementStreamWorker) handleProcessingFailure(ctx context.Context, message goredis.XMessage, cause error) {
	retries, err := w.client.HIncrBy(ctx, engagementRetryHash, message.ID, 1).Result()
	if err != nil {
		log.Error().Err(err).Str("message_id", message.ID).Msg("post engagement event retry count failed")
		return
	}

	if retries < int64(w.config.MaxRetries) {
		log.Error().
			Err(cause).
			Str("message_id", message.ID).
			Int64("retry", retries).
			Int("max_retries", w.config.MaxRetries).
			Msg("post engagement event processing failed")
		return
	}

	values := map[string]any{
		"original_id": message.ID,
		"error":       cause.Error(),
		"failed_at":   time.Now().UTC().Format(time.RFC3339Nano),
	}
	maps.Copy(values, message.Values)

	if err := w.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: engagementDeadStream,
		Values: values,
	}).Err(); err != nil {
		log.Error().Err(err).Str("message_id", message.ID).Msg("post engagement dead-letter publish failed")
		return
	}

	w.ackProcessed(ctx, message.ID)
	log.Error().
		Err(cause).
		Str("message_id", message.ID).
		Int("max_retries", w.config.MaxRetries).
		Msg("post engagement event moved to dead-letter stream")
}

func (w engagementStreamWorker) ackProcessed(ctx context.Context, messageID string) {
	if err := w.client.XAck(ctx, engagementStream, engagementConsumerGroup, messageID).Err(); err != nil {
		log.Error().Err(err).Str("message_id", messageID).Msg("post engagement event ack failed")
		return
	}

	if err := w.client.HDel(ctx, engagementRetryHash, messageID).Err(); err != nil {
		log.Error().Err(err).Str("message_id", messageID).Msg("post engagement retry cleanup failed")
	}
}

func parseUintField(value any) (uint64, error) {
	parsed, err := strconv.ParseUint(fmt.Sprint(value), 10, 64)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("invalid uint field %v", value)
	}

	return parsed, nil
}

func postLikesKey(postID uint64) string {
	return "post:" + strconv.FormatUint(postID, 10) + ":likes"
}

func postSavesKey(postID uint64) string {
	return "post:" + strconv.FormatUint(postID, 10) + ":saves"
}

func engagementConsumerName() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown-host"
	}

	return hostname + ":" + strconv.Itoa(os.Getpid())
}

func normalizeEngagementConfig(cfg config.EngagementConfig) config.EngagementConfig {
	if cfg.ClaimMinIdle <= 0 {
		cfg.ClaimMinIdle = defaultEngagementConfig.ClaimMinIdle
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = defaultEngagementConfig.MaxRetries
	}

	return cfg
}

var _ post.Repository = (*EngagementStreamRepository)(nil)
