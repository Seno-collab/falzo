package authapp

import (
	"be/internal/application/ports"
	domainuser "be/internal/domain/user"
	"be/internal/shared/clock"
	"context"
	"strings"
	"time"
)

type LoginUseCase struct {
	passwordHasher ports.PasswordHasher
	users          ports.UserRepository
	tokens         ports.TokenManager
	sessions       ports.TokenSessionStore
	maxAttempts    int
	clock          clock.Clock
	lockDuration   time.Duration
}

type LoginInput struct{ UserName, Password string }

type LoginOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

func NewLoginUseCase(hasher ports.PasswordHasher, users ports.UserRepository, tokens ports.TokenManager, sessions ports.TokenSessionStore, maxAttempts int, lockDuration time.Duration, c clock.Clock) *LoginUseCase {
	return &LoginUseCase{passwordHasher: hasher, users: users, tokens: tokens, sessions: sessions, maxAttempts: maxAttempts, lockDuration: lockDuration, clock: c}
}

func (uc *LoginUseCase) Execute(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	now := uc.clock.Now()
	u, err := uc.users.FindByUserName(ctx, strings.ToLower(strings.TrimSpace(input.UserName)))
	if err != nil {
		return nil, domainuser.ErrInvalidUsernameOrPassword
	}
	if err := u.CanLogin(now); err != nil {
		return nil, err
	}
	if !uc.passwordHasher.Compare(input.Password, u.PasswordHash) {
		u.RecordFailedLogin(now, uc.maxAttempts, uc.lockDuration)
		if err := uc.users.UpdateLoginState(ctx, u); err != nil {
			return nil, err
		}
		return nil, domainuser.ErrInvalidUsernameOrPassword
	}
	u.RecordSuccessfulLogin(now)
	if err := uc.users.UpdateLoginState(ctx, u); err != nil {
		return nil, err
	}
	pair, err := uc.tokens.GeneratePair(u.ID, u.UserName, now)
	if err != nil {
		return nil, err
	}
	if err := uc.sessions.SaveRefresh(ctx, pair.RefreshTokenID, u.ID, pair.RefreshExpiresAt.Sub(now)); err != nil {
		return nil, err
	}
	return &LoginOutput{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken, TokenType: "Bearer", ExpiresIn: int64(pair.AccessExpiresAt.Sub(now).Seconds())}, nil
}
