package realtime

import (
	"context"
	"encoding/json"
	"time"
)

const EventPresenceSync = "presence.sync"

type BackplaneMessage struct {
	RoomID  string          `json:"room_id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Backplane interface {
	Subscribe(ctx context.Context) (<-chan BackplaneMessage, error)
	Publish(ctx context.Context, roomID, eventType string, payload any) error
	TouchPresence(ctx context.Context, roomID, connectionID string, userID int64, expiresAt time.Time) error
	RemovePresence(ctx context.Context, roomID, connectionID string, userID int64) error
	OnlineUsers(ctx context.Context, roomID string, now time.Time) (map[int64]bool, error)
	Close() error
}

type PresenceSync struct {
	Members []Member `json:"members"`
}
