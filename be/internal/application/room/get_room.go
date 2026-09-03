package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"context"
	"strings"
)

type GetRoomUseCase struct {
	rooms roomports.Repository
}

func NewGetRoomUseCase(rooms roomports.Repository) *GetRoomUseCase {
	return &GetRoomUseCase{rooms: rooms}
}

func (uc *GetRoomUseCase) Execute(ctx context.Context, roomID string) (*domainroom.Room, error) {
	roomID = strings.TrimSpace(roomID)
	if !domainroom.IsValidID(roomID) {
		return nil, domainroom.ErrInvalidRoomID
	}
	return uc.rooms.FindByID(ctx, roomID)
}
