package app

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/application/query"
	authHTTP "falzo-be/internal/auth/interfaces/http"

	"github.com/go-chi/chi/v5"
)

type fakeAuthService struct{}

func (fakeAuthService) Register(ctx context.Context, cmd command.Register) error {
	return nil
}

func (fakeAuthService) Login(ctx context.Context, cmd command.Login) (query.TokenPair, error) {
	return query.TokenPair{}, nil
}

func (fakeAuthService) Refresh(ctx context.Context, cmd command.Refresh) (query.TokenPair, error) {
	return query.TokenPair{}, nil
}

func (fakeAuthService) Logout(ctx context.Context, cmd command.Logout) error {
	return nil
}

func (fakeAuthService) Authenticate(ctx context.Context, token string) (*query.AuthenticatedUser, error) {
	return nil, errors.New("not implemented")
}

func TestAuthRoutesMounted(t *testing.T) {
	r := chi.NewRouter()
	r.Mount("/auth", authHTTP.New(fakeAuthService{}).Routes())

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"user_name":"admin","email":"admin@example.com","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}
