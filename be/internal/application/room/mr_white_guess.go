package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"be/internal/shared/clock"
	"context"
	"strings"
)

type MrWhiteGuessInput struct {
	RoomID string
	UserID int64
	Guess  string
}

type MrWhiteGuessUseCase struct {
	rooms roomports.Repository
	clock clock.Clock
}

func NewMrWhiteGuessUseCase(rooms roomports.Repository, c clock.Clock) *MrWhiteGuessUseCase {
	return &MrWhiteGuessUseCase{rooms: rooms, clock: c}
}

func (uc *MrWhiteGuessUseCase) Execute(ctx context.Context, input MrWhiteGuessInput) (*domainroom.RoundState, error) {
	roomID := strings.TrimSpace(input.RoomID)
	guess := strings.TrimSpace(input.Guess)
	if !domainroom.IsValidID(roomID) {
		return nil, domainroom.ErrInvalidRoomID
	}
	if guess == "" || len([]rune(guess)) > 80 {
		return nil, domainroom.ErrMrWhiteGuessNotAllowed
	}
	return uc.rooms.SubmitMrWhiteGuess(ctx, roomports.MrWhiteGuessInput{
		RoomID: roomID, UserID: input.UserID, Guess: guess, At: uc.clock.Now(),
	})
}
