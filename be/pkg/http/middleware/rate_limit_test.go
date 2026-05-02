package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewKeyedRateLimiterLimitsByCustomKey(t *testing.T) {
	middleware := NewKeyedRateLimiter(2, time.Minute, func(r *http.Request) string {
		return r.Header.Get("X-User-ID")
	})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := range 2 {
		req := httptest.NewRequest(http.MethodPost, "/images/upload", nil)
		req.Header.Set("X-User-ID", "7")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d: expected status %d, got %d", i+1, http.StatusNoContent, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/images/upload", nil)
	req.Header.Set("X-User-ID", "7")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
}

func TestNewKeyedRateLimiterUsesSeparateBuckets(t *testing.T) {
	middleware := NewKeyedRateLimiter(1, time.Minute, func(r *http.Request) string {
		return r.Header.Get("X-User-ID")
	})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, userID := range []string{"7", "8"} {
		req := httptest.NewRequest(http.MethodPost, "/images/upload", nil)
		req.Header.Set("X-User-ID", userID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("user %s: expected status %d, got %d", userID, http.StatusNoContent, rec.Code)
		}
	}
}
