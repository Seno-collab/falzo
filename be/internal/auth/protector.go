package auth

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type authProtector struct {
	mu               sync.Mutex
	failureCount     int
	failureThreshold int
	openUntil        time.Time
	cooldown         time.Duration
	loginLimiter     *keyLimiter
	refreshLimiter   *keyLimiter
	registerLimiter  *keyLimiter
}

func newAuthProtector(limitPerMinute int, failureThreshold int, cooldown time.Duration) *authProtector {
	return &authProtector{
		failureThreshold: failureThreshold,
		cooldown:         cooldown,
		loginLimiter:     newKeyLimiter(limitPerMinute, time.Minute),
		refreshLimiter:   newKeyLimiter(limitPerMinute, time.Minute),
		registerLimiter:  newKeyLimiter(limitPerMinute, time.Minute),
	}
}

func (p *authProtector) allowOperation(now time.Time) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return now.After(p.openUntil)
}

func (p *authProtector) observe(err error, now time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err == nil {
		p.failureCount = 0
		return
	}

	if !errors.Is(err, ErrDependencyUnavailable) && !errors.Is(err, ErrInternal) {
		return
	}

	p.failureCount++
	if p.failureCount >= p.failureThreshold {
		p.openUntil = now.Add(p.cooldown)
		p.failureCount = 0
	}
}

type keyLimiter struct {
	mu              sync.Mutex
	limit           int
	window          time.Duration
	items           map[string]windowCounter
	maxEntries      int
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

type windowCounter struct {
	count     int
	expiresAt time.Time
}

func newKeyLimiter(limit int, window time.Duration) *keyLimiter {
	return &keyLimiter{
		limit:           limit,
		window:          window,
		items:           make(map[string]windowCounter),
		maxEntries:      10000,
		cleanupInterval: time.Minute,
	}
}

func (l *keyLimiter) allow(key string, now time.Time) bool {
	if l == nil || l.limit <= 0 {
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.cleanupExpiredLocked(now)

	entry, ok := l.items[key]
	if !ok || now.After(entry.expiresAt) {
		if !ok && len(l.items) >= l.maxEntries {
			return false
		}

		l.items[key] = windowCounter{
			count:     1,
			expiresAt: now.Add(l.window),
		}
		return true
	}

	if entry.count >= l.limit {
		return false
	}

	entry.count++
	l.items[key] = entry
	return true
}

func authClientIP(r *http.Request) string {
	host := strings.TrimSpace(r.RemoteAddr)

	if ip, _, err := net.SplitHostPort(host); err == nil {
		return ip
	}

	return host
}

func (l *keyLimiter) cleanupExpiredLocked(now time.Time) {
	if l.cleanupInterval <= 0 || l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= l.cleanupInterval {
		for key, item := range l.items {
			if now.After(item.expiresAt) {
				delete(l.items, key)
			}
		}
		l.lastCleanup = now
	}
}
