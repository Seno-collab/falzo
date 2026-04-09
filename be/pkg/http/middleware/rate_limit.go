package middleware

import (
	httpResponse "falzo-be/pkg/response"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type fixedWindowLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	items  map[string]windowCounter
}

type windowCounter struct {
	count     int
	expiresAt time.Time
}

func NewIPRateLimiter(limit int, window time.Duration) func(http.Handler) http.Handler {
	limiter := &fixedWindowLimiter{
		limit:  limit,
		window: window,
		items:  make(map[string]windowCounter),
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limit <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			key := clientIP(r)
			if !limiter.allow(key, time.Now()) {
				httpResponse.Error(w, http.StatusTooManyRequests, "Too many requests", r, httpResponse.ErrorDetail{
					Code:    "RATE_LIMITED",
					Message: "Too many requests, please try again later",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (l *fixedWindowLimiter) allow(key string, now time.Time) bool {
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

func clientIP(r *http.Request) string {
	host := strings.TrimSpace(r.RemoteAddr)
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		host = strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}

	if ip, _, err := net.SplitHostPort(host); err == nil {
		return ip
	}

	return host
}
