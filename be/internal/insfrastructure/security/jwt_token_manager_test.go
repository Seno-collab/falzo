package security

import (
	"strings"
	"testing"
	"time"
)

func TestJWTTokenManagerRoundTripAndTokenTypes(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	m, err := NewJWTTokenManager(strings.Repeat("s", 32), 15*time.Minute, 24*time.Hour, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	pair, err := m.GeneratePair(42, "alice", now)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := m.ParseRefresh(pair.RefreshToken, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 || claims.UserName != "alice" || claims.TokenID != pair.RefreshTokenID {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if _, err := m.ParseAccess(pair.AccessToken, now.Add(time.Minute)); err != nil {
		t.Fatalf("valid access token rejected: %v", err)
	}
	if _, err := m.ParsePasswordReset(pair.RefreshToken, now); err == nil {
		t.Fatal("refresh token accepted as reset token")
	}
	if _, err := m.ParseRefresh(pair.RefreshToken, pair.RefreshExpiresAt); err == nil {
		t.Fatal("expired token accepted")
	}
}

func TestJWTTokenManagerRejectsTamperedTokenAndWeakSecret(t *testing.T) {
	if _, err := NewJWTTokenManager("short", time.Minute, time.Hour, time.Minute); err == nil {
		t.Fatal("short secret accepted")
	}
	m, _ := NewJWTTokenManager(strings.Repeat("s", 32), time.Minute, time.Hour, time.Minute)
	pair, _ := m.GeneratePair(1, "alice", time.Now())
	tampered := pair.RefreshToken[:len(pair.RefreshToken)-1] + "x"
	if _, err := m.ParseRefresh(tampered, time.Now()); err == nil {
		t.Fatal("tampered token accepted")
	}
}
