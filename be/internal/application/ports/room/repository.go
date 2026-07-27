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
}

type StartRoundInput struct {
	RoomID             string
	HostUserID         int64
	WordPairID         int64
	CommonWord         string
	DifferentWord      string
	UndercoverPlayerID int64
	DealtAt            time.Time
}

type InviteCodeGenerator interface {
	Generate() (string, error)
}
