package application_test

import (
	"context"
	"testing"
	"time"

	"falzo/internal/auth/application"
	"falzo/internal/auth/application/command"
	"falzo/internal/auth/application/query"
	"falzo/internal/auth/domain"
	"falzo/internal/auth/domain/aggregate"
	"falzo/internal/auth/domain/entity"
	"falzo/internal/auth/domain/valueobject"
	"falzo/internal/auth/infrastructure/security/bcrypt"
	"falzo/internal/auth/infrastructure/token"
	"falzo/pkg/config"
)

type fakeAccountRepository struct {
	account *aggregate.Account
	saveErr error
	findErr error
}

func (f *fakeAccountRepository) Save(ctx context.Context, account *aggregate.Account) error {
	f.account = account
	return f.saveErr
}

func (f *fakeAccountRepository) FindActiveByUsername(ctx context.Context, username valueobject.Username) (*aggregate.Account, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}

	return f.account, nil
}

func TestAuthenticateToken(t *testing.T) {
	jwtManager := token.NewJWTManager(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	})

	service := application.New(nil, bcrypt.NewPasswordHasher(), jwtManager, jwtManager)

	rawToken, err := jwtManager.Issue(query.AuthenticatedUser{
		UserID:   42,
		Username: "admin",
	})
	if err != nil {
		t.Fatalf("unexpected sign error: %v", err)
	}

	principal, err := service.Authenticate(t.Context(), rawToken)
	if err != nil {
		t.Fatalf("unexpected auth error: %v", err)
	}

	if principal.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", principal.UserID)
	}

	if principal.Username != "admin" {
		t.Fatalf("expected username admin, got %q", principal.Username)
	}
}

func TestLoginUnavailableWithoutRepository(t *testing.T) {
	jwtManager := token.NewJWTManager(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	})

	service := application.New(nil, bcrypt.NewPasswordHasher(), jwtManager, jwtManager)

	if _, err := service.Login(t.Context(), command.Login{Username: "admin", Password: "admin123"}); err != domain.ErrAuthUnavailable {
		t.Fatalf("expected auth unavailable, got %v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	hasher := bcrypt.NewPasswordHasher()
	hash, err := hasher.Hash(valueobject.RawPassword("admin123"))
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	username, _ := valueobject.NewUsername("admin")
	email, _ := valueobject.NewEmail("admin@example.com")
	account := aggregate.RehydrateAccount(entity.User{
		ID:           7,
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		IsActive:     true,
	}, []string{"user"})

	repo := &fakeAccountRepository{account: account}
	jwtManager := token.NewJWTManager(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	})

	service := application.New(repo, hasher, jwtManager, jwtManager)

	rawToken, err := service.Login(t.Context(), command.Login{
		Username: "admin",
		Password: "admin123",
	})
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}

	principal, err := service.Authenticate(t.Context(), rawToken)
	if err != nil {
		t.Fatalf("unexpected auth error: %v", err)
	}

	if principal.UserID != 7 {
		t.Fatalf("expected user id 7, got %d", principal.UserID)
	}
}
