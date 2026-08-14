package domainroom

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MinPlayersPerRoom           = 4
	MaxPlayersPerRoom           = 12
	MinDiscussionSeconds        = 10
	MaxDiscussionSeconds        = 30
	DefaultDiscussionSeconds    = 30
	RoleRevealDurationSeconds   = 15
	VotingDurationSeconds       = 30
	ResultRevealDurationSeconds = 5
	MrWhiteGuessDurationSeconds = 30
)

type Status string
type LanguageCode string
type EventRoom string

const (
	StatusWaiting Status = "waiting"
	StatusPlaying Status = "playing"
	StatusClosed  Status = "closed"
)

const (
	EventRoomCreated      EventRoom = "room.created"
	EventRoomUpdated      EventRoom = "room.updated"
	EventRoomDeleted      EventRoom = "room.deleted"
	EventRoomMemberJoined EventRoom = "room.member.joined"
	EventRoomMemberLeft   EventRoom = "room.member.left"
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
	Eliminated bool
}

type CardRole string
type RoundPhase string

const (
	CardRoleCivilian   CardRole = "civilian"
	CardRoleUndercover CardRole = "undercover"
	CardRoleMrWhite    CardRole = "mr_white"
)

const (
	RoundPhaseWaiting         RoundPhase = "WAITING"
	RoundPhaseRevealingRole   RoundPhase = "REVEALING_ROLE"
	RoundPhaseDescribing      RoundPhase = "DESCRIBING"
	RoundPhaseVoting          RoundPhase = "VOTING"
	RoundPhaseRevealingResult RoundPhase = "REVEALING_RESULT"
	RoundPhaseMrWhiteGuessing RoundPhase = "MR_WHITE_GUESSING"
	RoundPhaseGameFinished    RoundPhase = "GAME_FINISHED"
)

type WinningSide string

const (
	WinningSideCivilians  WinningSide = "civilians"
	WinningSideUndercover WinningSide = "undercover"
	WinningSideMrWhite    WinningSide = "mr_white"
)

type RoundCard struct {
	RoomID          string
	RoundNumber     int
	PlayerID        int64
	Role            CardRole
	Word            string
	DealtAt         time.Time
	Phase           RoundPhase
	PhaseDeadlineAt time.Time
}

type RoundState struct {
	RoomID              string
	RoundNumber         int
	CycleNumber         int
	Phase               RoundPhase
	PhaseDeadlineAt     *time.Time
	ReadyPlayers        int
	EligiblePlayers     int
	CurrentUserReady    bool
	CurrentTurnPlayerID *int64
	TurnNumber          int
	TotalTurns          int
	TurnEndsAt          *time.Time
	EligibleVoters      int
	VotesCast           int
	CurrentUserVoteID   *int64
	UndercoverPlayerID  *int64
	MrWhitePlayerID     *int64
	EliminatedPlayerID  *int64
	EliminatedRole      *CardRole
	Winner              *WinningSide
	MrWhiteGuessCorrect *bool
	FinalizedNow        bool
}

// PhaseTransition is the public, non-player-specific result of a scheduler
// transition. Private card and role data must never be added here.
type PhaseTransition struct {
	RoomID              string
	RoundNumber         int
	CycleNumber         int
	From                RoundPhase
	To                  RoundPhase
	CurrentTurnPlayerID *int64
	PhaseDeadlineAt     *time.Time
	PreviousDeadlineAt  time.Time
	TransitionedAt      time.Time
	MembersChanged      bool
}

type WordPair struct {
	ID            int64
	CommonWord    string
	DifferentWord string
	Category      string
	LanguageCode  LanguageCode
}

type Room struct {
	ID                string
	InviteCode        string
	Name              string
	LanguageCode      LanguageCode
	HostUserID        int64
	Status            Status
	MaxPlayers        int
	CurrentRound      int
	DiscussionSeconds int
	MrWhiteEnabled    bool
	Version           int64
	ExpiresAt         time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Members           []Member
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
		ID:                uuid.NewString(),
		InviteCode:        inviteCode,
		Name:              name,
		LanguageCode:      languageCode,
		HostUserID:        hostUserID,
		Status:            StatusWaiting,
		MaxPlayers:        maxPlayers,
		DiscussionSeconds: DefaultDiscussionSeconds,
		MrWhiteEnabled:    true,
		Version:           1,
		ExpiresAt:         expiresAt.UTC(),
		CreatedAt:         now.UTC(),
		UpdatedAt:         now.UTC(),
	}, nil
}

func CanTransition(from, to RoundPhase) bool {
	switch from {
	case RoundPhaseWaiting:
		return to == RoundPhaseRevealingRole
	case RoundPhaseRevealingRole:
		return to == RoundPhaseDescribing
	case RoundPhaseDescribing:
		return to == RoundPhaseVoting
	case RoundPhaseVoting:
		return to == RoundPhaseRevealingResult
	case RoundPhaseRevealingResult:
		return to == RoundPhaseMrWhiteGuessing || to == RoundPhaseDescribing || to == RoundPhaseGameFinished
	case RoundPhaseMrWhiteGuessing:
		return to == RoundPhaseDescribing || to == RoundPhaseGameFinished
	case RoundPhaseGameFinished:
		return to == RoundPhaseRevealingRole
	default:
		return false
	}
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
