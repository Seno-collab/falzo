package roomports

import (
	domainroom "be/internal/domain/room"
	"context"
	"time"
)

type Repository interface {
	CreateWithHost(ctx context.Context, room *domainroom.Room) (*domainroom.Room, error)
	ListActive(ctx context.Context) ([]*domainroom.Room, error)
	FindByID(ctx context.Context, roomID string) (*domainroom.Room, error)
	JoinByInviteCode(ctx context.Context, inviteCode string, userID int64) (*domainroom.Room, error)
	FindRandomWordPair(ctx context.Context, languageCode domainroom.LanguageCode) (*domainroom.WordPair, error)
	StartRound(ctx context.Context, input StartRoundInput) (*domainroom.RoundCard, error)
	FindCurrentCard(ctx context.Context, roomID string, userID int64) (*domainroom.RoundCard, error)
	UpdateDiscussionSeconds(ctx context.Context, input UpdateDiscussionInput) (*domainroom.Room, error)
	FindCurrentRoundState(ctx context.Context, roomID string, userID int64, now time.Time) (*domainroom.RoundState, error)
	CastVote(ctx context.Context, input CastVoteInput) (*domainroom.RoundState, error)
	MarkPlayerReady(ctx context.Context, input PlayerActionInput) (*domainroom.RoundState, error)
	FinishTurn(ctx context.Context, input PlayerActionInput) (*domainroom.RoundState, error)
	SubmitMrWhiteGuess(ctx context.Context, input MrWhiteGuessInput) (*domainroom.RoundState, error)
}

type StartRoundInput struct {
	RoomID             string
	HostUserID         int64
	WordPairID         int64
	CommonWord         string
	DifferentWord      string
	UndercoverPlayerID int64
	MrWhitePlayerID    *int64
	TurnOrder          []int64
	DealtAt            time.Time
	RoleRevealEndsAt   time.Time
}

type PlayerActionInput struct {
	RoomID string
	UserID int64
	At     time.Time
}

type MrWhiteGuessInput struct {
	RoomID string
	UserID int64
	Guess  string
	At     time.Time
}

type UpdateDiscussionInput struct {
	RoomID            string
	HostUserID        int64
	DiscussionSeconds int
}

type CastVoteInput struct {
	RoomID       string
	VoterUserID  int64
	TargetUserID int64
	VotedAt      time.Time
}

type InviteCodeGenerator interface {
	Generate() (string, error)
}
