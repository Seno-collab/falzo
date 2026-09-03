package authapp

import (
	authports "be/internal/application/ports/auth"
	"be/internal/shared/clock"
	"context"
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidRegistrationInput = errors.New("invalid registration input")

var passwordUsernamePattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

type RegisterUseCase struct {
	passwordHasher authports.PasswordHasher
	users          authports.UserRepository
	tokens         authports.TokenManager
	sessions       authports.TokenSessionStore
	clock          clock.Clock
}

type RegisterInput struct {
	UserName string
	Password string
}

func NewRegisterUseCase(
	hasher authports.PasswordHasher,
	users authports.UserRepository,
	tokens authports.TokenManager,
	sessions authports.TokenSessionStore,
	c clock.Clock,
) *RegisterUseCase {
	return &RegisterUseCase{
		passwordHasher: hasher,
		users:          users,
		tokens:         tokens,
		sessions:       sessions,
		clock:          c,
	}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*LoginOutput, error) {
	username := strings.ToLower(strings.TrimSpace(input.UserName))
	if !validPasswordUsername(username) || len(input.Password) < 8 || len([]byte(input.Password)) > 72 {
		return nil, ErrInvalidRegistrationInput
	}

	passwordHash, err := uc.passwordHasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	now := uc.clock.Now()
	user, err := uc.users.CreatePasswordUser(ctx, username, passwordHash, now)
	if err != nil {
		return nil, err
	}

	pair, err := uc.tokens.GeneratePair(user.ID, user.UserName, now)
	if err != nil {
		return nil, err
	}
	if err := uc.sessions.SaveRefresh(ctx, pair.RefreshTokenID, user.ID, pair.RefreshExpiresAt.Sub(now)); err != nil {
		return nil, err
	}

	return &LoginOutput{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(pair.AccessExpiresAt.Sub(now).Seconds()),
		UserName:     user.UserName,
	}, nil
}

func validPasswordUsername(username string) bool {
	return len(username) >= 3 && len(username) <= 30 && passwordUsernamePattern.MatchString(username)
}
