package security

import (
	"be/internal/application/ports"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	tokenTypeAccess        = "access"
	tokenTypeRefresh       = "refresh"
	tokenTypePasswordReset = "password_reset"
)

type JWTTokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	resetTTL   time.Duration
}

type jwtClaims struct {
	Subject   string `json:"sub"`
	UserName  string `json:"username"`
	Type      string `json:"typ"`
	TokenID   string `json:"jti"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func NewJWTTokenManager(secret string, accessTTL, refreshTTL, resetTTL time.Duration) (*JWTTokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}
	if accessTTL <= 0 || refreshTTL <= 0 || resetTTL <= 0 {
		return nil, errors.New("token TTL must be positive")
	}
	return &JWTTokenManager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL, resetTTL: resetTTL}, nil
}

func (m *JWTTokenManager) GeneratePair(userID int64, username string, now time.Time) (ports.TokenPair, error) {
	accessID, err := randomID()
	if err != nil {
		return ports.TokenPair{}, err
	}
	refreshID, err := randomID()
	if err != nil {
		return ports.TokenPair{}, err
	}
	accessExpiry := now.Add(m.accessTTL)
	refreshExpiry := now.Add(m.refreshTTL)
	access, err := m.sign(userID, username, tokenTypeAccess, accessID, now, accessExpiry)
	if err != nil {
		return ports.TokenPair{}, err
	}
	refresh, err := m.sign(userID, username, tokenTypeRefresh, refreshID, now, refreshExpiry)
	if err != nil {
		return ports.TokenPair{}, err
	}
	return ports.TokenPair{AccessToken: access, RefreshToken: refresh, RefreshTokenID: refreshID, AccessExpiresAt: accessExpiry, RefreshExpiresAt: refreshExpiry}, nil
}

func (m *JWTTokenManager) ParseRefresh(token string, now time.Time) (ports.TokenClaims, error) {
	return m.parse(token, tokenTypeRefresh, now)
}

func (m *JWTTokenManager) ParseAccess(token string, now time.Time) (ports.TokenClaims, error) {
	return m.parse(token, tokenTypeAccess, now)
}

func (m *JWTTokenManager) GeneratePasswordReset(userID int64, username string, now time.Time) (string, ports.TokenClaims, error) {
	id, err := randomID()
	if err != nil {
		return "", ports.TokenClaims{}, err
	}
	expires := now.Add(m.resetTTL)
	token, err := m.sign(userID, username, tokenTypePasswordReset, id, now, expires)
	return token, ports.TokenClaims{UserID: userID, UserName: username, TokenID: id, ExpiresAt: expires}, err
}

func (m *JWTTokenManager) ParsePasswordReset(token string, now time.Time) (ports.TokenClaims, error) {
	return m.parse(token, tokenTypePasswordReset, now)
}

func (m *JWTTokenManager) sign(userID int64, username, tokenType, tokenID string, issuedAt, expiresAt time.Time) (string, error) {
	header, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payload, err := json.Marshal(jwtClaims{Subject: strconv.FormatInt(userID, 10), UserName: username, Type: tokenType, TokenID: tokenID, IssuedAt: issuedAt.Unix(), ExpiresAt: expiresAt.Unix()})
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func (m *JWTTokenManager) parse(token, expectedType string, now time.Time) (ports.TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ports.TokenClaims{}, errors.New("malformed token")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ports.TokenClaims{}, errors.New("malformed token header")
	}
	var header map[string]string
	if json.Unmarshal(headerBytes, &header) != nil || header["alg"] != "HS256" {
		return ports.TokenClaims{}, errors.New("unsupported token algorithm")
	}
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || subtle.ConstantTimeCompare(mac.Sum(nil), actual) != 1 {
		return ports.TokenClaims{}, errors.New("invalid token signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ports.TokenClaims{}, errors.New("malformed token payload")
	}
	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ports.TokenClaims{}, errors.New("malformed token claims")
	}
	if claims.Type != expectedType || claims.TokenID == "" || claims.ExpiresAt <= now.Unix() {
		return ports.TokenClaims{}, errors.New("invalid or expired token")
	}
	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return ports.TokenClaims{}, fmt.Errorf("invalid token subject")
	}
	return ports.TokenClaims{UserID: userID, UserName: claims.UserName, TokenID: claims.TokenID, ExpiresAt: time.Unix(claims.ExpiresAt, 0).UTC()}, nil
}

func randomID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
