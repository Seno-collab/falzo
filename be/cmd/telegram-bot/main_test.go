package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	for _, path := range []string{"/healthz", "/health/live", "/health/ready"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()

		newHealthHandler().ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("%s returned status %d", path, response.Code)
		}
		if !strings.Contains(response.Body.String(), `"service":"falzo-telegram-bot"`) {
			t.Fatalf("%s returned unexpected body %q", path, response.Body.String())
		}
	}
}

func TestHealthHandlerRejectsUnsupportedMethod(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	response := httptest.NewRecorder()
	newHealthHandler().ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, response.Code)
	}
}
