package nats

import (
	"context"
	"errors"
	"testing"
	"time"

	"be/internal/alerting"
)

func TestAlertPublisherDoesNotWaitForNATSAtStartup(t *testing.T) {
	startedAt := time.Now()
	publisher := NewAlertPublisher(
		"nats://127.0.0.1:1",
		"TEST_ALERTS",
		"test.alerts.error",
		nil,
	)
	if elapsed := time.Since(startedAt); elapsed > 500*time.Millisecond {
		t.Fatalf("publisher startup blocked on NATS for %s", elapsed)
	}

	notifyStartedAt := time.Now()
	if err := publisher.Notify(context.Background(), alerting.Event{SchemaVersion: alerting.SchemaVersion}); err != nil {
		t.Fatalf("enqueue alert: %v", err)
	}
	if elapsed := time.Since(notifyStartedAt); elapsed > 500*time.Millisecond {
		t.Fatalf("enqueue blocked on NATS for %s", elapsed)
	}

	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := publisher.Close(closeCtx); err != nil {
		t.Fatalf("close publisher: %v", err)
	}
	if err := publisher.Notify(context.Background(), alerting.Event{}); !errors.Is(err, ErrAlertPublisherClosed) {
		t.Fatalf("expected ErrAlertPublisherClosed, got %v", err)
	}
}

func TestAlertPublisherDropsWhenQueueIsFull(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	publisher := &AlertPublisher{
		queue:  make(chan alerting.Event, 1),
		done:   make(chan struct{}),
		cancel: cancel,
	}
	defer cancel()

	if err := publisher.Notify(context.Background(), alerting.Event{}); err != nil {
		t.Fatalf("enqueue first alert: %v", err)
	}
	if err := publisher.Notify(context.Background(), alerting.Event{}); !errors.Is(err, ErrAlertQueueFull) {
		t.Fatalf("expected ErrAlertQueueFull, got %v", err)
	}
}
