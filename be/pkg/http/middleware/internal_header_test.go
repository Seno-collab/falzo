package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInternalHeaderAllowsMatchingHeader(t *testing.T) {
	mw := InternalHeader(InternalHeaderConfig{
		Name:  "X-Test-Key",
		Value: "expected",
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("X-Test-Key", "expected")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestInternalHeaderRejectsMissingHeader(t *testing.T) {
	mw := InternalHeader(InternalHeaderConfig{
		Name:  "X-Test-Key",
		Value: "expected",
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("request should not reach next handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestInternalHeaderDisabledWhenValueEmpty(t *testing.T) {
	mw := InternalHeader(InternalHeaderConfig{
		Name: "X-Test-Key",
	})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
