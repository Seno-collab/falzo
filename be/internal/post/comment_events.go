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
}

type PostEventSubscriber interface {
	SubscribePosts(ctx context.Context) (<-chan PostView, func())
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
	subscribers map[chan PostView]struct{}
}

func NewPostEventBroker() *PostEventBroker {
	return &PostEventBroker{subscribers: make(map[chan PostView]struct{})}
}

func (b *PostEventBroker) PublishPostCreated(_ context.Context, post PostView) error {
	b.BroadcastPostCreated(post)
	return nil
}

func (b *PostEventBroker) SubscribePosts(ctx context.Context) (<-chan PostView, func()) {
	ch := make(chan PostView, 16)

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
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- post:
		default:
		}
	}
}

var _ PostEventPublisher = (*PostEventBroker)(nil)
var _ PostEventSubscriber = (*PostEventBroker)(nil)
