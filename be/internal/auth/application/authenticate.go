package application

import (
	"context"
	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain"
)

func (s *service) Authenticate(ctx context.Context, rawToken string) (*query.AuthenticatedUser, error) {
	if s.tokenAuth == nil || s.sessions == nil {
		return nil, domain.ErrInvalidToken
	}

	principal, err := s.tokenAuth.Authenticate(rawToken)
	if err != nil {
		return nil, err
	}

	active, err := s.sessions.IsSessionActive(ctx, principal.SessionID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, domain.ErrSessionRevoked
	}

	return principal, nil
}
