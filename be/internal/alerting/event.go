package alerting

import (
	"context"
	"time"
)

const SchemaVersion = 1

type Event struct {
	SchemaVersion int            `json:"schema_version"`
	ID            string         `json:"id"`
	OccurredAt    time.Time      `json:"occurred_at"`
	Severity      string         `json:"severity"`
	Service       string         `json:"service"`
	Environment   string         `json:"environment"`
	Message       string         `json:"message"`
	Source        string         `json:"source,omitempty"`
	Fields        map[string]any `json:"fields,omitempty"`
}

type Notifier interface {
	Notify(context.Context, Event) error
}
