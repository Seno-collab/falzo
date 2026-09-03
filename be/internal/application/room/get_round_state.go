package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"be/internal/shared/clock"
	"context"
	"strings"
)

type GetRoundStateInput struct {
	RoomID string
	UserID int64
}

type GetRoundStateUseCase struct {
	rooms roomports.Repository
	clock clock.Clock
}

func NewGetRoundStateUseCase(rooms roomports.Repository, c clock.Clock) *GetRoundStateUseCase {
	return &GetRoundStateUseCase{rooms: rooms, clock: c}
}

func (uc *GetRoundStateUseCase) Execute(ctx context.Context, input GetRoundStateInput) (*domainroom.RoundState, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if !domainroom.IsValidID(roomID) {
		return nil, domainroom.ErrInvalidRoomID
	}
	return uc.rooms.FindCurrentRoundState(ctx, roomID, input.UserID, uc.clock.Now())
}
