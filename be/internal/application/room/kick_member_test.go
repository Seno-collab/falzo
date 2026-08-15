package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"context"
	"testing"
)

type kickMemberTestRepository struct {
	roomports.Repository
	input roomports.KickMemberInput
	room  *domainroom.Room
}

func (r *kickMemberTestRepository) KickMember(_ context.Context, input roomports.KickMemberInput) (*domainroom.Room, error) {
	r.input = input
	return r.room, nil
}

func TestKickMemberPassesValidatedInputToRepository(t *testing.T) {
	repository := &kickMemberTestRepository{room: &domainroom.Room{ID: "56cae50e-5fd5-47f3-9942-a6ae7b1a48dc"}}
	useCase := NewKickMemberUseCase(repository)

	room, err := useCase.Execute(context.Background(), KickMemberInput{
		RoomID:       " 56cae50e-5fd5-47f3-9942-a6ae7b1a48dc ",
		HostUserID:   11,
		TargetUserID: 22,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if room != repository.room {
		t.Fatalf("Execute() room = %#v, want %#v", room, repository.room)
	}
	if repository.input.RoomID != repository.room.ID || repository.input.HostUserID != 11 || repository.input.TargetUserID != 22 {
		t.Fatalf("repository input = %#v", repository.input)
	}
}

func TestKickMemberRejectsInvalidTarget(t *testing.T) {
	useCase := NewKickMemberUseCase(&kickMemberTestRepository{})
	_, err := useCase.Execute(context.Background(), KickMemberInput{
		RoomID:       "56cae50e-5fd5-47f3-9942-a6ae7b1a48dc",
		HostUserID:   11,
		TargetUserID: 0,
	})
	if err != domainroom.ErrInvalidRoomMemberID {
		t.Fatalf("Execute() error = %v, want %v", err, domainroom.ErrInvalidRoomMemberID)
	}
}
