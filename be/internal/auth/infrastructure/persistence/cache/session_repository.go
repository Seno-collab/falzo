package cache

import (
	"context"
	"strconv"
	"time"

	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain/repository"
	pkgcache "falzo-be/pkg/cache"
)

type SessionRepository struct {
	next     repository.SessionRepository
	cache    pkgcache.Client
	cacheTTL time.Duration
}

func NewSessionRepository(next repository.SessionRepository, cache pkgcache.Client, cacheTTL time.Duration) repository.SessionRepository {
	if cache == nil || cache.Client() == nil {
		return next
	}

	return &SessionRepository{
		next:     next,
		cache:    cache,
		cacheTTL: cacheTTL,
	}
}

func (r *SessionRepository) Create(ctx context.Context, session query.Session) error {
	if err := r.next.Create(ctx, session); err != nil {
		return err
	}

	_ = r.cache.Client().Set(ctx, sessionCacheKey(session.SessionID), "1", r.cacheTTL).Err()
	return nil
}

func (r *SessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	value, err := r.cache.Client().Get(ctx, sessionCacheKey(sessionID)).Result()
	if err == nil {
		return value == "1", nil
	}

	active, err := r.next.IsSessionActive(ctx, sessionID)
	if err != nil {
		return false, err
	}

	cacheValue := "0"
	if active {
		cacheValue = "1"
	}

	_ = r.cache.Client().Set(ctx, sessionCacheKey(sessionID), cacheValue, r.cacheTTL).Err()

	return active, nil
}

func (r *SessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*query.Session, error) {
	return r.next.FindActiveByRefreshTokenHash(ctx, refreshTokenHash)
}

func (r *SessionRepository) RotateRefreshToken(ctx context.Context, sessionID string, refreshTokenHash string, expiresAtUnix int64) error {
	if err := r.next.RotateRefreshToken(ctx, sessionID, refreshTokenHash, expiresAtUnix); err != nil {
		return err
	}

	_ = r.cache.Client().Set(ctx, sessionCacheKey(sessionID), "1", r.cacheTTL).Err()
	return nil
}

func (r *SessionRepository) RevokeBySessionID(ctx context.Context, sessionID string) error {
	if err := r.next.RevokeBySessionID(ctx, sessionID); err != nil {
		return err
	}

	_ = r.cache.Client().Set(ctx, sessionCacheKey(sessionID), "0", r.cacheTTL).Err()
	return nil
}

func sessionCacheKey(sessionID string) string {
	return "auth:session:active:" + sessionID
}

var _ repository.SessionRepository = (*SessionRepository)(nil)

var _ = strconv.IntSize
