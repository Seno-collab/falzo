package ports

import (
	domainuser "be/internal/domain/user"
	"context"
	"time"
)

type UserRepository interface {
	FindByUserName(ctx context.Context, username string) (*domainuser.User, error)
	FindByID(ctx context.Context, id int64) (*domainuser.User, error)
	Create(ctx context.Context, user *domainuser.User) error
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	UpdateLoginState(ctx context.Context, user *domainuser.User) error
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	RefreshTokenID   string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type TokenClaims struct {
	UserID    int64
	UserName  string
	TokenID   string
	ExpiresAt time.Time
}

type TokenManager interface {
	GeneratePair(userID int64, username string, now time.Time) (TokenPair, error)
	ParseAccess(token string, now time.Time) (TokenClaims, error)
	ParseRefresh(token string, now time.Time) (TokenClaims, error)
	GeneratePasswordReset(userID int64, username string, now time.Time) (token string, claims TokenClaims, err error)
	ParsePasswordReset(token string, now time.Time) (TokenClaims, error)
}

type TokenSessionStore interface {
	SaveRefresh(ctx context.Context, tokenID string, userID int64, ttl time.Duration) error
	ConsumeRefresh(ctx context.Context, tokenID string) (bool, error)
	SavePasswordReset(ctx context.Context, tokenID string, userID int64, ttl time.Duration) error
	ConsumePasswordReset(ctx context.Context, tokenID string) (bool, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password string, hashedPassword string) bool
}
