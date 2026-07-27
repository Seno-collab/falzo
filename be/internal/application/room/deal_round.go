package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"be/internal/shared/clock"
	"context"
	"crypto/rand"
	"math/big"
	"strings"
)

type DealRoundInput struct {
	RoomID     string
	HostUserID int64
}

type DealRoundUseCase struct {
	rooms roomports.Repository
	clock clock.Clock
}

func NewDealRoundUseCase(rooms roomports.Repository, clock clock.Clock) *DealRoundUseCase {
	return &DealRoundUseCase{rooms: rooms, clock: clock}
}

func (uc *DealRoundUseCase) Execute(ctx context.Context, input DealRoundInput) (*domainroom.RoundCard, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if !domainroom.IsValidID(roomID) {
		return nil, domainroom.ErrInvalidRoomID
	}

	room, err := uc.rooms.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if input.HostUserID != room.HostUserID {
		return nil, domainroom.ErrNotRoomHost
	}
	if len(room.Members) < 2 {
		return nil, domainroom.ErrNotEnoughPlayers
	}

	pair, err := uc.rooms.FindRandomWordPair(ctx, room.LanguageCode)
	if err != nil {
		return nil, err
	}
	undercoverIndex, err := secureRandomIndex(len(room.Members))
	if err != nil {
		return nil, err
	}
	return uc.rooms.StartRound(ctx, roomports.StartRoundInput{
		RoomID:             room.ID,
		HostUserID:         input.HostUserID,
		WordPairID:         pair.ID,
		CommonWord:         pair.CommonWord,
		DifferentWord:      pair.DifferentWord,
		UndercoverPlayerID: room.Members[undercoverIndex].UserID,
		DealtAt:            uc.clock.Now(),
	})
}

func secureRandomIndex(size int) (int, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(int64(size)))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()), nil
}
