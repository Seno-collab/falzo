package infra

import (
	"context"
	"encoding/json"
	"time"

	"falzo-be/internal/auth"
	pkgcache "falzo-be/pkg/cache"

	goredis "github.com/redis/go-redis/v9"
)

const (
	sessionActiveCacheValue   = "1"
	sessionInactiveCacheValue = "0"
)

type sessionCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

type redisSessionCache struct {
	client *goredis.Client
}

func (c *redisSessionCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *redisSessionCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *redisSessionCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(ctx, keys...).Err()
}

type CachedSessionRepository struct {
	next      auth.SessionRepository
	cache     sessionCache
	activeTTL time.Duration
}

func NewCachedSessionRepository(next auth.SessionRepository, cache pkgcache.Client, cacheTTL time.Duration) auth.SessionRepository {
	if cache == nil || cache.Client() == nil {
		return next
	}

	return &CachedSessionRepository{
		next:      next,
		cache:     &redisSessionCache{client: cache.Client()},
		activeTTL: cacheTTL,
	}
}

func (r *CachedSessionRepository) Create(ctx context.Context, session auth.Session) error {
	if err := r.next.Create(ctx, session); err != nil {
		return err
	}

	r.cacheSessionActive(ctx, session.SessionID, true)
	r.cacheRefreshSession(ctx, session)
	return nil
}

func (r *CachedSessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	value, err := r.cache.Get(ctx, sessionCacheKey(sessionID))
	if err == nil {
		return value == sessionActiveCacheValue, nil
	}

	active, err := r.next.IsSessionActive(ctx, sessionID)
	if err != nil {
		return false, err
	}

	r.cacheSessionActive(ctx, sessionID, active)
	return active, nil
}

func (r *CachedSessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*auth.Session, error) {
	cached, err := r.cache.Get(ctx, refreshSessionCacheKey(refreshTokenHash))
	if err == nil {
		session, decodeErr := decodeCachedSession(cached)
		if decodeErr == nil {
			return session, nil
		}

		_ = r.cache.Del(ctx, refreshSessionCacheKey(refreshTokenHash))
	}

	session, err := r.next.FindActiveByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		return nil, err
	}

	r.cacheSessionActive(ctx, session.SessionID, true)
	r.cacheRefreshSession(ctx, *session)
	return session, nil
}

func (r *CachedSessionRepository) RotateRefreshToken(ctx context.Context, session auth.Session, newRefreshTokenHash string, expiresAtUnix int64) error {
	if err := r.next.RotateRefreshToken(ctx, session, newRefreshTokenHash, expiresAtUnix); err != nil {
		return err
	}

	r.cacheSessionActive(ctx, session.SessionID, true)
	_ = r.cache.Del(ctx, refreshSessionCacheKey(session.RefreshTokenHash))

	session.RefreshTokenHash = newRefreshTokenHash
	session.RefreshExpiresAtUnix = expiresAtUnix
	r.cacheRefreshSession(ctx, session)
	return nil
}

func (r *CachedSessionRepository) RevokeBySessionID(ctx context.Context, sessionID string) error {
	if err := r.next.RevokeBySessionID(ctx, sessionID); err != nil {
		return err
	}

	r.cacheSessionActive(ctx, sessionID, false)

	refreshTokenHash, err := r.cache.Get(ctx, sessionRefreshHashCacheKey(sessionID))
	if err == nil && refreshTokenHash != "" {
		_ = r.cache.Del(ctx, refreshSessionCacheKey(refreshTokenHash))
	}
	_ = r.cache.Del(ctx, sessionRefreshHashCacheKey(sessionID))

	return nil
}

func (r *CachedSessionRepository) cacheSessionActive(ctx context.Context, sessionID string, active bool) {
	cacheValue := sessionInactiveCacheValue
	if active {
		cacheValue = sessionActiveCacheValue
	}

	_ = r.cache.Set(ctx, sessionCacheKey(sessionID), cacheValue, r.activeTTL)
}

func (r *CachedSessionRepository) cacheRefreshSession(ctx context.Context, session auth.Session) {
	ttl := ttlUntilUnix(session.RefreshExpiresAtUnix)
	if ttl <= 0 {
		return
	}

	payload, err := json.Marshal(session)
	if err != nil {
		return
	}

	_ = r.cache.Set(ctx, refreshSessionCacheKey(session.RefreshTokenHash), string(payload), ttl)
	_ = r.cache.Set(ctx, sessionRefreshHashCacheKey(session.SessionID), session.RefreshTokenHash, ttl)
}

func decodeCachedSession(payload string) (*auth.Session, error) {
	var session auth.Session
	if err := json.Unmarshal([]byte(payload), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func ttlUntilUnix(expiresAtUnix int64) time.Duration {
	return time.Until(time.Unix(expiresAtUnix, 0))
}

func sessionCacheKey(sessionID string) string {
	return "auth:session:active:" + sessionID
}

func refreshSessionCacheKey(refreshTokenHash string) string {
	return "auth:refresh:session:" + refreshTokenHash
}

func sessionRefreshHashCacheKey(sessionID string) string {
	return "auth:session:refresh_hash:" + sessionID
}

var _ auth.SessionRepository = (*CachedSessionRepository)(nil)
