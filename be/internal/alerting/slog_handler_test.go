package alerting

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
)

type recordingNotifier struct{ events []Event }

func (n *recordingNotifier) Notify(_ context.Context, event Event) error {
	n.events = append(n.events, event)
	return nil
}

func TestSlogHandlerPublishesOnlyErrorsAndRedactsSensitiveFields(t *testing.T) {
	notifier := &recordingNotifier{}
	base := slog.NewJSONHandler(io.Discard, nil)
	logger := slog.New(NewSlogHandler(base, notifier, "falzo-api", "test")).With("component", "room")

	logger.Warn("not an alert", "error", "temporary")
	logger.Error("phase transition failed",
		"room_id", "room-1",
		"error", "connect postgres://falzo:db-password@database/falzo failed; token=abc123",
		"password", "must-not-leak",
		"request_body", "must-not-leak",
	)

	if len(notifier.events) != 1 {
		t.Fatalf("expected one alert, got %d", len(notifier.events))
	}
	event := notifier.events[0]
	if event.Message != "phase transition failed" || event.Service != "falzo-api" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if event.Fields["component"] != "room" || event.Fields["room_id"] != "room-1" {
		t.Fatalf("missing safe fields: %#v", event.Fields)
	}
	if _, exists := event.Fields["password"]; exists {
		t.Fatal("password field was not redacted")
	}
	if _, exists := event.Fields["request_body"]; exists {
		t.Fatal("body field was not redacted")
	}
	errorText, _ := event.Fields["error"].(string)
	if strings.Contains(errorText, "db-password") || strings.Contains(errorText, "abc123") {
		t.Fatalf("sensitive values were not redacted: %s", errorText)
	}
}
