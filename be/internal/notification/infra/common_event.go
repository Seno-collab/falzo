package infra

import (
	"context"
	"encoding/json"
	"time"

	"falzo-be/internal/post"
	pkgcache "falzo-be/pkg/cache"
	"falzo-be/pkg/logger"

	goredis "github.com/redis/go-redis/v9"
)

const (
	commentCreatedChannel = "post:comments:created"
	commentUpdatedChannel = "post:comments:updated"
	postCreatedChannel    = "posts:created"
	postDeletedChannel    = "posts:deleted"
	userAvatarChannel     = "users:avatar_updated"
)

var commonEventLog = logger.For("notification.common_event")

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
			commonEventLog.Error(ctx, err, "comment event subscription setup failed")
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
				commonEventLog.Error(ctx, err, "comment event payload decode failed")
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
	return p.publishPostEvent(ctx, postCreatedChannel, item, func() {
		if p.fallback != nil {
			_ = p.fallback.PublishPostCreated(ctx, item)
		}
	})
}

func (p *RedisPostEventPublisher) PublishPostDeleted(ctx context.Context, postID uint64) error {
	return p.publishPostEvent(ctx, postDeletedChannel, post.PostView{ID: postID}, func() {
		if p.fallback != nil {
			_ = p.fallback.PublishPostDeleted(ctx, postID)
		}
	})
}

func (p *RedisPostEventPublisher) PublishUserAvatarUpdated(ctx context.Context, event post.UserAvatarUpdatedEvent) error {
	if p == nil || p.client == nil {
		if p != nil && p.fallback != nil {
			_ = p.fallback.PublishUserAvatarUpdated(ctx, event)
		}
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err := p.client.Publish(ctx, userAvatarChannel, payload).Err(); err != nil {
		if p.fallback != nil {
			_ = p.fallback.PublishUserAvatarUpdated(ctx, event)
		}
		return err
	}

	return nil
}

func (p *RedisPostEventPublisher) publishPostEvent(ctx context.Context, channel string, item post.PostView, fallback func()) error {
	if p == nil || p.client == nil {
		if fallback != nil {
			fallback()
		}
		return nil
	}

	payload, err := json.Marshal(item)
	if err != nil {
		return err
	}

	if err := p.client.Publish(ctx, channel, payload).Err(); err != nil {
		if fallback != nil {
			fallback()
		}
		return err
	}

	return nil
}

func RunRedisPostEventSubscriber(ctx context.Context, broker *post.PostEventBroker, cache pkgcache.Client) {
	if broker == nil || cache == nil || cache.Client() == nil {
		return
	}

	pubsub := cache.Client().Subscribe(ctx, postCreatedChannel, postDeletedChannel, userAvatarChannel)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		if ctx.Err() == nil {
			commonEventLog.Error(ctx, err, "post event subscription setup failed")
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

			switch message.Channel {
			case userAvatarChannel:
				var event post.UserAvatarUpdatedEvent
				if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
					commonEventLog.Error(ctx, err, "user avatar event payload decode failed")
					continue
				}
				broker.BroadcastUserAvatarUpdated(event)
			case postDeletedChannel:
				var item post.PostView
				if err := json.Unmarshal([]byte(message.Payload), &item); err != nil {
					commonEventLog.Error(ctx, err, "post event payload decode failed")
					continue
				}
				broker.BroadcastPostDeleted(item.ID)
			default:
				var item post.PostView
				if err := json.Unmarshal([]byte(message.Payload), &item); err != nil {
					commonEventLog.Error(ctx, err, "post event payload decode failed")
					continue
				}
				broker.BroadcastPostCreated(item)
			}
		}
	}
}

var _ post.PostEventPublisher = (*RedisPostEventPublisher)(nil)
