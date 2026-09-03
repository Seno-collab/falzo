package chatapp

import (
	chatports "be/internal/application/ports/chat"
	domainchat "be/internal/domain/chat"
	"context"
	"time"
)

const (
	DefaultHistoryLimit = 50
	MaxHistoryLimit     = 100
)

type Service struct {
	repository chatports.Repository
}

func NewService(repository chatports.Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) SaveMessage(ctx context.Context, message domainchat.Message) error {
	return s.repository.Save(ctx, message)
}

func (s *Service) ListRoomMessages(
	ctx context.Context,
	roomID string,
	before *time.Time,
	limit int,
) ([]domainchat.Message, error) {
	if limit <= 0 {
		limit = DefaultHistoryLimit
	}
	if limit > MaxHistoryLimit {
		limit = MaxHistoryLimit
	}
	return s.repository.ListRoom(ctx, roomID, before, limit)
}
