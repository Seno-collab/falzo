package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain"
	"falzo-be/internal/auth/domain/repository"
)

type fakeSessionRepository struct {
	active          map[string]bool
	refreshSessions map[string]query.Session
	activeLookups   int
	refreshLookups  int
	revokedIDs      []string
}

func (f *fakeSessionRepository) Create(ctx context.Context, session query.Session) error {
	if f.active == nil {
		f.active = map[string]bool{}
	}
	if f.refreshSessions == nil {
		f.refreshSessions = map[string]query.Session{}
	}

	f.active[session.SessionID] = true
	f.refreshSessions[session.RefreshTokenHash] = session
	return nil
}

func (f *fakeSessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	f.activeLookups++
	return f.active[sessionID], nil
}

func (f *fakeSessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*query.Session, error) {
	f.refreshLookups++
	session, ok := f.refreshSessions[refreshTokenHash]
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	return &session, nil
}

func (f *fakeSessionRepository) RotateRefreshToken(ctx context.Context, session query.Session, newRefreshTokenHash string, expiresAtUnix int64) error {
	currentSession, ok := f.refreshSessions[session.RefreshTokenHash]
	if !ok {
		return domain.ErrInvalidToken
	}

	delete(f.refreshSessions, session.RefreshTokenHash)
	currentSession.RefreshTokenHash = newRefreshTokenHash
	currentSession.RefreshExpiresAtUnix = expiresAtUnix
	f.refreshSessions[newRefreshTokenHash] = currentSession
	return nil
}

func (f *fakeSessionRepository) RevokeBySessionID(ctx context.Context, sessionID string) error {
	f.revokedIDs = append(f.revokedIDs, sessionID)
	if f.active != nil {
		f.active[sessionID] = false
	}
	return nil
}

type memoryCache struct {
	values map[string]memoryCacheEntry
}

type memoryCacheEntry struct {
	value     string
	expiresAt time.Time
}

var errCacheMiss = errors.New("cache miss")

func (c *memoryCache) Get(ctx context.Context, key string) (string, error) {
	entry, ok := c.values[key]
	if !ok {
		return "", errCacheMiss
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		delete(c.values, key)
		return "", errCacheMiss
	}

	return entry.value, nil
}

func (c *memoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if c.values == nil {
		c.values = map[string]memoryCacheEntry{}
	}

	entry := memoryCacheEntry{value: value}
	if ttl > 0 {
		entry.expiresAt = time.Now().Add(ttl)
	}
	c.values[key] = entry
	return nil
}

func (c *memoryCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(c.values, key)
	}
	return nil
}

func TestSessionCacheKey(t *testing.T) {
	if got := sessionCacheKey("abc"); got != "auth:session:active:abc" {
		t.Fatalf("unexpected session cache key %q", got)
	}

	if got := refreshSessionCacheKey("hash"); got != "auth:refresh:session:hash" {
		t.Fatalf("unexpected refresh cache key %q", got)
	}

	if got := sessionRefreshHashCacheKey("session"); got != "auth:session:refresh_hash:session" {
		t.Fatalf("unexpected refresh hash index key %q", got)
	}
}

func TestCreateCachesSessionAndRefreshToken(t *testing.T) {
	repo := &SessionRepository{
		next:      &fakeSessionRepository{},
		cache:     &memoryCache{},
		activeTTL: time.Minute,
	}

	session := query.Session{
		SessionID:            "session-1",
		UserID:               42,
		Username:             "admin",
		Subject:              "42",
		RefreshTokenHash:     "refresh-hash-1",
		RefreshExpiresAtUnix: time.Now().Add(time.Hour).Unix(),
	}

	if err := repo.Create(t.Context(), session); err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	active, err := repo.cache.Get(t.Context(), sessionCacheKey(session.SessionID))
	if err != nil || active != sessionActiveCacheValue {
		t.Fatalf("expected active session cache to be primed, got value=%q err=%v", active, err)
	}

	refreshPayload, err := repo.cache.Get(t.Context(), refreshSessionCacheKey(session.RefreshTokenHash))
	if err != nil {
		t.Fatalf("expected refresh cache to be primed: %v", err)
	}

	cachedSession, err := decodeCachedSession(refreshPayload)
	if err != nil {
		t.Fatalf("decode cached session: %v", err)
	}

	if cachedSession.SessionID != session.SessionID {
		t.Fatalf("expected cached session id %q, got %q", session.SessionID, cachedSession.SessionID)
	}
}

func TestIsSessionActiveUsesCacheBeforeRepository(t *testing.T) {
	next := &fakeSessionRepository{active: map[string]bool{"session-1": true}}
	cache := &memoryCache{values: map[string]memoryCacheEntry{
		sessionCacheKey("session-1"): {
			value:     sessionActiveCacheValue,
			expiresAt: time.Now().Add(time.Minute),
		},
	}}
	repo := &SessionRepository{
		next:      next,
		cache:     cache,
		activeTTL: time.Minute,
	}

	active, err := repo.IsSessionActive(t.Context(), "session-1")
	if err != nil {
		t.Fatalf("unexpected active lookup error: %v", err)
	}
	if !active {
		t.Fatal("expected session to be active")
	}
	if next.activeLookups != 0 {
		t.Fatalf("expected repository lookup to be skipped, got %d", next.activeLookups)
	}
}

