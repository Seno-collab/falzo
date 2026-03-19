package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"falzo/internal/config"
	"falzo/internal/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Chi defines authentication behaviors used by handlers and middleware.
type Chi interface {
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

// em implements Chi with MySQL-backed users and JWT-based authentication.
type em struct {
	cfg config.AuthConfig
	db  database.Chi
}

func New(cfg config.AuthConfig, db database.Chi) Chi {
	return &em{cfg: cfg, db: db}
}

func (s *em) Register(ctx context.Context, username string, email string, password string) error {
	if s.db == nil || s.db.DB() == nil {
		return ErrAuthUnavailable
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
			return ErrUserExists
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

func (s *em) Login(ctx context.Context, username string, password string) (string, error) {
	if s.db == nil || s.db.DB() == nil {
		return "", ErrAuthUnavailable
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
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	return s.signToken(userID, username)
}

func (s *em) Logout(ctx context.Context, token string) error {
	return nil
}

func (s *em) ParseToken(token string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}

		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *em) signToken(userID uint64, username string) (string, error) {
	now := time.Now()
	claims := Claims{
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

func isDuplicateError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}
