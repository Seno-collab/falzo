package application_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"falzo-be/internal/auth/application"
	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain"
	"falzo-be/internal/auth/domain/aggregate"
	"falzo-be/internal/auth/domain/entity"
	"falzo-be/internal/auth/domain/value_object"
	"falzo-be/internal/auth/infrastructure/security/bcrypt"
	"falzo-be/internal/auth/infrastructure/token"
	"falzo-be/pkg/config"
	"testing"
	"time"
)

type fakeAccountRepository struct {
	account *aggregate.Account
	saveErr error
	findErr error
}

type fakeSessionRepository struct {
	createdSessions []query.Session
	active          map[string]bool
	refreshSessions map[string]query.Session
	revokedIDs      []string
	createErr       error
	activeErr       error
	revokeErr       error
	findErr         error
	rotateErr       error
}

func (f *fakeAccountRepository) Save(ctx context.Context, account *aggregate.Account) error {
	f.account = account
	return f.saveErr
}

func (f *fakeAccountRepository) FindActiveByEmail(ctx context.Context, email value_object.Email) (*aggregate.Account, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}

	return f.account, nil
}

func (f *fakeSessionRepository) Create(ctx context.Context, session query.Session) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.createdSessions = append(f.createdSessions, session)
	if f.active == nil {
		f.active = map[string]bool{}
	}
	f.active[session.SessionID] = true
	if f.refreshSessions == nil {
		f.refreshSessions = map[string]query.Session{}
	}
	f.refreshSessions[session.RefreshTokenHash] = session
	return nil
}

func (f *fakeSessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	if f.activeErr != nil {
		return false, f.activeErr
	}
	return f.active[sessionID], nil
}

func (f *fakeSessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*query.Session, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	session, ok := f.refreshSessions[refreshTokenHash]
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	return &session, nil
}

func (f *fakeSessionRepository) RotateRefreshToken(ctx context.Context, session query.Session, newRefreshTokenHash string, expiresAtUnix int64) error {
	if f.rotateErr != nil {
		return f.rotateErr
	}
	for key, currentSession := range f.refreshSessions {
		if currentSession.SessionID == session.SessionID && currentSession.RefreshTokenHash == session.RefreshTokenHash {
			delete(f.refreshSessions, key)
			currentSession.RefreshTokenHash = newRefreshTokenHash
			currentSession.RefreshExpiresAtUnix = expiresAtUnix
			f.refreshSessions[newRefreshTokenHash] = currentSession
			return nil
		}
	}
	return domain.ErrInvalidToken
}

func (f *fakeSessionRepository) RevokeBySessionID(ctx context.Context, sessionID string) error {
	if f.revokeErr != nil {
		return f.revokeErr
	}
	f.revokedIDs = append(f.revokedIDs, sessionID)
	if f.active != nil {
		f.active[sessionID] = false
	}
	return nil
}

func TestAuthenticateToken(t *testing.T) {
	jwtManager := token.NewJWTManager(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	})

	service := application.New(nil, &fakeSessionRepository{active: map[string]bool{}}, bcrypt.NewPasswordHasher(), jwtManager, jwtManager, 24*time.Hour)

	rawToken, err := jwtManager.Issue(query.AuthenticatedUser{
		UserID:    42,
		Username:  "admin",
		SessionID: "session-1",
	})
	if err != nil {
		t.Fatalf("unexpected sign error: %v", err)
	}

	sessions := &fakeSessionRepository{active: map[string]bool{"session-1": true}}
	service = application.New(nil, sessions, bcrypt.NewPasswordHasher(), jwtManager, jwtManager, 24*time.Hour)

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

	service := application.New(nil, &fakeSessionRepository{active: map[string]bool{}}, bcrypt.NewPasswordHasher(), jwtManager, jwtManager, 24*time.Hour)

	if _, err := service.Login(t.Context(), command.Login{Email: "admin@example.com", Password: "admin123"}); err != domain.ErrAuthDependencyUnavailable {
		t.Fatalf("expected auth dependency unavailable, got %v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	hasher := bcrypt.NewPasswordHasher()
	hash, err := hasher.Hash(value_object.RawPassword("admin123"))
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	username, _ := value_object.NewUsername("admin")
	email, _ := value_object.NewEmail("admin@example.com")
	account := aggregate.RehydrateAccount(entity.User{
		ID:           7,
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		IsActive:     true,
	}, []string{"user"})

	repo := &fakeAccountRepository{account: account}
	sessions := &fakeSessionRepository{active: map[string]bool{}}
	jwtManager := token.NewJWTManager(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	})

	service := application.New(repo, sessions, hasher, jwtManager, jwtManager, 24*time.Hour)

	tokens, err := service.Login(t.Context(), command.Login{
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

func TestAuthenticateFailsForRevokedSession(t *testing.T) {
	jwtManager := token.NewJWTManager(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	})

	rawToken, err := jwtManager.Issue(query.AuthenticatedUser{
		UserID:    42,
		Username:  "admin",
		SessionID: "session-1",
	})
	if err != nil {
		t.Fatalf("unexpected sign error: %v", err)
	}

	service := application.New(nil, &fakeSessionRepository{active: map[string]bool{"session-1": false}}, bcrypt.NewPasswordHasher(), jwtManager, jwtManager, 24*time.Hour)

	if _, err := service.Authenticate(t.Context(), rawToken); err != domain.ErrSessionRevoked {
		t.Fatalf("expected session revoked, got %v", err)
	}
}

func TestRefreshRotatesRefreshToken(t *testing.T) {
	jwtManager := token.NewJWTManager(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	})

	expiresAt := time.Now().Add(time.Hour).Unix()
	currentHash := tokenHash("current-refresh")
	sessions := &fakeSessionRepository{
		active: map[string]bool{"session-1": true},
		refreshSessions: map[string]query.Session{
			currentHash: {
				SessionID:            "session-1",
				UserID:               42,
				Username:             "admin",
				Subject:              "42",
				RefreshTokenHash:     currentHash,
				RefreshExpiresAtUnix: expiresAt,
			},
		},
	}

	service := application.New(nil, sessions, bcrypt.NewPasswordHasher(), jwtManager, jwtManager, 24*time.Hour)

	tokens, err := service.Refresh(t.Context(), command.Refresh{RefreshToken: "current-refresh"})
	if err != nil {
		t.Fatalf("unexpected refresh error: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Fatal("expected access token to be returned")
	}

	if tokens.RefreshToken == "" {
		t.Fatal("expected refresh token to be returned")
	}

	if _, ok := sessions.refreshSessions[currentHash]; ok {
		t.Fatal("expected old refresh token hash to be replaced")
	}

	if _, ok := sessions.refreshSessions[tokenHash(tokens.RefreshToken)]; !ok {
		t.Fatal("expected new refresh token hash to be stored")
	}
}

func tokenHash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
