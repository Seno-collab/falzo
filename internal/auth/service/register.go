package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"falzo/internal/auth"

	"golang.org/x/crypto/bcrypt"
)

func (s *authService) Register(ctx context.Context, username string, email string, password string) error {
	if s.db == nil || s.db.DB() == nil {
		return auth.ErrAuthUnavailable
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	tx, err := s.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO users (username, email, password_hash)
		VALUES (?, ?, ?)
	`, username, email, string(passwordHash))
	if err != nil {
		if isDuplicateError(err) {
			return auth.ErrUserExists
		}
		return err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	var roleID uint64
	err = tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = 'user' LIMIT 1`).Scan(&roleID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if roleID != 0 {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			VALUES (?, ?)
			ON DUPLICATE KEY UPDATE role_id = VALUES(role_id)
		`, userID, roleID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func isDuplicateError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}
