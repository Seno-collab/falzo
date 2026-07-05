package authapp

import (
	"be/internal/application/ports"
	domainuser "be/internal/domain/user"
	"be/internal/shared/clock"
	"context"
	"strings"
)

type ForgotPasswordUseCase struct {
	users    ports.UserRepository
	tokens   ports.TokenManager
	sessions ports.TokenSessionStore
	clock    clock.Clock
}

type ForgotPasswordInput struct{ UserName string }
type ForgotPasswordOutput struct {
	ResetToken string `json:"reset_token,omitempty"`
}

func NewForgotPasswordUseCase(users ports.UserRepository, tokens ports.TokenManager, sessions ports.TokenSessionStore, c clock.Clock) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{users: users, tokens: tokens, sessions: sessions, clock: c}
}

func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, input ForgotPasswordInput) (*ForgotPasswordOutput, error) {
	u, err := uc.users.FindByUserName(ctx, strings.ToLower(strings.TrimSpace(input.UserName)))
	if err != nil {
		// Do not reveal whether an account exists.
		return &ForgotPasswordOutput{}, nil
	}
	now := uc.clock.Now()
	token, claims, err := uc.tokens.GeneratePasswordReset(u.ID, u.UserName, now)
	if err != nil {
		return nil, err
	}
	if err := uc.sessions.SavePasswordReset(ctx, claims.TokenID, u.ID, claims.ExpiresAt.Sub(now)); err != nil {
		return nil, err
	}
	return &ForgotPasswordOutput{ResetToken: token}, nil
}

type ResetPasswordUseCase struct {
	users    ports.UserRepository
	hasher   ports.PasswordHasher
	tokens   ports.TokenManager
	sessions ports.TokenSessionStore
	clock    clock.Clock
}

type ResetPasswordInput struct{ Token, NewPassword string }

func NewResetPasswordUseCase(users ports.UserRepository, hasher ports.PasswordHasher, tokens ports.TokenManager, sessions ports.TokenSessionStore, c clock.Clock) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{users: users, hasher: hasher, tokens: tokens, sessions: sessions, clock: c}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, input ResetPasswordInput) error {
	claims, err := uc.tokens.ParsePasswordReset(input.Token, uc.clock.Now())
	if err != nil {
		return domainuser.ErrInvalidToken
	}
	consumed, err := uc.sessions.ConsumePasswordReset(ctx, claims.TokenID)
	if err != nil {
		return err
	}
	if !consumed {
		return domainuser.ErrInvalidToken
	}
	hash, err := uc.hasher.Hash(input.NewPassword)
	if err != nil {
		return err
	}
	return uc.users.UpdatePassword(ctx, claims.UserID, hash)
}
