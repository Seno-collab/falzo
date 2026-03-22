package repository

import (
	"context"

	"falzo-be/internal/auth/application/query"
)

type SessionRepository interface {
	Create(ctx context.Context, session query.Session) error
	IsSessionActive(ctx context.Context, sessionID string) (bool, error)
	FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*query.Session, error)
	RotateRefreshToken(ctx context.Context, sessionID string, refreshTokenHash string, expiresAtUnix int64) error
	RevokeBySessionID(ctx context.Context, sessionID string) error
}
