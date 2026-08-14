package redis

import (
	"be/internal/observability"
	"be/internal/realtime"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	realtimeSubscriptionBuffer = 256
	presenceKeyTTL             = 2 * time.Minute
)

type RealtimeBackplane struct {
	client  *goredis.Client
	prefix  string
	logger  *slog.Logger
	metrics *observability.Metrics
	mu      sync.Mutex
	pubsub  *goredis.PubSub
	closed  bool
}

func NewRealtimeBackplane(client *goredis.Client, prefix string, logger *slog.Logger, metrics *observability.Metrics) *RealtimeBackplane {
	if logger == nil {
		logger = slog.Default()
	}
	return &RealtimeBackplane{client: client, prefix: prefix, logger: logger, metrics: metrics}
}

func (b *RealtimeBackplane) Subscribe(ctx context.Context) (<-chan realtime.BackplaneMessage, error) {
	startedAt := time.Now()
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		b.observeOperation("subscribe", startedAt, fmt.Errorf("realtime backplane is closed"))
		return nil, fmt.Errorf("realtime backplane is closed")
	}
	if b.pubsub != nil {
		b.observeOperation("subscribe", startedAt, fmt.Errorf("realtime backplane is already subscribed"))
		return nil, fmt.Errorf("realtime backplane is already subscribed")
	}

	pubsub := b.client.Subscribe(ctx, b.channel())
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		b.observeOperation("subscribe", startedAt, err)
		return nil, fmt.Errorf("subscribe realtime channel: %w", err)
	}
	b.pubsub = pubsub
	b.observeOperation("subscribe", startedAt, nil)

	messages := make(chan realtime.BackplaneMessage, realtimeSubscriptionBuffer)
	go b.consume(ctx, pubsub.Channel(goredis.WithChannelSize(realtimeSubscriptionBuffer)), messages)
	return messages, nil
}

func (b *RealtimeBackplane) Publish(ctx context.Context, roomID string, event realtime.Event) error {
	startedAt := time.Now()
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		b.observeOperation("publish", startedAt, err)
		return fmt.Errorf("marshal realtime payload: %w", err)
	}
	messageJSON, err := json.Marshal(realtime.BackplaneMessage{
		EventID:    event.EventID,
		RoomID:     roomID,
		Type:       event.Type,
		RequestID:  event.RequestID,
		OccurredAt: event.OccurredAt,
		Payload:    payloadJSON,
	})
	if err != nil {
		b.observeOperation("publish", startedAt, err)
		return fmt.Errorf("marshal realtime message: %w", err)
	}
	err = b.client.Publish(ctx, b.channel(), messageJSON).Err()
	b.observeOperation("publish", startedAt, err)
	return err
}

func (b *RealtimeBackplane) ClaimConnection(
	ctx context.Context,
	roomID string,
	userID int64,
	connectionID string,
	expiresAt time.Time,
) (string, error) {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}
	result, err := b.client.SetArgs(ctx, b.connectionKey(roomID, userID), connectionID, goredis.SetArgs{
		TTL: ttl,
		Get: true,
	}).Result()
	if err == goredis.Nil {
		return "", nil
	}
	return result, err
}

func (b *RealtimeBackplane) RefreshConnection(
	ctx context.Context,
	roomID string,
	userID int64,
	connectionID string,
	expiresAt time.Time,
) (bool, error) {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}
	result, err := b.client.Eval(ctx, `
		if redis.call('GET', KEYS[1]) == ARGV[1] then
			return redis.call('PEXPIRE', KEYS[1], ARGV[2])
		end
		return 0`, []string{b.connectionKey(roomID, userID)}, connectionID, ttl.Milliseconds()).Int()
	return result == 1, err
}

func (b *RealtimeBackplane) ReleaseConnection(
	ctx context.Context,
	roomID string,
	userID int64,
	connectionID string,
) error {
	return b.compareAndDelete(ctx, b.connectionKey(roomID, userID), connectionID)
}

