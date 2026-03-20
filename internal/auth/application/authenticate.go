package application

import (
	"context"

	"falzo/internal/auth/application/query"
	"falzo/internal/auth/domain"
)

func (s *service) Authenticate(ctx context.Context, rawToken string) (*query.AuthenticatedUser, error) {
	if s.tokenAuth == nil {
		return nil, domain.ErrInvalidToken
	}

	return s.tokenAuth.Authenticate(rawToken)
}
