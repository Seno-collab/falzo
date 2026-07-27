package domainroom

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MinPlayersPerRoom = 4
	MaxPlayersPerRoom = 12
)

type Status string
type LanguageCode string

const (
	StatusWaiting Status = "waiting"
	StatusPlaying Status = "playing"
	StatusClosed  Status = "closed"
)

const (
	LanguageEnglish    LanguageCode = "en"
	LanguageVietnamese LanguageCode = "vi"
)

type Member struct {
	UserID     int64
	UserName   string
	SeatNumber int
	IsHost     bool
	JoinedAt   time.Time
}

type CardRole string

const (
	CardRoleCivilian   CardRole = "civilian"
	CardRoleUndercover CardRole = "undercover"
)

type RoundCard struct {
	RoomID      string
	RoundNumber int
	PlayerID    int64
	Role        CardRole
	Word        string
	DealtAt     time.Time
}

type WordPair struct {
	ID            int64
	CommonWord    string
	DifferentWord string
	Category      string
	LanguageCode  LanguageCode
}

type Room struct {
	ID           string
	InviteCode   string
	Name         string
	LanguageCode LanguageCode
	HostUserID   int64
	Status       Status
	MaxPlayers   int
	CurrentRound int
	Version      int64
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Members      []Member
}

func New(
	name string,
	inviteCode string,
	languageCode LanguageCode,
	hostUserID int64,
	maxPlayers int,
	now,
	expiresAt time.Time,
) (*Room, error) {
	name = strings.TrimSpace(name)
	inviteCode = NormalizeInviteCode(inviteCode)
	languageCode = NormalizeLanguageCode(languageCode)

	switch {
	case name == "":
		return nil, ErrRoomNameRequired
	case len(name) > 80:
		return nil, ErrRoomNameTooLong
	case !IsValidInviteCode(inviteCode):
		return nil, ErrInvalidInviteCode
	case !IsValidLanguageCode(languageCode):
		return nil, ErrInvalidLanguageCode
	case hostUserID <= 0:
		return nil, ErrInvalidHostUserID
	case maxPlayers < MinPlayersPerRoom || maxPlayers > MaxPlayersPerRoom:
		return nil, ErrInvalidMaxPlayers
	case !expiresAt.After(now):
		return nil, ErrInvalidRoomExpiration
	}

	return &Room{
		ID:           uuid.NewString(),
		InviteCode:   inviteCode,
		Name:         name,
		LanguageCode: languageCode,
		HostUserID:   hostUserID,
		Status:       StatusWaiting,
		MaxPlayers:   maxPlayers,
		Version:      1,
		ExpiresAt:    expiresAt.UTC(),
		CreatedAt:    now.UTC(),
		UpdatedAt:    now.UTC(),
	}, nil
}

func NormalizeLanguageCode(code LanguageCode) LanguageCode {
	return LanguageCode(strings.ToLower(strings.TrimSpace(string(code))))
}

func IsValidLanguageCode(code LanguageCode) bool {
	return code == LanguageEnglish || code == LanguageVietnamese
}

func NormalizeInviteCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func IsValidInviteCode(code string) bool {
	if len(code) < 6 || len(code) > 8 {
		return false
	}
	for _, char := range code {
		if (char < 'A' || char > 'Z') && (char < '0' || char > '9') {
			return false
		}
	}
	return true
}

func IsValidID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
