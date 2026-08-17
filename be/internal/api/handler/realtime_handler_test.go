package handler

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/coder/websocket"
)

func TestRealtimeDisconnectLogAttrsNormalClosure(t *testing.T) {
	err := fmt.Errorf("failed to read JSON message: %w", websocket.CloseError{
		Code:   websocket.StatusNormalClosure,
		Reason: "Room closed",
	})

	level, attrs := realtimeDisconnectLogAttrs("room-1", 1, err)
	if level != slog.LevelInfo {
		t.Fatalf("level = %v, want INFO", level)
	}

	values := attrValues(attrs)
	if _, ok := values["error"]; ok {
		t.Fatal("normal closure must not be logged as an error")
	}
	if got := values["close_status"].Int64(); got != int64(websocket.StatusNormalClosure) {
		t.Fatalf("close_status = %d, want %d", got, websocket.StatusNormalClosure)
	}
	if got := values["close_reason"].String(); got != "Room closed" {
		t.Fatalf("close_reason = %q, want %q", got, "Room closed")
	}
	if got := values["disconnect_reason"].String(); got != "normal" {
		t.Fatalf("disconnect_reason = %q, want normal", got)
	}
}

func TestRealtimeDisconnectLogAttrsTransportError(t *testing.T) {
	level, attrs := realtimeDisconnectLogAttrs("room-1", 1, fmt.Errorf("network unavailable"))
	if level != slog.LevelWarn {
		t.Fatalf("level = %v, want WARN", level)
	}
	if _, ok := attrValues(attrs)["error"]; !ok {
		t.Fatal("transport failure must include the error")
	}
}

func attrValues(attrs []slog.Attr) map[string]slog.Value {
	values := make(map[string]slog.Value, len(attrs))
	for _, attr := range attrs {
		values[attr.Key] = attr.Value
	}
	return values
}
