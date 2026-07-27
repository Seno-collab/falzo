package domainroom

import "errors"

var (
	ErrRoomNotFound          = errors.New("room not found")
	ErrRoomNameRequired      = errors.New("room name is required")
	ErrRoomNameTooLong       = errors.New("room name must be at most 80 characters")
	ErrInvalidRoomID         = errors.New("invalid room id")
	ErrInvalidHostUserID     = errors.New("invalid host user id")
	ErrInvalidInviteCode     = errors.New("invalid invite code")
	ErrInvalidMaxPlayers     = errors.New("invalid maximum players")
	ErrInvalidLanguageCode   = errors.New("room language must be en or vi")
	ErrInvalidRoomExpiration = errors.New("invalid room expiration")
	ErrInviteCodeConflict    = errors.New("invite code already exists")
	ErrInviteCodeUnavailable = errors.New("could not allocate an invite code")
	ErrRoomFull              = errors.New("room is full")
	ErrRoomNotWaiting        = errors.New("room is not waiting for players")
	ErrNotRoomHost           = errors.New("only the room host can deal a round")
	ErrNotEnoughPlayers      = errors.New("at least two players are required to deal a round")
	ErrRoundCardNotFound     = errors.New("round card not found")
	ErrWordPairNotFound      = errors.New("no active word pair found")
)
