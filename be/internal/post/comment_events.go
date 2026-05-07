package post

import (
	"context"
	"sync"
)

type CommentEventPublisher interface {
	PublishCommentCreated(ctx context.Context, comment CommentView) error
}

type CommentEventSubscriber interface {
	SubscribeComments(ctx context.Context, postID uint64) (<-chan CommentView, func())
}

type CommentEventBroker struct {
	mu          sync.RWMutex
	subscribers map[uint64]map[chan CommentView]struct{}
}

func NewCommentEventBroker() *CommentEventBroker {
	return &CommentEventBroker{
		subscribers: make(map[uint64]map[chan CommentView]struct{}),
	}
}

func (b *CommentEventBroker) PublishCommentCreated(_ context.Context, comment CommentView) error {
	b.BroadcastCommentCreated(comment)
	return nil
}

func (b *CommentEventBroker) SubscribeComments(ctx context.Context, postID uint64) (<-chan CommentView, func()) {
	ch := make(chan CommentView, 16)

	b.mu.Lock()
	if b.subscribers[postID] == nil {
		b.subscribers[postID] = make(map[chan CommentView]struct{})
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
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers[comment.PostID] {
		select {
		case ch <- comment:
		default:
		}
	}
}

var _ CommentEventPublisher = (*CommentEventBroker)(nil)
var _ CommentEventSubscriber = (*CommentEventBroker)(nil)
