package service

import (
	"testing"
	"time"

	"falzo/internal/auth"
	"falzo/internal/config"
)

func TestSignTokenAndParseToken(t *testing.T) {
	service := &authService{
		cfg: config.AuthConfig{
			JWTSecret: "test-secret",
			TokenTTL:  time.Hour,
		},
	}

	token, err := service.signToken(42, "admin")
	if err != nil {
		t.Fatalf("unexpected sign error: %v", err)
	}

	claims, err := service.ParseToken(token)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if claims.UserID != 42 {
		t.Fatalf("expected user id 42, got %d", claims.UserID)
	}

	if claims.Username != "admin" {
		t.Fatalf("expected username admin, got %q", claims.Username)
	}
}

func TestLoginUnavailableWithoutDatabase(t *testing.T) {
	service := New(config.AuthConfig{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	}, nil)

	if _, err := service.Login(t.Context(), "admin", "admin123"); err != auth.ErrAuthUnavailable {
		t.Fatalf("expected auth unavailable, got %v", err)
	}
}
