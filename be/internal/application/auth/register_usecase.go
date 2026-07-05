package authapp

import (
	"be/internal/application/ports"
	domainuser "be/internal/domain/user"
	"be/internal/shared/clock"
	"context"
	"strings"
)

type RegisterUseCase struct {
	users  ports.UserRepository
	hasher ports.PasswordHasher
	clock  clock.Clock
}

type RegisterInput struct {
	UserName string
	Password string
}

type RegisterOutput struct {
	ID       int64  `json:"id"`
	UserName string `json:"username"`
}

func NewRegisterUseCase(users ports.UserRepository, hasher ports.PasswordHasher, c clock.Clock) *RegisterUseCase {
	return &RegisterUseCase{users: users, hasher: hasher, clock: c}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	hash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}
	now := uc.clock.Now()
	u := &domainuser.User{
		UserName:     strings.ToLower(strings.TrimSpace(input.UserName)),
		PasswordHash: hash,
		Status:       domainuser.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := uc.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return &RegisterOutput{ID: u.ID, UserName: u.UserName}, nil
}
