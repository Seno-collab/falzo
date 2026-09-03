package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"context"
	"strings"
)

type UpdateDiscussionInput struct {
	RoomID            string
	HostUserID        int64
	DiscussionSeconds int
}

type UpdateDiscussionUseCase struct {
	rooms roomports.Repository
}

func NewUpdateDiscussionUseCase(rooms roomports.Repository) *UpdateDiscussionUseCase {
	return &UpdateDiscussionUseCase{rooms: rooms}
}

func (uc *UpdateDiscussionUseCase) Execute(ctx context.Context, input UpdateDiscussionInput) (*domainroom.Room, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if !domainroom.IsValidID(roomID) {
		return nil, domainroom.ErrInvalidRoomID
	}
	if input.DiscussionSeconds < domainroom.MinDiscussionSeconds || input.DiscussionSeconds > domainroom.MaxDiscussionSeconds {
		return nil, domainroom.ErrInvalidDiscussionTime
	}
	return uc.rooms.UpdateDiscussionSeconds(ctx, roomports.UpdateDiscussionInput{
		RoomID:            roomID,
		HostUserID:        input.HostUserID,
		DiscussionSeconds: input.DiscussionSeconds,
	})
}
