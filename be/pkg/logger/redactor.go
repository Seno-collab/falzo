package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

const redactedLogValue = "[REDACTED]"

var defaultSensitiveKeyMarkers = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"apikey",
	"authorization",
	"cookie",
	"credential",
	"privatekey",
	"clientsecret",
}

type sensitiveDataWriter struct {
	out        io.Writer
	keyMarkers []string
}

func newSensitiveDataWriter(out io.Writer, keyMarkers []string) io.Writer {
	if len(keyMarkers) == 0 {
		keyMarkers = defaultSensitiveKeyMarkers
	}

	return sensitiveDataWriter{
		out:        out,
		keyMarkers: keyMarkers,
	}
}

func (w sensitiveDataWriter) Write(p []byte) (int, error) {
	sanitized := sanitizeLogPayloadWithKeyMarkers(p, w.keyMarkers)

	n, err := w.out.Write(sanitized)
	if err != nil {
		return n, err
	}

	return len(p), nil
}

func sanitizeLogPayload(payload []byte) []byte {
	return sanitizeLogPayloadWithKeyMarkers(payload, defaultSensitiveKeyMarkers)
}

func sanitizeLogPayloadWithKeyMarkers(payload []byte, keyMarkers []string) []byte {
	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 {
		return payload
	}

	var parsed any
	if err := json.Unmarshal(trimmed, &parsed); err != nil {
		return payload
	}

	sanitized := sanitizeLogValue("", parsed, keyMarkers)
	encoded, err := json.Marshal(sanitized)
	if err != nil {
		return payload
	}

	if bytes.HasSuffix(payload, []byte("\n")) {
		encoded = append(encoded, '\n')
	}

	return encoded
}

func sanitizeLogValue(key string, value any, keyMarkers []string) any {
	if isSensitiveLogKey(key, keyMarkers) {
		return redactedLogValue
	}

	switch v := value.(type) {
	case map[string]any:
		masked := make(map[string]any, len(v))
		for field, fieldValue := range v {
			masked[field] = sanitizeLogValue(field, fieldValue, keyMarkers)
		}

		return masked
	case []any:
		masked := make([]any, len(v))
		for idx, item := range v {
			masked[idx] = sanitizeLogValue("", item, keyMarkers)
		}

		return masked
	default:
		return value
	}
}

func isSensitiveLogKey(key string, keyMarkers []string) bool {
	if key == "" {
		return false
	}

	normalized := normalizeLogKey(key)
	if normalized == "" {
		return false
	}

	for _, marker := range keyMarkers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}

	return false
}

func normalizeLogKey(key string) string {
	lowered := strings.ToLower(key)
	var b strings.Builder
	b.Grow(len(lowered))

	for _, char := range lowered {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			b.WriteRune(char)
		}
	}

	return b.String()
}

func loadSensitiveKeyMarkers(raw string) []string {
	markers := make([]string, 0, len(defaultSensitiveKeyMarkers))
	seen := make(map[string]struct{}, len(defaultSensitiveKeyMarkers))

	for _, marker := range defaultSensitiveKeyMarkers {
		normalized := normalizeLogKey(marker)
		if normalized == "" {
			continue
		}

		if _, exists := seen[normalized]; exists {
			continue
		}

		seen[normalized] = struct{}{}
		markers = append(markers, normalized)
	}

	for part := range strings.SplitSeq(raw, ",") {
		normalized := normalizeLogKey(strings.TrimSpace(part))
		if normalized == "" {
			continue
		}

		if _, exists := seen[normalized]; exists {
			continue
		}

		seen[normalized] = struct{}{}
		markers = append(markers, normalized)
	}

	return markers
}
