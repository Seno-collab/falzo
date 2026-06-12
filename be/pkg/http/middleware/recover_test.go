package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestRecoverHandlesPanic(t *testing.T) {
	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})
	log.Logger = zerolog.New(&output)

	handler := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected one panic log, got %d logs: %q", len(lines), output.String())
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("failed to parse log entry json: %v", err)
	}
	if entry["message"] != "request failed" {
		t.Fatalf("expected request failed log, got %#v", entry["message"])
	}
	if entry["app_code"] != "REQUEST_PANIC" {
		t.Fatalf("expected request panic app code, got %#v", entry["app_code"])
	}
	metadata, ok := entry["app_metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected panic metadata, got %#v", entry["app_metadata"])
	}
	if metadata["panic"] != "boom" {
		t.Fatalf("expected panic value in metadata, got %#v", metadata["panic"])
	}
	if metadata["stack"] == "" {
		t.Fatalf("expected panic stack in metadata, got %#v", metadata["stack"])
	}
}
