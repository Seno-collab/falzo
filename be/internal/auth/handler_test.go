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
	registerErr   error
	loginTokens   TokenPair
	refreshTokens TokenPair
	loginErr      error
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

func (f fakeService) Authenticate(ctx context.Context, token string) (*AuthenticatedUser, error) {
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
