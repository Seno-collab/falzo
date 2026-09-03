package authapp

import (
	authports "be/internal/application/ports/auth"
	domainuser "be/internal/domain/user"
	"be/internal/shared/clock"
	"context"
)

type RefreshTokenUseCase struct {
	users    authports.UserRepository
	tokens   authports.TokenManager
	sessions authports.TokenSessionStore
	clock    clock.Clock
}

type RefreshTokenInput struct{ RefreshToken string }

func NewRefreshTokenUseCase(users authports.UserRepository, tokens authports.TokenManager, sessions authports.TokenSessionStore, c clock.Clock) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{users: users, tokens: tokens, sessions: sessions, clock: c}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshTokenInput) (*LoginOutput, error) {
	now := uc.clock.Now()
	claims, err := uc.tokens.ParseRefresh(input.RefreshToken, now)
	if err != nil {
		return nil, domainuser.ErrInvalidToken
	}
	consumed, err := uc.sessions.ConsumeRefresh(ctx, claims.TokenID)
	if err != nil {
		return nil, err
	}
	if !consumed {
		return nil, domainuser.ErrInvalidToken
	}
	u, err := uc.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, domainuser.ErrInvalidToken
	}
	if err := u.CanLogin(now); err != nil {
		return nil, err
	}
	pair, err := uc.tokens.GeneratePair(u.ID, u.UserName, now)
	if err != nil {
		return nil, err
	}
	if err := uc.sessions.SaveRefresh(ctx, pair.RefreshTokenID, u.ID, pair.RefreshExpiresAt.Sub(now)); err != nil {
		return nil, err
	}
	return &LoginOutput{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken, TokenType: "Bearer", ExpiresIn: int64(pair.AccessExpiresAt.Sub(now).Seconds()), UserName: u.UserName}, nil
}
