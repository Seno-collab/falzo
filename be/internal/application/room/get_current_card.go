package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"context"
	"strings"
)

type GetCurrentCardInput struct {
	RoomID string
	UserID int64
}

type GetCurrentCardUseCase struct {
	rooms roomports.Repository
}

func NewGetCurrentCardUseCase(rooms roomports.Repository) *GetCurrentCardUseCase {
	return &GetCurrentCardUseCase{rooms: rooms}
}

func (uc *GetCurrentCardUseCase) Execute(ctx context.Context, input GetCurrentCardInput) (*domainroom.RoundCard, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if !domainroom.IsValidID(roomID) {
		return nil, domainroom.ErrInvalidRoomID
	}
	if input.UserID <= 0 {
		return nil, domainroom.ErrRoundCardNotFound
	}
	return uc.rooms.FindCurrentCard(ctx, roomID, input.UserID)
}
