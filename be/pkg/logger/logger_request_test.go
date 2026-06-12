package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"falzo-be/internal/share"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestCaptureRequestBodyRestoresBody(t *testing.T) {
	const original = `{"email":"admin@example.com","password":"top-secret"}`

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(original))
	captured := captureRequestBody(req, 2048)
	if captured.Truncated {
		t.Fatal("expected non-truncated request body")
	}

	payload, ok := captured.Payload.(map[string]any)
	if !ok {
		t.Fatalf("expected json payload to be parsed, got %#v", captured.Payload)
	}
	if payload["email"] != "admin@example.com" {
		t.Fatalf("expected email to be parsed, got %#v", payload["email"])
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read replayed request body: %v", err)
	}
	if string(body) != original {
		t.Fatalf("expected replayed request body %q, got %q", original, string(body))
	}
}

func TestCaptureRequestBodyMarksTruncated(t *testing.T) {
	original := `{"payload":"` + strings.Repeat("a", 128) + `"}`

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(original))
	captured := captureRequestBody(req, 16)
	if !captured.Truncated {
		t.Fatal("expected request body to be marked truncated")
	}
	if captured.Payload != nil {
		t.Fatalf("expected truncated payload to be skipped, got %#v", captured.Payload)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read replayed request body: %v", err)
	}
	if string(body) != original {
		t.Fatalf("expected replayed request body %q, got %q", original, string(body))
	}
}

func TestRequestLoggerLogsSanitizedJSONBodies(t *testing.T) {
	t.Setenv("LOG_HTTP_BODY_ENABLED", "true")
	t.Setenv("LOG_HTTP_BODY_MAX_BYTES", "4096")

	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})

	log.Logger = zerolog.New(newSensitiveDataWriter(&output, nil)).With().Timestamp().Logger()

	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read downstream request body: %v", err)
		}
		if got := string(data); got != `{"username":"alice","password":"secret"}` {
			t.Fatalf("expected downstream request body to match original, got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"token":"raw-token","ok":true}`))
	}))

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"username":"alice","password":"secret"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		t.Fatal("expected request log output")
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("failed to parse log entry json: %v", err)
	}

	requestBody, ok := entry["request_body"].(map[string]any)
	if !ok {
		t.Fatalf("expected request_body object, got %#v", entry["request_body"])
	}
	if requestBody["password"] != redactedLogValue {
		t.Fatalf("expected request password to be redacted, got %#v", requestBody["password"])
	}

	responseBody, ok := entry["response_body"].(map[string]any)
	if !ok {
		t.Fatalf("expected response_body object, got %#v", entry["response_body"])
	}
	if responseBody["token"] != redactedLogValue {
		t.Fatalf("expected response token to be redacted, got %#v", responseBody["token"])
	}
}

func TestRequestLoggerSkipsSuccessfulRequests(t *testing.T) {
	t.Setenv("LOG_HTTP_BODY_ENABLED", "false")

	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})

	log.Logger = zerolog.New(&output)
	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello world!"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := strings.TrimSpace(output.String()); got != "" {
		t.Fatalf("expected no log output for successful request, got %q", got)
	}
}

func TestRequestLoggerSkipsSuccessfulStreams(t *testing.T) {
	t.Setenv("LOG_HTTP_BODY_ENABLED", "false")

	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})

	log.Logger = zerolog.New(&output)
	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(": connected\n\n"))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/posts/events", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := strings.TrimSpace(output.String()); got != "" {
		t.Fatalf("expected no log output for successful stream, got %q", got)
	}
}

func TestStatusRecorderForwardsFlush(t *testing.T) {
	rec := httptest.NewRecorder()
	status := &statusRecorder{ResponseWriter: rec}

	flusher, ok := any(status).(http.Flusher)
	if !ok {
		t.Fatal("expected status recorder to implement http.Flusher")
	}

	flusher.Flush()

	if !rec.Flushed {
		t.Fatal("expected flush to be forwarded to underlying response writer")
	}
	if status.status != http.StatusOK {
		t.Fatalf("expected flush to mark status OK, got %d", status.status)
	}
	if status.Unwrap() != rec {
		t.Fatal("expected unwrap to return underlying response writer")
	}
}

func TestRequestLoggerUsesErrorLevelForServerErrors(t *testing.T) {
	t.Setenv("LOG_HTTP_BODY_ENABLED", "false")

	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})

	log.Logger = zerolog.New(&output)
	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/images/upload", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry); err != nil {
		t.Fatalf("failed to parse log entry json: %v", err)
	}

	if entry["level"] != "error" {
		t.Fatalf("expected error level for 5xx response, got %#v", entry["level"])
	}
	if got := int(entry["status"].(float64)); got != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, got)
	}
}

func TestRequestLoggerSkipsErrorsAlreadyLoggedByWriteError(t *testing.T) {
	t.Setenv("LOG_HTTP_BODY_ENABLED", "false")

	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})

	log.Logger = zerolog.New(&output)
	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		share.WriteError(w, r, errors.New("database unavailable"), "load_profile", func(error) share.ApiError {
			return share.Internal()
		})
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/users/42", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected exactly one error log, got %d logs: %q", len(lines), output.String())
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("failed to parse log entry json: %v", err)
	}
	if entry["message"] != "request failed" {
		t.Fatalf("expected request failed log, got %#v", entry["message"])
	}
}
