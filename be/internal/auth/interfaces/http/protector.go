package httpapi

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"falzo-be/internal/auth/domain"
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

	if !errors.Is(err, domain.ErrAuthDependencyUnavailable) && !errors.Is(err, domain.ErrAuthInternal) {
		return
	}

	p.failureCount++
	if p.failureCount >= p.failureThreshold {
		p.openUntil = now.Add(p.cooldown)
		p.failureCount = 0
	}
}

type keyLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	items  map[string]windowCounter
}

type windowCounter struct {
	count     int
	expiresAt time.Time
}

func newKeyLimiter(limit int, window time.Duration) *keyLimiter {
	return &keyLimiter{
		limit:  limit,
		window: window,
		items:  make(map[string]windowCounter),
	}
}

func (l *keyLimiter) allow(key string, now time.Time) bool {
	if l == nil || l.limit <= 0 {
		return true
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.items[key]
	if !ok || now.After(entry.expiresAt) {
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
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		host = strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}

	if ip, _, err := net.SplitHostPort(host); err == nil {
		return ip
	}

	return host
}
