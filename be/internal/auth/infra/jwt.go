package infra

import (
	"strconv"
	"time"

	"falzo-be/internal/auth"
	"falzo-be/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	cfg config.AuthConfig
}

type claims struct {
	UserID    uint64   `json:"user_id"`
	Username  string   `json:"user_name"`
	Roles     []string `json:"roles,omitempty"`
	SessionID string   `json:"session_id"`
	jwt.RegisteredClaims
}

func NewJWTManager(cfg config.AuthConfig) *JWTManager {
	return &JWTManager{cfg: cfg}
}

func (m *JWTManager) Issue(principal auth.AuthenticatedUser) (string, error) {
	now := time.Now()
	tokenClaims := claims{
		UserID:    principal.UserID,
		Username:  principal.Username,
		Roles:     append([]string(nil), principal.Roles...),
		SessionID: principal.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        principal.SessionID,
			Subject:   strconv.FormatUint(principal.UserID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.cfg.TokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return token.SignedString([]byte(m.cfg.JWTSecret))
}

func (m *JWTManager) Authenticate(rawToken string) (*auth.AuthenticatedUser, error) {
	tokenClaims := &claims{}
	parsed, err := jwt.ParseWithClaims(rawToken, tokenClaims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, auth.ErrInvalidToken
		}

		return []byte(m.cfg.JWTSecret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, auth.ErrInvalidToken
	}

	var expiresAt *time.Time
	if tokenClaims.ExpiresAt != nil {
		value := tokenClaims.ExpiresAt.Time
		expiresAt = &value
	}

	return &auth.AuthenticatedUser{
		UserID:    tokenClaims.UserID,
		Username:  tokenClaims.Username,
		Roles:     append([]string(nil), tokenClaims.Roles...),
		Subject:   tokenClaims.Subject,
		SessionID: tokenClaims.SessionID,
		ExpiresAt: expiresAt,
	}, nil
}

var _ auth.TokenIssuer = (*JWTManager)(nil)
var _ auth.TokenAuthenticator = (*JWTManager)(nil)
