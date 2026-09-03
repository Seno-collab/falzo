package authapp

import (
	authports "be/internal/application/ports/auth"
	domainuser "be/internal/domain/user"
	"be/internal/shared/clock"
	"context"
	"fmt"
	"strings"
	"time"
)

type AccountLockedError struct {
	UserID         int64
	UserName       string
	FailedAttempts int
	LockedUntil    time.Time
}

func (e *AccountLockedError) Error() string {
	return fmt.Sprintf("account %q is locked until %s", e.UserName, e.LockedUntil.UTC().Format(time.RFC3339))
}

func (e *AccountLockedError) Unwrap() error {
	return domainuser.ErrUserLocked
}

type LoginUseCase struct {
	passwordHasher authports.PasswordHasher
	users          authports.UserRepository
	tokens         authports.TokenManager
	sessions       authports.TokenSessionStore
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
	UserName     string `json:"username"`
}

func NewLoginUseCase(hasher authports.PasswordHasher, users authports.UserRepository, tokens authports.TokenManager, sessions authports.TokenSessionStore, maxAttempts int, lockDuration time.Duration, c clock.Clock) *LoginUseCase {
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
		if u.Status == domainuser.UserStatusLocked && u.LockUntil != nil {
			return nil, &AccountLockedError{
				UserID:         u.ID,
				UserName:       u.UserName,
				FailedAttempts: u.FailedAttempts,
				LockedUntil:    *u.LockUntil,
			}
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
	return &LoginOutput{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken, TokenType: "Bearer", ExpiresIn: int64(pair.AccessExpiresAt.Sub(now).Seconds()), UserName: u.UserName}, nil
}
