package authapp

import (
	"be/internal/application/ports"
	domainuser "be/internal/domain/user"
	"be/internal/shared/clock"
	"context"
)

type LogoutUseCase struct {
	tokens   ports.TokenManager
	sessions ports.TokenSessionStore
	clock    clock.Clock
}

func NewLogoutUseCase(tokens ports.TokenManager, sessions ports.TokenSessionStore, c clock.Clock) *LogoutUseCase {
	return &LogoutUseCase{tokens: tokens, sessions: sessions, clock: c}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, refreshToken string) error {
	claims, err := uc.tokens.ParseRefresh(refreshToken, uc.clock.Now())
	if err != nil {
		return domainuser.ErrInvalidToken
	}
	consumed, err := uc.sessions.ConsumeRefresh(ctx, claims.TokenID)
	if err != nil {
		return err
	}
	if !consumed {
		return domainuser.ErrInvalidToken
	}
	return nil
}
