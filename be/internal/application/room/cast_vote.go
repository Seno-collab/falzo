package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"be/internal/shared/clock"
	"context"
	"strings"
)

type CastVoteInput struct {
	RoomID       string
	VoterUserID  int64
	TargetUserID int64
}

type CastVoteUseCase struct {
	rooms roomports.Repository
	clock clock.Clock
}

func NewCastVoteUseCase(rooms roomports.Repository, c clock.Clock) *CastVoteUseCase {
	return &CastVoteUseCase{rooms: rooms, clock: c}
}

func (uc *CastVoteUseCase) Execute(ctx context.Context, input CastVoteInput) (*domainroom.RoundState, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if !domainroom.IsValidID(roomID) {
		return nil, domainroom.ErrInvalidRoomID
	}
	if input.VoterUserID <= 0 || input.TargetUserID <= 0 || input.VoterUserID == input.TargetUserID {
		return nil, domainroom.ErrInvalidVoteTarget
	}
	return uc.rooms.CastVote(ctx, roomports.CastVoteInput{
		RoomID:       roomID,
		VoterUserID:  input.VoterUserID,
		TargetUserID: input.TargetUserID,
		VotedAt:      uc.clock.Now(),
	})
}
