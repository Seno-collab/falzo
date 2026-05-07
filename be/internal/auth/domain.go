package auth

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidToken           = errors.New("invalid token")
	ErrSessionRevoked         = errors.New("session revoked or expired")
	ErrUserExists             = errors.New("user already exists")
	ErrDependencyUnavailable  = errors.New("auth dependency unavailable")
	ErrInternal               = errors.New("auth internal error")
	ErrTemporarilyUnavailable = errors.New("auth temporarily unavailable")
	ErrInvalidEmail           = errors.New("Invalid email")
	ErrInvalidPassword        = errors.New("Invalid password")
	ErrInvalidPasswordHash    = errors.New("invalid password hash")
	ErrInvalidUsername        = errors.New("invalid username")
)

type AccountRepository interface {
	Save(ctx context.Context, account *Account) error
	FindActiveByEmail(ctx context.Context, email Email) (*Account, error)
	FindActiveByID(ctx context.Context, userID uint64) (*Account, error)
	UpdatePasswordHash(ctx context.Context, userID uint64, passwordHash PasswordHash) error
}

type SessionRepository interface {
	Create(ctx context.Context, session Session) error
	IsSessionActive(ctx context.Context, sessionID string) (bool, error)
	FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*Session, error)
	RotateRefreshToken(ctx context.Context, session Session, newRefreshTokenHash string, expiresAtUnix int64) error
	RevokeBySessionID(ctx context.Context, sessionID string) error
	CleanupExpired(ctx context.Context, retention time.Duration) (int64, error)
	SessionCleanupConfig(ctx context.Context) (SessionCleanupConfig, error)
	WaitSessionCleanupConfigChange(ctx context.Context) error
}

type PasswordHasher interface {
	Hash(password RawPassword) (PasswordHash, error)
	Compare(hash PasswordHash, password RawPassword) error
}

type TokenIssuer interface {
	Issue(principal AuthenticatedUser) (string, error)
}

type TokenAuthenticator interface {
	Authenticate(rawToken string) (*AuthenticatedUser, error)
}

type User struct {
	ID           uint64
	Username     Username
	Email        Email
	PasswordHash PasswordHash
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Account struct {
	User  User
	Roles []string
}

func NewAccount(username Username, email Email, passwordHash PasswordHash, roles []string) *Account {
	now := time.Now().UTC()
	return &Account{
		User: User{
			Username:     username,
			Email:        email,
			PasswordHash: passwordHash,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		Roles: append([]string(nil), roles...),
	}
}

func RehydrateAccount(user User, roles []string) *Account {
	return &Account{
		User:  user,
		Roles: append([]string(nil), roles...),
	}
}

type Username string

func NewUsername(raw string) (Username, error) {
	value := strings.TrimSpace(raw)
	if len(value) < 3 || len(value) > 50 {
		return "", ErrInvalidUsername
	}

	return Username(value), nil
}

func (u Username) String() string { return string(u) }

type Email string

func NewEmail(raw string) (Email, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" || !strings.Contains(value, "@") {
		return "", ErrInvalidEmail
	}

	return Email(value), nil
}

func (e Email) String() string { return string(e) }

type RawPassword string

func NewRawPassword(raw string) (RawPassword, error) {
	if len(raw) < 8 {
		return "", ErrInvalidPassword
	}

	hasLetter := false
	hasDigit := false
	for _, r := range raw {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "", ErrInvalidPassword
	}

	return RawPassword(raw), nil
}

func (p RawPassword) String() string { return string(p) }

type PasswordHash string

func NewPasswordHash(raw string) (PasswordHash, error) {
	if raw == "" {
		return "", ErrInvalidPasswordHash
	}

	return PasswordHash(raw), nil
}

func (p PasswordHash) String() string { return string(p) }

type AuthenticatedUser struct {
	UserID    uint64
	Username  string
	Subject   string
	SessionID string
	ExpiresAt *time.Time
}

type Session struct {
	SessionID            string
	UserID               uint64
	Username             string
	Subject              string
	RefreshTokenHash     string
	RefreshExpiresAtUnix int64
}

type SessionCleanupConfig struct {
	Enabled   bool
	Interval  time.Duration
	Retention time.Duration
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}
