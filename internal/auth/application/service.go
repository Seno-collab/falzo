package application

import (
	"context"

	"falzo/internal/auth/application/command"
	"falzo/internal/auth/application/query"
	"falzo/internal/auth/domain/repository"
	domainservice "falzo/internal/auth/domain/service"
)

type TokenIssuer interface {
	Issue(principal query.AuthenticatedUser) (string, error)
}

type TokenAuthenticator interface {
	Authenticate(rawToken string) (*query.AuthenticatedUser, error)
}

type Service interface {
	Register(ctx context.Context, cmd command.Register) error
	Login(ctx context.Context, cmd command.Login) (string, error)
	Logout(ctx context.Context, cmd command.Logout) error
	Authenticate(ctx context.Context, rawToken string) (*query.AuthenticatedUser, error)
}

type service struct {
	accounts    repository.AccountRepository
	passwords   domainservice.PasswordHasher
	tokenIssuer TokenIssuer
	tokenAuth   TokenAuthenticator
}

func New(
	accounts repository.AccountRepository,
	passwords domainservice.PasswordHasher,
	tokenIssuer TokenIssuer,
	tokenAuth TokenAuthenticator,
) Service {
	return &service{
		accounts:    accounts,
		passwords:   passwords,
		tokenIssuer: tokenIssuer,
		tokenAuth:   tokenAuth,
	}
}
