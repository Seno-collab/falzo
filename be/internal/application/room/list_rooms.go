package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"context"
)

type ListRoomsUseCase struct {
	rooms roomports.Repository
}

func NewListRoomsUseCase(rooms roomports.Repository) *ListRoomsUseCase {
	return &ListRoomsUseCase{rooms: rooms}
}

func (uc *ListRoomsUseCase) Execute(ctx context.Context) ([]*domainroom.Room, error) {
	return uc.rooms.ListActive(ctx)
}
