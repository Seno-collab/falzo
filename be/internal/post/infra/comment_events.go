package infra

import (
	"context"
	"encoding/json"
	"time"

	"falzo-be/internal/post"
	pkgcache "falzo-be/pkg/cache"

	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const commentCreatedChannel = "post:comments:created"

type RedisCommentEventPublisher struct {
	client   *goredis.Client
	fallback post.CommentEventPublisher
}

func NewRedisCommentEventPublisher(cache pkgcache.Client, fallback post.CommentEventPublisher) post.CommentEventPublisher {
	if cache == nil || cache.Client() == nil {
		return fallback
	}

	return &RedisCommentEventPublisher{
		client:   cache.Client(),
		fallback: fallback,
	}
}

func (p *RedisCommentEventPublisher) PublishCommentCreated(ctx context.Context, comment post.CommentView) error {
	if p == nil || p.client == nil {
		if p != nil && p.fallback != nil {
			return p.fallback.PublishCommentCreated(ctx, comment)
		}
		return nil
	}

	payload, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	if err := p.client.Publish(ctx, commentCreatedChannel, payload).Err(); err != nil {
		if p.fallback != nil {
			_ = p.fallback.PublishCommentCreated(ctx, comment)
		}
		return err
	}

	return nil
}

func RunRedisCommentEventSubscriber(ctx context.Context, broker *post.CommentEventBroker, cache pkgcache.Client) {
	if broker == nil || cache == nil || cache.Client() == nil {
		return
	}

	pubsub := cache.Client().Subscribe(ctx, commentCreatedChannel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		if ctx.Err() == nil {
			log.Error().Err(err).Msg("comment event subscription setup failed")
		}
		return
	}

	channel := pubsub.Channel(
		goredis.WithChannelSize(100),
		goredis.WithChannelHealthCheckInterval(30*time.Second),
	)

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-channel:
			if !ok {
				return
			}

			var comment post.CommentView
			if err := json.Unmarshal([]byte(message.Payload), &comment); err != nil {
				log.Error().Err(err).Msg("comment event payload decode failed")
				continue
			}

			broker.BroadcastCommentCreated(comment)
		}
	}
}

var _ post.CommentEventPublisher = (*RedisCommentEventPublisher)(nil)
