package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"context"
	"strings"
)

type KickMemberInput struct {
	RoomID       string
	HostUserID   int64
	TargetUserID int64
}

type KickMemberUseCase struct {
	rooms roomports.Repository
}

func NewKickMemberUseCase(rooms roomports.Repository) *KickMemberUseCase {
	return &KickMemberUseCase{rooms: rooms}
}

func (uc *KickMemberUseCase) Execute(ctx context.Context, input KickMemberInput) (*domainroom.Room, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if !domainroom.IsValidID(roomID) {
		return nil, domainroom.ErrInvalidRoomID
	}
	if input.TargetUserID <= 0 {
		return nil, domainroom.ErrInvalidRoomMemberID
	}
	return uc.rooms.KickMember(ctx, roomports.KickMemberInput{
		RoomID:       roomID,
		HostUserID:   input.HostUserID,
		TargetUserID: input.TargetUserID,
	})
}
