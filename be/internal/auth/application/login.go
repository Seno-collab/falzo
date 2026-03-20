package application

import (
	"context"

	"falzo/internal/auth/application/command"
	"falzo/internal/auth/application/query"
	"falzo/internal/auth/domain"
	"falzo/internal/auth/domain/aggregate"
	"falzo/internal/auth/domain/valueobject"
)

func (s *service) Login(ctx context.Context, cmd command.Login) (string, error) {
	if s.accounts == nil || s.passwords == nil || s.tokenIssuer == nil {
		return "", domain.ErrAuthUnavailable
	}

	username, err := valueobject.NewUsername(cmd.Username)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	password, err := valueobject.NewRawPassword(cmd.Password)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	account, err := s.accounts.FindActiveByUsername(ctx, username)
	if err != nil {
		return "", err
	}

	if err := s.passwords.Compare(account.User.PasswordHash, password); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	return s.tokenIssuer.Issue(principalFromAccount(account))
}

func principalFromAccount(account *aggregate.Account) query.AuthenticatedUser {
	return query.AuthenticatedUser{
		UserID:   account.User.ID,
		Username: account.User.Username.String(),
	}
}
