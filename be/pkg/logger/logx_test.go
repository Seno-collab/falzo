package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestLoggerAddsCategoryScopeAndRequestID(t *testing.T) {
	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})

	log.Logger = zerolog.New(&output)
	ctx := context.WithValue(context.Background(), chimiddleware.RequestIDKey, "host/request-123")

	For("post.handler").
		With(Str("operation", "create_post")).
		Warn(ctx, errors.New("publish failed"), "post event publish failed", Uint64("post_id", 42))

	var entry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(output.String())), &entry); err != nil {
		t.Fatalf("failed to parse log entry json: %v", err)
	}

	if entry["level"] != "warn" {
		t.Fatalf("expected warn level, got %#v", entry["level"])
	}
	if entry["category"] != "post.handler" {
		t.Fatalf("expected category post.handler, got %#v", entry["category"])
	}
	if entry["operation"] != "create_post" {
		t.Fatalf("expected operation create_post, got %#v", entry["operation"])
	}
	if entry["request_id"] != "request-123" {
		t.Fatalf("expected sanitized request id, got %#v", entry["request_id"])
	}
	if got := uint64(entry["post_id"].(float64)); got != 42 {
		t.Fatalf("expected post_id 42, got %d", got)
	}
	if entry["error"] != "publish failed" {
		t.Fatalf("expected error field, got %#v", entry["error"])
	}
}
