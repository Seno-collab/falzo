package alerting

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"

	"github.com/google/uuid"
)

const (
	maxAlertFields   = 20
	maxFieldLength   = 1_000
	maxMessageLength = 500
)

var sensitiveValuePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(password|passwd|token|secret|authorization|cookie)(\s*[:=]\s*)([^\s,;]+)`),
	regexp.MustCompile(`(?i)([a-z][a-z0-9+.-]*://[^:/\s]+:)([^@\s]+)(@)`),
	regexp.MustCompile(`(?i)(bearer\s+)([a-z0-9._~+/=-]+)`),
}

type SlogHandler struct {
	base        slog.Handler
	notifier    Notifier
	service     string
	environment string
	attrs       []slog.Attr
	groups      []string
}

func NewSlogHandler(base slog.Handler, notifier Notifier, service, environment string) *SlogHandler {
	return &SlogHandler{base: base, notifier: notifier, service: service, environment: environment}
}

func (h *SlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *SlogHandler) Handle(ctx context.Context, record slog.Record) error {
	if err := h.base.Handle(ctx, record); err != nil {
		return err
	}
	if record.Level < slog.LevelError || h.notifier == nil {
		return nil
	}

	event := Event{
		SchemaVersion: SchemaVersion,
		ID:            uuid.NewString(),
		OccurredAt:    record.Time.UTC(),
		Severity:      "error",
		Service:       h.service,
		Environment:   h.environment,
		Message:       truncate(sanitizeText(record.Message), maxMessageLength),
		Source:        sourceFor(record.PC),
		Fields:        make(map[string]any),
	}
	for _, attr := range h.attrs {
		appendSafeAttr(event.Fields, h.groups, attr)
	}
	record.Attrs(func(attr slog.Attr) bool {
		if len(event.Fields) >= maxAlertFields {
			return false
		}
		appendSafeAttr(event.Fields, h.groups, attr)
		return true
	})
	if len(event.Fields) == 0 {
		event.Fields = nil
	}
	_ = h.notifier.Notify(ctx, event)
	return nil
}

func (h *SlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.base = h.base.WithAttrs(attrs)
	clone.attrs = append(slices.Clone(h.attrs), attrs...)
	clone.groups = slices.Clone(h.groups)
	return &clone
}

func (h *SlogHandler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.base = h.base.WithGroup(name)
	clone.attrs = slices.Clone(h.attrs)
	clone.groups = append(slices.Clone(h.groups), name)
	return &clone
}

func appendSafeAttr(fields map[string]any, groups []string, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}
	key := strings.Join(append(slices.Clone(groups), attr.Key), ".")
	if sensitiveKey(key) {
		return
	}
	if attr.Value.Kind() == slog.KindGroup {
		for _, child := range attr.Value.Group() {
			if len(fields) >= maxAlertFields {
				return
			}
			appendSafeAttr(fields, append(slices.Clone(groups), attr.Key), child)
		}
		return
	}
	if len(fields) >= maxAlertFields {
		return
	}
	fields[key] = safeValue(attr.Value)
}

func safeValue(value slog.Value) any {
	switch value.Kind() {
	case slog.KindBool:
		return value.Bool()
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindFloat64:
		return value.Float64()
	case slog.KindInt64:
		return value.Int64()
	case slog.KindString:
		return truncate(sanitizeText(value.String()), maxFieldLength)
	case slog.KindTime:
		return value.Time().UTC().Format("2006-01-02T15:04:05.000Z07:00")
	case slog.KindUint64:
		return value.Uint64()
	default:
		return truncate(sanitizeText(fmt.Sprint(value.Any())), maxFieldLength)
	}
}

func sanitizeText(value string) string {
	value = sensitiveValuePatterns[0].ReplaceAllString(value, `${1}${2}[REDACTED]`)
	value = sensitiveValuePatterns[1].ReplaceAllString(value, `${1}[REDACTED]${3}`)
	return sensitiveValuePatterns[2].ReplaceAllString(value, `${1}[REDACTED]`)
}

func sensitiveKey(key string) bool {
	key = strings.ToLower(key)
	for _, fragment := range []string{"password", "secret", "token", "authorization", "cookie", "body", "credential"} {
		if strings.Contains(key, fragment) {
			return true
		}
	}
	return false
}

func sourceFor(pc uintptr) string {
	if pc == 0 {
		return ""
	}
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	if frame.File == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit-1]) + "…"
}
