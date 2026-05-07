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

const (
	commentCreatedChannel = "post:comments:created"
	commentUpdatedChannel = "post:comments:updated"
	postCreatedChannel    = "posts:created"
)

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

func (p *RedisCommentEventPublisher) PublishCommentUpdated(ctx context.Context, comment post.CommentView) error {
	if p == nil || p.client == nil {
		if p != nil && p.fallback != nil {
			return p.fallback.PublishCommentUpdated(ctx, comment)
		}
		return nil
	}

	payload, err := json.Marshal(comment)
	if err != nil {
		return err
	}

	if err := p.client.Publish(ctx, commentUpdatedChannel, payload).Err(); err != nil {
		if p.fallback != nil {
			_ = p.fallback.PublishCommentUpdated(ctx, comment)
		}
		return err
	}

	return nil
}

func RunRedisCommentEventSubscriber(ctx context.Context, broker *post.CommentEventBroker, cache pkgcache.Client) {
	if broker == nil || cache == nil || cache.Client() == nil {
		return
	}

	pubsub := cache.Client().Subscribe(ctx, commentCreatedChannel, commentUpdatedChannel)
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

			switch message.Channel {
			case commentUpdatedChannel:
				broker.BroadcastCommentUpdated(comment)
			default:
				broker.BroadcastCommentCreated(comment)
			}
		}
	}
}

var _ post.CommentEventPublisher = (*RedisCommentEventPublisher)(nil)

type RedisPostEventPublisher struct {
	client   *goredis.Client
	fallback post.PostEventPublisher
}

func NewRedisPostEventPublisher(cache pkgcache.Client, fallback post.PostEventPublisher) post.PostEventPublisher {
	if cache == nil || cache.Client() == nil {
		return fallback
	}

	return &RedisPostEventPublisher{
		client:   cache.Client(),
		fallback: fallback,
	}
}

func (p *RedisPostEventPublisher) PublishPostCreated(ctx context.Context, item post.PostView) error {
	if p == nil || p.client == nil {
		if p != nil && p.fallback != nil {
			return p.fallback.PublishPostCreated(ctx, item)
		}
		return nil
	}

	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}

	if err := p.client.Publish(ctx, postCreatedChannel, payload).Err(); err != nil {
		if p.fallback != nil {
			_ = p.fallback.PublishPostCreated(ctx, item)
		}
		return err
	}

	return nil
}

func RunRedisPostEventSubscriber(ctx context.Context, broker *post.PostEventBroker, cache pkgcache.Client) {
	if broker == nil || cache == nil || cache.Client() == nil {
		return
	}

	pubsub := cache.Client().Subscribe(ctx, postCreatedChannel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		if ctx.Err() == nil {
			log.Error().Err(err).Msg("post event subscription setup failed")
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

			var item post.PostView
			if err := json.Unmarshal([]byte(message.Payload), &item); err != nil {
				log.Error().Err(err).Msg("post event payload decode failed")
				continue
			}

			broker.BroadcastPostCreated(item)
		}
	}
}

var _ post.PostEventPublisher = (*RedisPostEventPublisher)(nil)
