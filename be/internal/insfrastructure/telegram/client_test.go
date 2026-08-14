package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"be/internal/alerting"
)

func TestClientSendsFormattedAlert(t *testing.T) {
	var received struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/bottest-token/sendMessage") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test-token", "-100123", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	event := alerting.Event{
		ID: "event-1", OccurredAt: time.Date(2026, 8, 14, 1, 2, 3, 0, time.UTC),
		Service: "falzo-api", Environment: "test", Message: "phase failed",
		Fields: map[string]any{"room_id": "room-1"},
	}
	if err := client.SendAlert(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if received.ChatID != "-100123" || !strings.Contains(received.Text, "phase failed") || !strings.Contains(received.Text, "room_id: room-1") {
		t.Fatalf("unexpected Telegram payload: %#v", received)
	}
}

func TestFormatAlertHonorsTelegramLimit(t *testing.T) {
	event := alerting.Event{Message: strings.Repeat("x", 5_000)}
	if got := len([]rune(FormatAlert(event))); got > telegramMessageLimit {
		t.Fatalf("message length %d exceeds Telegram limit", got)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestClientRedactsBotTokenFromTransportErrors(t *testing.T) {
	const token = "123456:secret-token"
	client, err := NewClient("https://api.telegram.org", token, "123", &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return nil, errors.New(request.URL.String())
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = client.SendAlert(context.Background(), alerting.Event{Message: "test"})
	if err == nil || strings.Contains(err.Error(), token) || !strings.Contains(err.Error(), "[REDACTED]") {
		t.Fatalf("token was not redacted: %v", err)
	}
}