func TestFindActiveByRefreshTokenHashBackfillsCache(t *testing.T) {
	session := query.Session{
		SessionID:            "session-1",
		UserID:               42,
		Username:             "admin",
		Subject:              "42",
		RefreshTokenHash:     "refresh-hash-1",
		RefreshExpiresAtUnix: time.Now().Add(time.Hour).Unix(),
	}
	next := &fakeSessionRepository{
		active:          map[string]bool{"session-1": true},
		refreshSessions: map[string]query.Session{"refresh-hash-1": session},
	}
	repo := &SessionRepository{
		next:      next,
		cache:     &memoryCache{},
		activeTTL: time.Minute,
	}

	first, err := repo.FindActiveByRefreshTokenHash(t.Context(), "refresh-hash-1")
	if err != nil {
		t.Fatalf("unexpected refresh lookup error: %v", err)
	}
	if first.SessionID != session.SessionID {
		t.Fatalf("expected session id %q, got %q", session.SessionID, first.SessionID)
	}
	if next.refreshLookups != 1 {
		t.Fatalf("expected exactly one repository refresh lookup, got %d", next.refreshLookups)
	}

	second, err := repo.FindActiveByRefreshTokenHash(t.Context(), "refresh-hash-1")
	if err != nil {
		t.Fatalf("unexpected cached refresh lookup error: %v", err)
	}
	if second.SessionID != session.SessionID {
		t.Fatalf("expected cached session id %q, got %q", session.SessionID, second.SessionID)
	}
	if next.refreshLookups != 1 {
		t.Fatalf("expected cached refresh lookup to skip repository, got %d total lookups", next.refreshLookups)
	}
}

func TestRotateRefreshTokenReplacesRefreshCache(t *testing.T) {
	next := &fakeSessionRepository{
		active: map[string]bool{"session-1": true},
		refreshSessions: map[string]query.Session{
			"refresh-hash-1": {
				SessionID:            "session-1",
				UserID:               42,
				Username:             "admin",
				Subject:              "42",
				RefreshTokenHash:     "refresh-hash-1",
				RefreshExpiresAtUnix: time.Now().Add(time.Hour).Unix(),
			},
		},
	}
	cache := &memoryCache{}
	repo := &SessionRepository{
		next:      next,
		cache:     cache,
		activeTTL: time.Minute,
	}

	current := next.refreshSessions["refresh-hash-1"]
	repo.cacheRefreshSession(t.Context(), current)

	newExpiry := time.Now().Add(2 * time.Hour).Unix()
	if err := repo.RotateRefreshToken(t.Context(), current, "refresh-hash-2", newExpiry); err != nil {
		t.Fatalf("unexpected rotate error: %v", err)
	}

	if _, err := repo.cache.Get(t.Context(), refreshSessionCacheKey("refresh-hash-1")); err == nil {
		t.Fatal("expected old refresh cache entry to be deleted")
	}

	refreshPayload, err := repo.cache.Get(t.Context(), refreshSessionCacheKey("refresh-hash-2"))
	if err != nil {
		t.Fatalf("expected new refresh cache entry to exist: %v", err)
	}

	rotated, err := decodeCachedSession(refreshPayload)
	if err != nil {
		t.Fatalf("decode rotated session: %v", err)
	}
	if rotated.RefreshTokenHash != "refresh-hash-2" {
		t.Fatalf("expected rotated refresh hash to be updated, got %q", rotated.RefreshTokenHash)
	}
}

func TestRevokeBySessionIDMarksSessionInactiveAndInvalidatesRefreshCache(t *testing.T) {
	session := query.Session{
		SessionID:            "session-1",
		UserID:               42,
		Username:             "admin",
		Subject:              "42",
		RefreshTokenHash:     "refresh-hash-1",
		RefreshExpiresAtUnix: time.Now().Add(time.Hour).Unix(),
	}
	repo := &SessionRepository{
		next:      &fakeSessionRepository{active: map[string]bool{"session-1": true}},
		cache:     &memoryCache{},
		activeTTL: time.Minute,
	}
	repo.cacheSessionActive(t.Context(), session.SessionID, true)
	repo.cacheRefreshSession(t.Context(), session)

	if err := repo.RevokeBySessionID(t.Context(), session.SessionID); err != nil {
		t.Fatalf("unexpected revoke error: %v", err)
	}

	active, err := repo.cache.Get(t.Context(), sessionCacheKey(session.SessionID))
	if err != nil || active != sessionInactiveCacheValue {
		t.Fatalf("expected inactive session cache after revoke, got value=%q err=%v", active, err)
	}

	if _, err := repo.cache.Get(t.Context(), refreshSessionCacheKey(session.RefreshTokenHash)); err == nil {
		t.Fatal("expected refresh cache to be invalidated on revoke")
	}

	if _, err := repo.cache.Get(t.Context(), sessionRefreshHashCacheKey(session.SessionID)); err == nil {
		t.Fatal("expected refresh hash index to be invalidated on revoke")
	}
}

var _ repository.SessionRepository = (*fakeSessionRepository)(nil)
