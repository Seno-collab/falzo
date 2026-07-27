package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"context"
)

type JoinRoomInput struct {
	UserID     int64
	InviteCode string
}

type JoinRoomUseCase struct {
	rooms roomports.Repository
}

func NewJoinRoomUseCase(rooms roomports.Repository) *JoinRoomUseCase {
	return &JoinRoomUseCase{rooms: rooms}
}

func (uc *JoinRoomUseCase) Execute(ctx context.Context, input JoinRoomInput) (*domainroom.Room, error) {
	inviteCode := domainroom.NormalizeInviteCode(input.InviteCode)
	if input.UserID <= 0 || !domainroom.IsValidInviteCode(inviteCode) {
		return nil, domainroom.ErrInvalidInviteCode
	}
	return uc.rooms.JoinByInviteCode(ctx, inviteCode, input.UserID)
}
