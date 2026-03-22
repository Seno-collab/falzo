package application

import (
	"context"
	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/domain"
)

func (s *service) Logout(ctx context.Context, cmd command.Logout) error {
	if s.sessions == nil || s.tokenAuth == nil {
		return domain.ErrAuthDependencyUnavailable
	}

	if cmd.Token == "" {
		return domain.ErrInvalidToken
	}

	principal, err := s.tokenAuth.Authenticate(cmd.Token)
	if err != nil {
		return err
	}

	return s.sessions.RevokeBySessionID(ctx, principal.SessionID)
}
