package notification

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TypeImageUploaded = "image.uploaded"
	TypePostCommented = "post.commented"

	ResourceImage   = "image"
	ResourceComment = "comment"
)

var (
	ErrUserIDRequired        = errors.New("notification user id is required")
	ErrNilHub                = errors.New("notification hub is nil")
	ErrDependencyUnavailable = errors.New("notification dependency unavailable")
	ErrInternal              = errors.New("notification internal error")
	ErrInvalidLimit          = errors.New("notification limit is invalid")
)

type Notification struct {
	ID          string    `json:"id"`
	UserID      uint64    `json:"user_id"`
	ActorUserID uint64    `json:"actor_user_id"`
	ActorName   string    `json:"actor_name"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Resource    string    `json:"resource"`
	ResourceID  string    `json:"resource_id"`
	PostID      uint64    `json:"post_id,omitempty"`
	ImageID     int64     `json:"image_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Publisher interface {
	Publish(ctx context.Context, item Notification) error
}

type Subscriber interface {
	Subscribe(ctx context.Context, userID uint64) (<-chan Notification, func(), error)
}

type Lister interface {
	ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error)
}

type Repository interface {
	Save(ctx context.Context, item Notification) error
	ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error)
}

type Hub struct {
	mu          sync.RWMutex
	sequence    atomic.Uint64
	subscribers map[uint64]map[chan Notification]struct{}
	repository  Repository
}

func NewHub(repositories ...Repository) *Hub {
	var repository Repository
	if len(repositories) > 0 {
		repository = repositories[0]
	}

	return &Hub{
		subscribers: make(map[uint64]map[chan Notification]struct{}),
		repository:  repository,
	}
}

func (h *Hub) Publish(ctx context.Context, item Notification) error {
	if h == nil {
		return ErrNilHub
	}
	if item.UserID == 0 {
		return ErrUserIDRequired
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}
	if item.ID == "" {
		item.ID = fmt.Sprintf("%d-%d", item.CreatedAt.UnixNano(), h.sequence.Add(1))
	}

	if h.repository != nil {
		if err := h.repository.Save(ctx, item); err != nil {
			return err
		}
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subscribers[item.UserID] {
		select {
		case ch <- item:
		default:
		}
	}

	return nil
}

func (h *Hub) Subscribe(ctx context.Context, userID uint64) (<-chan Notification, func(), error) {
	if h == nil {
		return nil, nil, ErrNilHub
	}
	if userID == 0 {
		return nil, nil, ErrUserIDRequired
	}

	ch := make(chan Notification, 32)

	h.mu.Lock()
	if h.subscribers[userID] == nil {
		h.subscribers[userID] = make(map[chan Notification]struct{})
	}
	h.subscribers[userID][ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()

			if subscribers := h.subscribers[userID]; subscribers != nil {
				delete(subscribers, ch)
				if len(subscribers) == 0 {
					delete(h.subscribers, userID)
				}
			}
			close(ch)
		})
	}

	go func() {
		<-ctx.Done()
		unsubscribe()
	}()

	return ch, unsubscribe, nil
}

func (h *Hub) ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error) {
	if h == nil {
		return nil, ErrNilHub
	}
	if userID == 0 {
		return nil, ErrUserIDRequired
	}
	if h.repository == nil {
		return []Notification{}, nil
	}

	return h.repository.ListByUser(ctx, userID, limit)
}

var _ Publisher = (*Hub)(nil)
var _ Subscriber = (*Hub)(nil)
var _ Lister = (*Hub)(nil)
