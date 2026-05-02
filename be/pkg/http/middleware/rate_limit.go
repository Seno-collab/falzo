package middleware

import (
	"errors"
	"falzo-be/internal/share"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var errRateLimited = errors.New("rate limited")

type fixedWindowLimiter struct {
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

type KeyFunc func(*http.Request) string

const (
	defaultLimiterMaxEntries      = 10000
	defaultLimiterCleanupInterval = time.Minute
)

func NewIPRateLimiter(limit int, window time.Duration) func(http.Handler) http.Handler {
	return NewKeyedRateLimiter(limit, window, clientIP)
}

func NewKeyedRateLimiter(limit int, window time.Duration, keyFunc KeyFunc) func(http.Handler) http.Handler {
	limiter := &fixedWindowLimiter{
		limit:           limit,
		window:          window,
		items:           make(map[string]windowCounter),
		maxEntries:      defaultLimiterMaxEntries,
		cleanupInterval: defaultLimiterCleanupInterval,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limit <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			key := ""
			if keyFunc != nil {
				key = strings.TrimSpace(keyFunc(r))
			}
			if key == "" {
				key = clientIP(r)
			}
			if !limiter.allow(key, time.Now()) {
				share.WriteError(w, r, errRateLimited, "rate_limit", mapRateLimitError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func mapRateLimitError(err error) share.ApiError {
	return share.TooManyRequests("Too many requests, please try again later")
}

func (l *fixedWindowLimiter) allow(key string, now time.Time) bool {
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

func clientIP(r *http.Request) string {
	host := strings.TrimSpace(r.RemoteAddr)

	if ip, _, err := net.SplitHostPort(host); err == nil {
		return ip
	}

	return host
}

func (l *fixedWindowLimiter) cleanupExpiredLocked(now time.Time) {
	if l.cleanupInterval <= 0 || (l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= l.cleanupInterval) {
		for key, item := range l.items {
			if now.After(item.expiresAt) {
				delete(l.items, key)
			}
		}
		l.lastCleanup = now
	}
}
