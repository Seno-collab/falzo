package application

import (
	"context"
	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain"
)

func (s *service) Authenticate(ctx context.Context, rawToken string) (*query.AuthenticatedUser, error) {
	if s.tokenAuth == nil {
		return nil, domain.ErrInvalidToken
	}

	return s.tokenAuth.Authenticate(rawToken)
}
