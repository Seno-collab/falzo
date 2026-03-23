package cache

import (
	"context"
	"testing"
	"time"

	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain/repository"
	pkgcache "falzo-be/pkg/cache"

	goredis "github.com/redis/go-redis/v9"
)

type fakeSessionRepository struct {
	active bool
}

func (f *fakeSessionRepository) Create(ctx context.Context, session query.Session) error { return nil }
func (f *fakeSessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	return f.active, nil
}
func (f *fakeSessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*query.Session, error) {
	return nil, nil
}
func (f *fakeSessionRepository) RotateRefreshToken(ctx context.Context, sessionID string, refreshTokenHash string, expiresAtUnix int64) error {
	return nil
}
func (f *fakeSessionRepository) RevokeBySessionID(ctx context.Context, sessionID string) error {
	return nil
}

type fakeCacheClient struct{ client *goredis.Client }

func (f *fakeCacheClient) Client() *goredis.Client { return f.client }
func (f *fakeCacheClient) Close() error            { return nil }

func TestSessionCacheKey(t *testing.T) {
	if got := sessionCacheKey("abc"); got != "auth:session:active:abc" {
		t.Fatalf("unexpected cache key %q", got)
	}
}

var _ repository.SessionRepository = (*fakeSessionRepository)(nil)
var _ pkgcache.Client = (*fakeCacheClient)(nil)

var _ = time.Second
