package service

import (
	"strconv"
	"time"

	"falzo/internal/auth"

	"github.com/golang-jwt/jwt/v5"
)

func (s *authService) ParseToken(token string) (*auth.Claims, error) {
	claims := &auth.Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, auth.ErrInvalidToken
		}

		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, auth.ErrInvalidToken
	}

	return claims, nil
}

func (s *authService) signToken(userID uint64, username string) (string, error) {
	now := time.Now()
	claims := auth.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.TokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