func (b *RealtimeBackplane) ClaimRequest(
	ctx context.Context,
	roomID string,
	userID int64,
	requestID string,
	expiresAt time.Time,
) (bool, error) {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		ttl = time.Second
	}
	return b.client.SetNX(ctx, b.requestKey(roomID, userID, requestID), "1", ttl).Result()
}

func (b *RealtimeBackplane) TouchPresence(
	ctx context.Context,
	roomID,
	connectionID string,
	userID int64,
	expiresAt time.Time,
) error {
	_, err := b.client.Pipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.ZAdd(ctx, b.presenceKey(roomID), goredis.Z{
			Score:  float64(expiresAt.UnixMilli()),
			Member: presenceMember(connectionID, userID),
		})
		pipe.Expire(ctx, b.presenceKey(roomID), presenceKeyTTL)
		return nil
	})
	return err
}

func (b *RealtimeBackplane) RemovePresence(ctx context.Context, roomID, connectionID string, userID int64) error {
	return b.client.ZRem(ctx, b.presenceKey(roomID), presenceMember(connectionID, userID)).Err()
}

func (b *RealtimeBackplane) OnlineUsers(ctx context.Context, roomID string, now time.Time) (map[int64]bool, error) {
	var active *goredis.StringSliceCmd
	_, err := b.client.Pipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.ZRemRangeByScore(ctx, b.presenceKey(roomID), "-inf", strconv.FormatInt(now.UnixMilli(), 10))
		active = pipe.ZRangeByScore(ctx, b.presenceKey(roomID), &goredis.ZRangeBy{
			Min: strconv.FormatInt(now.UnixMilli()+1, 10),
			Max: "+inf",
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	users := make(map[int64]bool)
	for _, member := range active.Val() {
		separator := strings.LastIndexByte(member, ':')
		if separator < 0 {
			continue
		}
		userID, err := strconv.ParseInt(member[separator+1:], 10, 64)
		if err == nil {
			users[userID] = true
		}
	}
	return users, nil
}

func (b *RealtimeBackplane) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	b.closed = true
	if b.pubsub != nil {
		return b.pubsub.Close()
	}
	return nil
}

func (b *RealtimeBackplane) compareAndDelete(ctx context.Context, key, expected string) error {
	_, err := b.client.Eval(ctx, `
		if redis.call('GET', KEYS[1]) == ARGV[1] then
			return redis.call('DEL', KEYS[1])
		end
		return 0`, []string{key}, expected).Result()
	return err
}

func (b *RealtimeBackplane) consume(
	ctx context.Context,
	source <-chan *goredis.Message,
	destination chan<- realtime.BackplaneMessage,
) {
	defer close(destination)
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-source:
			if !ok {
				return
			}
			var decoded realtime.BackplaneMessage
			if err := json.Unmarshal([]byte(message.Payload), &decoded); err != nil {
				b.logger.Warn("invalid redis realtime message", slog.Any("error", err))
				continue
			}
			select {
			case destination <- decoded:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (b *RealtimeBackplane) channel() string {
	return b.prefix + ":realtime:events"
}

func (b *RealtimeBackplane) presenceKey(roomID string) string {
	return b.prefix + ":realtime:presence:" + roomID
}

func (b *RealtimeBackplane) connectionKey(roomID string, userID int64) string {
	return b.prefix + ":realtime:connection:" + roomID + ":" + strconv.FormatInt(userID, 10)
}

func (b *RealtimeBackplane) requestKey(roomID string, userID int64, requestID string) string {
	return b.prefix + ":realtime:request:" + roomID + ":" + strconv.FormatInt(userID, 10) + ":" + requestID
}

func presenceMember(connectionID string, userID int64) string {
	return connectionID + ":" + strconv.FormatInt(userID, 10)
}

func (b *RealtimeBackplane) observeOperation(operation string, startedAt time.Time, err error) {
	if b.metrics == nil {
		return
	}
	result := "success"
	if err != nil {
		result = "error"
	}
	b.metrics.RealtimeRedisOperations.WithLabelValues(operation, result).Inc()
	b.metrics.RealtimeRedisDuration.WithLabelValues(operation).Observe(time.Since(startedAt).Seconds())
}
