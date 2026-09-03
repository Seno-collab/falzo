package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"be/internal/shared/clock"
	"context"
	"strings"
)

type FinishTurnUseCase struct {
	rooms roomports.Repository
	clock clock.Clock
}

func NewFinishTurnUseCase(rooms roomports.Repository, c clock.Clock) *FinishTurnUseCase {
	return &FinishTurnUseCase{rooms: rooms, clock: c}
}

func (uc *FinishTurnUseCase) Execute(ctx context.Context, input GetRoundStateInput) (*domainroom.RoundState, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if !domainroom.IsValidID(roomID) {
		return nil, domainroom.ErrInvalidRoomID
	}
	return uc.rooms.FinishTurn(ctx, roomports.PlayerActionInput{RoomID: roomID, UserID: input.UserID, At: uc.clock.Now()})
}
