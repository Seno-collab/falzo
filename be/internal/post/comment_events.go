package post

import (
	"context"
	"sync"
)

type CommentEventPublisher interface {
	PublishCommentCreated(ctx context.Context, comment CommentView) error
	PublishCommentUpdated(ctx context.Context, comment CommentView) error
}

type CommentEventSubscriber interface {
	SubscribeComments(ctx context.Context, postID uint64) (<-chan CommentEvent, func())
}

type CommentEvent struct {
	Type    string
	Comment CommentView
}

type PostEventPublisher interface {
	PublishPostCreated(ctx context.Context, post PostView) error
	PublishPostDeleted(ctx context.Context, postID uint64) error
	PublishUserAvatarUpdated(ctx context.Context, event UserAvatarUpdatedEvent) error
}

type PostEventSubscriber interface {
	SubscribePosts(ctx context.Context) (<-chan PostEvent, func())
}

type PostEvent struct {
	Type       string
	Post       PostView
	UserAvatar UserAvatarUpdatedEvent
}

type UserAvatarUpdatedEvent struct {
	UserID         uint64 `json:"user_id"`
	AvatarURL      string `json:"avatar_url"`
	AvatarURLAlias string `json:"avatarUrl"`
}

type CommentEventBroker struct {
	mu          sync.RWMutex
	subscribers map[uint64]map[chan CommentEvent]struct{}
}

func NewCommentEventBroker() *CommentEventBroker {
	return &CommentEventBroker{
		subscribers: make(map[uint64]map[chan CommentEvent]struct{}),
	}
}

func (b *CommentEventBroker) PublishCommentCreated(_ context.Context, comment CommentView) error {
	b.BroadcastCommentCreated(comment)
	return nil
}

func (b *CommentEventBroker) PublishCommentUpdated(_ context.Context, comment CommentView) error {
	b.BroadcastCommentUpdated(comment)
	return nil
}

func (b *CommentEventBroker) SubscribeComments(ctx context.Context, postID uint64) (<-chan CommentEvent, func()) {
	ch := make(chan CommentEvent, 16)

	b.mu.Lock()
	if b.subscribers[postID] == nil {
		b.subscribers[postID] = make(map[chan CommentEvent]struct{})
	}
	b.subscribers[postID][ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()

			if subscribers := b.subscribers[postID]; subscribers != nil {
				delete(subscribers, ch)
				if len(subscribers) == 0 {
					delete(b.subscribers, postID)
				}
			}
			close(ch)
		})
	}

	go func() {
		<-ctx.Done()
		unsubscribe()
	}()

	return ch, unsubscribe
}

func (b *CommentEventBroker) BroadcastCommentCreated(comment CommentView) {
	b.broadcastCommentEvent("comment.created", comment)
}

func (b *CommentEventBroker) BroadcastCommentUpdated(comment CommentView) {
	b.broadcastCommentEvent("comment.updated", comment)
}

func (b *CommentEventBroker) broadcastCommentEvent(eventType string, comment CommentView) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers[comment.PostID] {
		event := CommentEvent{Type: eventType, Comment: comment}
		select {
		case ch <- event:
		default:
		}
	}
}

var _ CommentEventPublisher = (*CommentEventBroker)(nil)
var _ CommentEventSubscriber = (*CommentEventBroker)(nil)

type PostEventBroker struct {
	mu          sync.RWMutex
	subscribers map[chan PostEvent]struct{}
}

func NewPostEventBroker() *PostEventBroker {
	return &PostEventBroker{subscribers: make(map[chan PostEvent]struct{})}
}

func (b *PostEventBroker) PublishPostCreated(_ context.Context, post PostView) error {
	b.BroadcastPostCreated(post)
	return nil
}

func (b *PostEventBroker) PublishPostDeleted(_ context.Context, postID uint64) error {
	b.BroadcastPostDeleted(postID)
	return nil
}

func (b *PostEventBroker) PublishUserAvatarUpdated(_ context.Context, event UserAvatarUpdatedEvent) error {
	b.BroadcastUserAvatarUpdated(event)
	return nil
}

func (b *PostEventBroker) SubscribePosts(ctx context.Context) (<-chan PostEvent, func()) {
	ch := make(chan PostEvent, 16)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			defer b.mu.Unlock()

			delete(b.subscribers, ch)
			close(ch)
		})
	}

	go func() {
		<-ctx.Done()
		unsubscribe()
	}()

	return ch, unsubscribe
}

func (b *PostEventBroker) BroadcastPostCreated(post PostView) {
	b.broadcastPostEvent("post.created", post)
}

func (b *PostEventBroker) BroadcastPostDeleted(postID uint64) {
	b.broadcastPostEvent("post.deleted", PostView{ID: postID})
}

func (b *PostEventBroker) BroadcastUserAvatarUpdated(event UserAvatarUpdatedEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		postEvent := PostEvent{Type: "user.avatar_updated", UserAvatar: event}
		select {
		case ch <- postEvent:
		default:
		}
	}
}

func (b *PostEventBroker) broadcastPostEvent(eventType string, post PostView) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		event := PostEvent{Type: eventType, Post: post}
		select {
		case ch <- event:
		default:
		}
	}
}

var _ PostEventPublisher = (*PostEventBroker)(nil)
var _ PostEventSubscriber = (*PostEventBroker)(nil)
