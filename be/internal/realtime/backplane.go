package realtime

import (
	"context"
	"encoding/json"
	"time"
)

const EventPresenceSync = "presence.sync"

type BackplaneMessage struct {
	EventID    string          `json:"event_id"`
	RoomID     string          `json:"room_id"`
	Type       string          `json:"type"`
	RequestID  string          `json:"request_id,omitempty"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

type Backplane interface {
	Subscribe(ctx context.Context) (<-chan BackplaneMessage, error)
	Publish(ctx context.Context, roomID string, event Event) error
	ClaimConnection(ctx context.Context, roomID string, userID int64, connectionID string, expiresAt time.Time) (string, error)
	RefreshConnection(ctx context.Context, roomID string, userID int64, connectionID string, expiresAt time.Time) (bool, error)
	ReleaseConnection(ctx context.Context, roomID string, userID int64, connectionID string) error
	ClaimRequest(ctx context.Context, roomID string, userID int64, requestID string, expiresAt time.Time) (bool, error)
	TouchPresence(ctx context.Context, roomID, connectionID string, userID int64, expiresAt time.Time) error
	RemovePresence(ctx context.Context, roomID, connectionID string, userID int64) error
	OnlineUsers(ctx context.Context, roomID string, now time.Time) (map[int64]bool, error)
	Close() error
}

type PresenceSync struct {
	Members []Member `json:"members"`
}
