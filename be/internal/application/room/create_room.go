package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"be/internal/shared/clock"
	"context"
	"errors"
	"strings"
	"time"
)

type CreateRoomInput struct {
	HostUserID   int64
	Name         string
	LanguageCode domainroom.LanguageCode
	MaxPlayers   int
}

type CreateRoomUseCase struct {
	rooms             roomports.Repository
	inviteCodes       roomports.InviteCodeGenerator
	clock             clock.Clock
	maxAllowedPlayers int
	roomTimeout       time.Duration
}

func NewCreateRoomUseCase(
	rooms roomports.Repository,
	inviteCodes roomports.InviteCodeGenerator,
	c clock.Clock,
	maxAllowedPlayers int,
	roomTimeout time.Duration,
) *CreateRoomUseCase {
	return &CreateRoomUseCase{
		rooms:             rooms,
		inviteCodes:       inviteCodes,
		clock:             c,
		maxAllowedPlayers: maxAllowedPlayers,
		roomTimeout:       roomTimeout,
	}
}

func (uc *CreateRoomUseCase) Execute(ctx context.Context, input CreateRoomInput) (*domainroom.Room, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.MaxPlayers < domainroom.MinPlayersPerRoom || input.MaxPlayers > uc.maxAllowedPlayers {
		return nil, domainroom.ErrInvalidMaxPlayers
	}

	now := uc.clock.Now()
	for range 5 {
		inviteCode, err := uc.inviteCodes.Generate()
		if err != nil {
			return nil, err
		}

		room, err := domainroom.New(
			input.Name,
			inviteCode,
			input.LanguageCode,
			input.HostUserID,
			input.MaxPlayers,
			now,
			now.Add(uc.roomTimeout),
		)
		if err != nil {
			return nil, err
		}

		created, err := uc.rooms.CreateWithHost(ctx, room)
		if errors.Is(err, domainroom.ErrInviteCodeConflict) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return created, nil
	}

	return nil, domainroom.ErrInviteCodeUnavailable
}
