package token

import (
	"strconv"
	"time"

	"falzo/internal/auth/application"
	"falzo/internal/auth/application/query"
	"falzo/internal/auth/domain"
	"falzo/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	cfg config.AuthConfig
}

type claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewJWTManager(cfg config.AuthConfig) *JWTManager {
	return &JWTManager{cfg: cfg}
}

func (m *JWTManager) Issue(principal query.AuthenticatedUser) (string, error) {
	now := time.Now()
	tokenClaims := claims{
		UserID:   principal.UserID,
		Username: principal.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(principal.UserID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.cfg.TokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return token.SignedString([]byte(m.cfg.JWTSecret))
}

func (m *JWTManager) Authenticate(rawToken string) (*query.AuthenticatedUser, error) {
	tokenClaims := &claims{}
	parsed, err := jwt.ParseWithClaims(rawToken, tokenClaims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, domain.ErrInvalidToken
		}

		return []byte(m.cfg.JWTSecret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, domain.ErrInvalidToken
	}

	var expiresAt *time.Time
	if tokenClaims.ExpiresAt != nil {
		value := tokenClaims.ExpiresAt.Time
		expiresAt = &value
	}

	return &query.AuthenticatedUser{
		UserID:    tokenClaims.UserID,
		Username:  tokenClaims.Username,
		Subject:   tokenClaims.Subject,
		ExpiresAt: expiresAt,
	}, nil
}

var _ application.TokenIssuer = (*JWTManager)(nil)
var _ application.TokenAuthenticator = (*JWTManager)(nil)
