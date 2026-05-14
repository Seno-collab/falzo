package notification

import (
	"context"
	"testing"
	"time"
)

type fakeNotificationRepository struct {
	items []Notification
}

func (r *fakeNotificationRepository) Save(_ context.Context, item Notification) error {
	r.items = append(r.items, item)
	return nil
}

func (r *fakeNotificationRepository) ListByUser(_ context.Context, userID uint64, limit int) ([]Notification, error) {
	items := make([]Notification, 0, limit)
	for _, item := range r.items {
		if item.UserID == userID {
			items = append(items, item)
		}
	}
	if len(items) > limit {
		return items[:limit], nil
	}
	return items, nil
}

func (r *fakeNotificationRepository) MarkRead(_ context.Context, userID uint64, ids []string) error {
	readAt := time.Now().UTC()
	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	for index := range r.items {
		if r.items[index].UserID == userID {
			if _, ok := idSet[r.items[index].ID]; ok {
				r.items[index].ReadAt = &readAt
			}
		}
	}
	return nil
}

func TestHubPublishesOnlyToTargetUser(t *testing.T) {
	hub := NewHub()
	target, unsubscribeTarget, err := hub.Subscribe(t.Context(), 7)
	if err != nil {
		t.Fatalf("subscribe target: %v", err)
	}
	defer unsubscribeTarget()

	other, unsubscribeOther, err := hub.Subscribe(t.Context(), 8)
	if err != nil {
		t.Fatalf("subscribe other: %v", err)
	}
	defer unsubscribeOther()

	item := Notification{UserID: 7, Type: TypeImageUploaded, Title: "Image uploaded"}
	if err := hub.Publish(t.Context(), item); err != nil {
		t.Fatalf("publish notification: %v", err)
	}

	select {
	case got := <-target:
		if got.UserID != item.UserID || got.Type != item.Type || got.ID == "" || got.CreatedAt.IsZero() {
			t.Fatalf("unexpected notification: %+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("expected target subscriber to receive notification")
	}

	select {
	case got := <-other:
		t.Fatalf("expected other subscriber to stay idle, got %+v", got)
	case <-time.After(25 * time.Millisecond):
	}
}

func TestHubRejectsMissingUserID(t *testing.T) {
	hub := NewHub()
	if err := hub.Publish(t.Context(), Notification{Type: TypeImageUploaded}); err != ErrUserIDRequired {
		t.Fatalf("expected ErrUserIDRequired, got %v", err)
	}
}

func TestHubPersistsNotificationsForLaterListing(t *testing.T) {
	repository := &fakeNotificationRepository{}
	hub := NewHub(repository)

	item := Notification{UserID: 7, Type: TypePostCommented, Title: "New comment"}
	if err := hub.Publish(t.Context(), item); err != nil {
		t.Fatalf("publish notification: %v", err)
	}

	items, err := hub.ListByUser(t.Context(), 7, 30)
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	if len(items) != 1 || items[0].UserID != item.UserID || items[0].Type != item.Type || items[0].ID == "" {
		t.Fatalf("expected persisted notification, got %+v", items)
	}
}
