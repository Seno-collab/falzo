package service

import (
	"context"
	"database/sql"
	"errors"

	"falzo/internal/auth"

	"golang.org/x/crypto/bcrypt"
)

func (s *authService) Login(ctx context.Context, username string, password string) (string, error) {
	if s.db == nil || s.db.DB() == nil {
		return "", auth.ErrAuthUnavailable
	}

	var (
		userID       uint64
		passwordHash string
	)

	err := s.db.DB().QueryRowContext(ctx, `
		SELECT id, password_hash
		FROM users
		WHERE username = ? AND is_active = TRUE
		LIMIT 1
	`, username).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", auth.ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return "", auth.ErrInvalidCredentials
	}

	return s.signToken(userID, username)
}

func (s *authService) Logout(ctx context.Context, token string) error {
	return nil
}
