package auth

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeService struct {
	registerErr         error
	loginTokens         TokenPair
	refreshTokens       TokenPair
	loginErr            error
	changePasswordInput ChangePasswordInput
	changePasswordErr   error
	profile             UserProfile
	updateAvatarInput   UpdateAvatarInput
	updateAvatarErr     error
}

func (f fakeService) Register(ctx context.Context, input RegisterInput) error {
	return f.registerErr
}

func (f fakeService) Login(ctx context.Context, input LoginInput) (TokenPair, error) {
	return f.loginTokens, f.loginErr
}

func (f fakeService) Refresh(ctx context.Context, input RefreshInput) (TokenPair, error) {
	return f.refreshTokens, nil
}

func (f fakeService) Logout(ctx context.Context, input LogoutInput) error {
	return nil
}

func (f fakeService) ChangePassword(ctx context.Context, input ChangePasswordInput) error {
	return f.changePasswordErr
}

func (f fakeService) Profile(ctx context.Context, userID uint64) (UserProfile, error) {
	if f.profile.UserID != 0 {
		return f.profile, nil
	}

	return UserProfile{
		UserID:        userID,
		UserIDAlias:   userID,
		Username:      "admin",
		UsernameAlias: "admin",
		Email:         "admin@example.com",
	}, nil
}

func (f fakeService) UpdateAvatar(ctx context.Context, input UpdateAvatarInput) (UserProfile, error) {
	if f.updateAvatarErr != nil {
		return UserProfile{}, f.updateAvatarErr
	}

	return UserProfile{
		UserID:         input.UserID,
		UserIDAlias:    input.UserID,
		Username:       "admin",
		UsernameAlias:  "admin",
		Email:          "admin@example.com",
		AvatarURL:      input.AvatarURL,
		AvatarURLAlias: input.AvatarURL,
	}, nil
}

func (f fakeService) Authenticate(ctx context.Context, token string) (*AuthenticatedUser, error) {
	if token == "signed-token" {
		return &AuthenticatedUser{
			UserID:    7,
			Username:  "admin",
			SessionID: "session-1",
		}, nil
	}

	return nil, errors.New("not implemented")
}

func TestRegisterHandler(t *testing.T) {
	handler := NewHandler(fakeService{})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"user_name":"admin","email":"admin@example.com","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestLoginHandler(t *testing.T) {
	handler := NewHandler(fakeService{loginTokens: TokenPair{AccessToken: "signed-token", RefreshToken: "refresh-token", TokenType: "Bearer"}})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"admin@example.com","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRefreshHandler(t *testing.T) {
	handler := NewHandler(fakeService{refreshTokens: TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh", TokenType: "Bearer"}})

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(`{"refresh_token":"refresh-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestChangePasswordHandler(t *testing.T) {
	handler := NewHandler(fakeService{})

	req := httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewBufferString(`{"current_password":"admin123","new_password":"newpass123"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer signed-token")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUpdateAvatarHandler(t *testing.T) {
	handler := NewHandler(fakeService{})

	req := httptest.NewRequest(http.MethodPatch, "/me/avatar", bytes.NewBufferString(`{"avatar_url":"https://example.com/avatar.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer signed-token")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
