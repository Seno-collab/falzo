package game

import (
	domainroom "be/internal/domain/room"
	"crypto/rand"
	"errors"
	"math/big"
)

var ErrNotEnoughPlayers = errors.New("at least four players are required")

type IndexPicker func(size int) (int, error)

type RoleAssignment struct {
	PlayerID int64
	Role     domainroom.CardRole
}

func AssignClassicRoles(playerIDs []int64, mrWhiteEnabled bool, pick IndexPicker) ([]RoleAssignment, error) {
	if len(playerIDs) < domainroom.MinPlayersPerRoom {
		return nil, ErrNotEnoughPlayers
	}
	if pick == nil {
		pick = SecureIndex
	}
	undercoverIndex, err := pick(len(playerIDs))
	if err != nil {
		return nil, err
	}
	assignments := make([]RoleAssignment, len(playerIDs))
	for index, playerID := range playerIDs {
		assignments[index] = RoleAssignment{PlayerID: playerID, Role: domainroom.CardRoleCivilian}
	}
	assignments[undercoverIndex].Role = domainroom.CardRoleUndercover

	if mrWhiteEnabled {
		whiteIndex, err := pick(len(playerIDs) - 1)
		if err != nil {
			return nil, err
		}
		if whiteIndex >= undercoverIndex {
			whiteIndex++
		}
		assignments[whiteIndex].Role = domainroom.CardRoleMrWhite
	}
	return assignments, nil
}

func SecureIndex(size int) (int, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(int64(size)))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()), nil
}
