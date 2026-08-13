package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	gameengine "be/internal/game"
	"be/internal/shared/clock"
	"context"
	"strings"
	"time"
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
	// A new game reuses the whole room roster, including spectators from the
	// previous finished game. StartRound resets their eliminated state atomically.
	activeMembers := append([]domainroom.Member(nil), room.Members...)
	if len(activeMembers) < domainroom.MinPlayersPerRoom {
		return nil, domainroom.ErrNotEnoughPlayers
	}

	pair, err := uc.rooms.FindRandomWordPair(ctx, room.LanguageCode)
	if err != nil {
		return nil, err
	}
	playerIDs := make([]int64, 0, len(activeMembers))
	for _, member := range activeMembers {
		playerIDs = append(playerIDs, member.UserID)
	}
	assignments, err := gameengine.AssignClassicRoles(playerIDs, room.MrWhiteEnabled, gameengine.SecureIndex)
	if err != nil {
		return nil, err
	}
	var undercoverPlayerID int64
	var mrWhitePlayerID *int64
	for _, assignment := range assignments {
		switch assignment.Role {
		case domainroom.CardRoleUndercover:
			undercoverPlayerID = assignment.PlayerID
		case domainroom.CardRoleMrWhite:
			playerID := assignment.PlayerID
			mrWhitePlayerID = &playerID
		}
	}
	turnOrder, err := shufflePlayerIDs(playerIDs)
	if err != nil {
		return nil, err
	}
	now := uc.clock.Now()
	return uc.rooms.StartRound(ctx, roomports.StartRoundInput{
		RoomID:             room.ID,
		HostUserID:         input.HostUserID,
		WordPairID:         pair.ID,
		CommonWord:         pair.CommonWord,
		DifferentWord:      pair.DifferentWord,
		UndercoverPlayerID: undercoverPlayerID,
		MrWhitePlayerID:    mrWhitePlayerID,
		TurnOrder:          turnOrder,
		DealtAt:            now,
		RoleRevealEndsAt:   now.Add(time.Duration(domainroom.RoleRevealDurationSeconds) * time.Second),
	})
}

func shufflePlayerIDs(playerIDs []int64) ([]int64, error) {
	shuffled := append([]int64(nil), playerIDs...)
	for index := len(shuffled) - 1; index > 0; index-- {
		other, err := gameengine.SecureIndex(index + 1)
		if err != nil {
			return nil, err
		}
		shuffled[index], shuffled[other] = shuffled[other], shuffled[index]
	}
	return shuffled, nil
}
