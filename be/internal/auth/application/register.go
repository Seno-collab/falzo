package application

import (
	"context"
	"falzo-be/internal/auth/domain"
	"falzo-be/internal/auth/domain/aggregate"
	"falzo-be/internal/auth/domain/valueobject"

	"falzo-be/internal/auth/application/command"
)

func (s *service) Register(ctx context.Context, cmd command.Register) error {
	if s.accounts == nil || s.passwords == nil {
		return domain.ErrAuthDependencyUnavailable
	}

	username, err := valueobject.NewUsername(cmd.Username)
	if err != nil {
		return err
	}

	email, err := valueobject.NewEmail(cmd.Email)
	if err != nil {
		return err
	}

	password, err := valueobject.NewRawPassword(cmd.Password)
	if err != nil {
		return err
	}

	hash, err := s.passwords.Hash(password)
	if err != nil {
		return err
	}

	account := aggregate.NewAccount(0, username, email, hash, []string{"user"})
	return s.accounts.Save(ctx, account)
}
