package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"falzo/internal/auth"
)

type fakeService struct {
	registerErr error
	loginToken  string
	loginErr    error
}

func (f fakeService) Register(ctx context.Context, username string, email string, password string) error {
	return f.registerErr
}

func (f fakeService) Login(ctx context.Context, username string, password string) (string, error) {
	return f.loginToken, f.loginErr
}

func (f fakeService) Logout(ctx context.Context, token string) error {
	return nil
}

func (f fakeService) ParseToken(token string) (*auth.Claims, error) {
	return nil, errors.New("not implemented")
}

func TestRegisterHandler(t *testing.T) {
	handler := New(fakeService{})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"username":"admin","email":"admin@example.com","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestLoginHandler(t *testing.T) {
	handler := New(fakeService{loginToken: "signed-token"})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"username":"admin","password":"admin123"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
