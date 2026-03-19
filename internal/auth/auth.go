package auth

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	Register(ctx context.Context, username string, email string, password string) error
	Login(ctx context.Context, username string, password string) (string, error)
	Logout(ctx context.Context, token string) error
	ParseToken(token string) (*Claims, error)
}

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid token")
var ErrUserExists = errors.New("user already exists")
var ErrAuthUnavailable = errors.New("auth database unavailable")

type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
