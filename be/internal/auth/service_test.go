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
	account            *auth.Account
	saveErr            error
	findErr            error
	updatePasswordHash auth.PasswordHash
	updatePasswordErr  error
	updateAvatarURL    auth.AvatarURL
	updateAvatarErr    error
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

func (f *fakeAccountRepository) FindActiveByID(ctx context.Context, userID uint64) (*auth.Account, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if f.account == nil || f.account.User.ID != userID {
		return nil, auth.ErrInvalidCredentials
	}
	return f.account, nil
}

func (f *fakeAccountRepository) UpdatePasswordHash(ctx context.Context, userID uint64, passwordHash auth.PasswordHash) error {
	if f.updatePasswordErr != nil {
		return f.updatePasswordErr
	}
	f.updatePasswordHash = passwordHash
	return nil
}

func (f *fakeAccountRepository) UpdateAvatarURL(ctx context.Context, userID uint64, avatarURL auth.AvatarURL) error {
	if f.updateAvatarErr != nil {
		return f.updateAvatarErr
	}
	f.updateAvatarURL = avatarURL
	if f.account != nil && f.account.User.ID == userID {
		f.account.User.AvatarURL = avatarURL
	}
	return nil
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

func (f *fakeSessionRepository) CleanupExpired(ctx context.Context, retention time.Duration) (int64, error) {
	return 0, nil
}

func (f *fakeSessionRepository) SessionCleanupConfig(ctx context.Context) (auth.SessionCleanupConfig, error) {
	return auth.SessionCleanupConfig{}, nil
}

func (f *fakeSessionRepository) WaitSessionCleanupConfigChange(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
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

func TestChangePasswordSuccess(t *testing.T) {
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

	accounts := &fakeAccountRepository{account: account}
	service := auth.NewService(accounts, nil, hasher, nil, nil, 24*time.Hour)

	err = service.ChangePassword(t.Context(), auth.ChangePasswordInput{
		UserID:          7,
		CurrentPassword: "admin123",
		NewPassword:     "newpass123",
	})
	if err != nil {
		t.Fatalf("unexpected change password error: %v", err)
	}

	if accounts.updatePasswordHash == "" {
		t.Fatal("expected password hash to be updated")
	}
	if err := hasher.Compare(accounts.updatePasswordHash, auth.RawPassword("newpass123")); err != nil {
		t.Fatalf("updated hash does not match new password: %v", err)
	}
}

func TestChangePasswordRejectsInvalidCurrentPassword(t *testing.T) {
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

	accounts := &fakeAccountRepository{account: account}
	service := auth.NewService(accounts, nil, hasher, nil, nil, 24*time.Hour)

	err = service.ChangePassword(t.Context(), auth.ChangePasswordInput{
		UserID:          7,
		CurrentPassword: "wrongpass123",
		NewPassword:     "newpass123",
	})
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if accounts.updatePasswordHash != "" {
		t.Fatal("expected password hash to remain unchanged")
	}
}

func TestUpdateAvatarSuccess(t *testing.T) {
	username, _ := auth.NewUsername("admin")
	email, _ := auth.NewEmail("admin@example.com")
	hash, _ := auth.NewPasswordHash("hash")
	account := auth.RehydrateAccount(auth.User{
		ID:           7,
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		IsActive:     true,
	}, []string{"user"})

	accounts := &fakeAccountRepository{account: account}
	service := auth.NewService(accounts, nil, nil, nil, nil, 24*time.Hour)

	profile, err := service.UpdateAvatar(t.Context(), auth.UpdateAvatarInput{
		UserID:    7,
		AvatarURL: "https://example.com/avatar.jpg",
	})
	if err != nil {
		t.Fatalf("unexpected update avatar error: %v", err)
	}

	if accounts.updateAvatarURL.String() != "https://example.com/avatar.jpg" {
		t.Fatalf("expected avatar url to be updated, got %q", accounts.updateAvatarURL.String())
	}
	if profile.AvatarURL != "https://example.com/avatar.jpg" {
		t.Fatalf("expected profile avatar url, got %q", profile.AvatarURL)
	}
}

func TestUpdateAvatarRejectsInvalidURL(t *testing.T) {
	service := auth.NewService(&fakeAccountRepository{}, nil, nil, nil, nil, 24*time.Hour)

	_, err := service.UpdateAvatar(t.Context(), auth.UpdateAvatarInput{
		UserID:    7,
		AvatarURL: "not-a-url",
	})
	if !errors.Is(err, auth.ErrInvalidAvatarURL) {
		t.Fatalf("expected invalid avatar url, got %v", err)
	}
}
