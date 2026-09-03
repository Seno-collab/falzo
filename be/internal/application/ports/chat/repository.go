package chatports

import (
	domainchat "be/internal/domain/chat"
	"context"
	"time"
)

type Repository interface {
	Save(ctx context.Context, message domainchat.Message) error
	ListRoom(ctx context.Context, roomID string, before *time.Time, limit int) ([]domainchat.Message, error)
}
