package application

import (
	"context"
	"time"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain/repository"
	domainservice "falzo-be/internal/auth/domain/service"
)

type TokenIssuer interface {
	Issue(principal query.AuthenticatedUser) (string, error)
}

type TokenAuthenticator interface {
	Authenticate(rawToken string) (*query.AuthenticatedUser, error)
}

type Service interface {
	Register(ctx context.Context, cmd command.Register) error
	Login(ctx context.Context, cmd command.Login) (query.TokenPair, error)
	Refresh(ctx context.Context, cmd command.Refresh) (query.TokenPair, error)
	Logout(ctx context.Context, cmd command.Logout) error
	Authenticate(ctx context.Context, rawToken string) (*query.AuthenticatedUser, error)
}

type service struct {
	accounts    repository.AccountRepository
	sessions    repository.SessionRepository
	passwords   domainservice.PasswordHasher
	tokenIssuer TokenIssuer
	tokenAuth   TokenAuthenticator
	refreshTTL  time.Duration
}

func New(
	accounts repository.AccountRepository,
	sessions repository.SessionRepository,
	passwords domainservice.PasswordHasher,
	tokenIssuer TokenIssuer,
	tokenAuth TokenAuthenticator,
	refreshTTL time.Duration,
) Service {
	return &service{
		accounts:    accounts,
		sessions:    sessions,
		passwords:   passwords,
		tokenIssuer: tokenIssuer,
		tokenAuth:   tokenAuth,
		refreshTTL:  refreshTTL,
	}
}
