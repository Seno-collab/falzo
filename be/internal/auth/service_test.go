package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"falzo-be/internal/auth"
	authInfra "falzo-be/internal/auth/infra"
	"falzo-be/pkg/config"
)

type fakeAccountRepository struct {
	account *auth.Account
	saveErr error
	findErr error
}

func (f *fakeAccountRepository) Save(ctx context.Context, account *auth.Account) error {
	f.account = account
	return f.saveErr
}

func (f *fakeAccountRepository) FindActiveByEmail(ctx context.Context, email auth.Email) (*auth.Account, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.account, nil
}

type fakeSessionRepository struct {
	createdSessions []auth.Session
	active          map[string]bool
	createErr       error
	activeErr       error
}

func (f *fakeSessionRepository) Create(ctx context.Context, session auth.Session) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createdSessions = append(f.createdSessions, session)
	if f.active == nil {
		f.active = map[string]bool{}
	}
	f.active[session.SessionID] = true
	return nil
}

func (f *fakeSessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	if f.activeErr != nil {
		return false, f.activeErr
	}
	return f.active[sessionID], nil
}

func (f *fakeSessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*auth.Session, error) {
	return nil, auth.ErrInvalidToken
}

func (f *fakeSessionRepository) RotateRefreshToken(ctx context.Context, session auth.Session, newRefreshTokenHash string, expiresAtUnix int64) error {
	return auth.ErrInvalidToken
}

func (f *fakeSessionRepository) RevokeBySessionID(ctx context.Context, sessionID string) error {
	if f.active != nil {
		f.active[sessionID] = false
	}
	return nil
}

func TestLoginSuccess(t *testing.T) {
	hasher := authInfra.NewPasswordHasher()
	hash, err := hasher.Hash(auth.RawPassword("admin123"))
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	username, _ := auth.NewUsername("admin")
	email, _ := auth.NewEmail("admin@example.com")
	account := auth.RehydrateAccount(auth.User{
		ID:           7,
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		IsActive:     true,
	}, []string{"user"})

	sessions := &fakeSessionRepository{active: map[string]bool{}}
	jwtManager := authInfra.NewJWTManager(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	})
	service := auth.NewService(&fakeAccountRepository{account: account}, sessions, hasher, jwtManager, jwtManager, 24*time.Hour)

	tokens, err := service.Login(t.Context(), auth.LoginInput{
		Email:    "admin@example.com",
		Password: "admin123",
	})
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}

	principal, err := service.Authenticate(t.Context(), tokens.AccessToken)
	if err != nil {
		t.Fatalf("unexpected auth error: %v", err)
	}
	if principal.UserID != 7 {
		t.Fatalf("expected user id 7, got %d", principal.UserID)
	}
	if len(sessions.createdSessions) != 1 {
		t.Fatalf("expected one session to be created, got %d", len(sessions.createdSessions))
	}
}

func TestLoginUnavailableWithoutRepository(t *testing.T) {
	jwtManager := authInfra.NewJWTManager(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	})
	service := auth.NewService(nil, &fakeSessionRepository{}, authInfra.NewPasswordHasher(), jwtManager, jwtManager, 24*time.Hour)

	if _, err := service.Login(t.Context(), auth.LoginInput{Email: "admin@example.com", Password: "admin123"}); !errors.Is(err, auth.ErrDependencyUnavailable) {
		t.Fatalf("expected dependency unavailable, got %v", err)
	}
}

func TestAuthenticateFailsForRevokedSession(t *testing.T) {
	jwtManager := authInfra.NewJWTManager(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	})
	rawToken, err := jwtManager.Issue(auth.AuthenticatedUser{
		UserID:    42,
		Username:  "admin",
		SessionID: "session-1",
	})
	if err != nil {
		t.Fatalf("unexpected sign error: %v", err)
	}

	service := auth.NewService(nil, &fakeSessionRepository{active: map[string]bool{"session-1": false}}, authInfra.NewPasswordHasher(), jwtManager, jwtManager, 24*time.Hour)
	if _, err := service.Authenticate(t.Context(), rawToken); !errors.Is(err, auth.ErrSessionRevoked) {
		t.Fatalf("expected session revoked, got %v", err)
	}
}
