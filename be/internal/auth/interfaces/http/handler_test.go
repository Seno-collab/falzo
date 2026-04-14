package httpapi

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/application/query"
)

type fakeService struct {
	registerErr   error
	loginTokens   query.TokenPair
	refreshTokens query.TokenPair
	loginErr      error
}

func (f fakeService) Register(ctx context.Context, cmd command.Register) error {
	return f.registerErr
}

func (f fakeService) Login(ctx context.Context, cmd command.Login) (query.TokenPair, error) {
	return f.loginTokens, f.loginErr
}

func (f fakeService) Refresh(ctx context.Context, cmd command.Refresh) (query.TokenPair, error) {
	return f.refreshTokens, nil
}

func (f fakeService) Logout(ctx context.Context, cmd command.Logout) error {
	return nil
}

func (f fakeService) Authenticate(ctx context.Context, token string) (*query.AuthenticatedUser, error) {
	return nil, errors.New("not implemented")
}

func TestRegisterHandler(t *testing.T) {
	handler := New(fakeService{})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"user_name":"admin","email":"admin@example.com","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestLoginHandler(t *testing.T) {
	handler := New(fakeService{loginTokens: query.TokenPair{AccessToken: "signed-token", RefreshToken: "refresh-token", TokenType: "Bearer"}})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"admin@example.com","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRefreshHandler(t *testing.T) {
	handler := New(fakeService{refreshTokens: query.TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh", TokenType: "Bearer"}})

	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(`{"refresh_token":"refresh-token"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
