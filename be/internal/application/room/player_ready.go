package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"be/internal/shared/clock"
	"context"
	"strings"
)

type PlayerReadyUseCase struct {
	rooms roomports.Repository
	clock clock.Clock
}

func NewPlayerReadyUseCase(rooms roomports.Repository, c clock.Clock) *PlayerReadyUseCase {
	return &PlayerReadyUseCase{rooms: rooms, clock: c}
}

func (uc *PlayerReadyUseCase) Execute(ctx context.Context, input GetRoundStateInput) (*domainroom.RoundState, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if !domainroom.IsValidID(roomID) {
		return nil, domainroom.ErrInvalidRoomID
	}
	return uc.rooms.MarkPlayerReady(ctx, roomports.PlayerActionInput{RoomID: roomID, UserID: input.UserID, At: uc.clock.Now()})
}
