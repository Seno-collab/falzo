package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// RequestLogger records one structured event after each HTTP request.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := time.Now()
			writer := &statusWriter{ResponseWriter: w}
			requestBody, requestLMID := readRequestBody(r)

			requestAttrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			}
			if requestLMID != "" {
				requestAttrs = append(requestAttrs, slog.String("lmid", requestLMID))
			}
			if requestBody != "" {
				requestAttrs = append(requestAttrs, slog.String("body", requestBody))
			}
			appendRequestIDAndRoute(r, &requestAttrs)
			logger.LogAttrs(r.Context(), slog.LevelInfo, "http request", requestAttrs...)

			next.ServeHTTP(writer, r)

			status := writer.status
			if status == 0 {
				status = http.StatusOK
			}

			level := slog.LevelInfo
			switch {
			case status >= http.StatusInternalServerError:
				level = slog.LevelError
			case status >= http.StatusBadRequest:
				level = slog.LevelWarn
			}

			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Int("bytes", writer.bytes),
				slog.Duration("duration", time.Since(startedAt)),
			}
			responseBody := sanitizeJSON(writer.body.Bytes())
			responseLMID := extractLMID(writer.body.Bytes())
			if responseLMID == "" {
				responseLMID = requestLMID
			}
			if responseLMID != "" {
				attrs = append(attrs, slog.String("lmid", responseLMID))
			}
			if responseBody != "" {
				attrs = append(attrs, slog.String("body", responseBody))
			}
			appendRequestIDAndRoute(r, &attrs)
			logger.LogAttrs(r.Context(), level, "http response", attrs...)
		})
	}
}

const maxLoggedBodySize = 64 * 1024

func readRequestBody(r *http.Request) (body, lmid string) {
	if r.Body == nil || r.ContentLength == 0 {
		return "", ""
	}
	if !strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		return "[body omitted]", ""
	}

	rawBody, err := io.ReadAll(io.LimitReader(r.Body, maxLoggedBodySize+1))
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(rawBody))
		return "[body unavailable]", ""
	}
	r.Body = io.NopCloser(bytes.NewReader(rawBody))
	if len(rawBody) > maxLoggedBodySize {
		return "[body truncated]", ""
	}
	return sanitizeJSON(rawBody), extractLMID(rawBody)
}

func sanitizeJSON(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return "[invalid-json]"
	}
	redactJSON(value)
	encoded, err := json.Marshal(value)
	if err != nil {
		return "[body unavailable]"
	}
	return string(encoded)
}

func redactJSON(value any) {
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			if isSensitiveField(key) {
				current[key] = "[REDACTED]"
				continue
			}
			redactJSON(child)
		}
	case []any:
		for _, child := range current {
			redactJSON(child)
		}
	}
}

func isSensitiveField(key string) bool {
	normalized := normalizeJSONField(key)
	switch normalized {
	case "password", "newpassword", "token", "accesstoken", "refreshtoken", "resettoken", "credential", "idtoken", "authorization", "cookie", "secret":
		return true
	default:
		return false
	}
}

func extractLMID(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var value any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return ""
	}

	lmid, _ := findJSONField(value, "lmid")
	return lmid
}

func findJSONField(value any, target string) (string, bool) {
	switch current := value.(type) {
	case map[string]any:
		for key, child := range current {
			if normalizeJSONField(key) == target {
				switch scalar := child.(type) {
				case string:
					trimmed := strings.TrimSpace(scalar)
					return trimmed, trimmed != ""
				case json.Number:
					return scalar.String(), true
				}
			}
		}
		for _, child := range current {
			if found, ok := findJSONField(child, target); ok {
				return found, true
			}
		}
	case []any:
		for _, child := range current {
			if found, ok := findJSONField(child, target); ok {
				return found, true
			}
		}
	}
	return "", false
}

func normalizeJSONField(key string) string {
	return strings.ToLower(strings.NewReplacer("-", "", "_", "", " ", "").Replace(key))
}

func appendRequestIDAndRoute(r *http.Request, attrs *[]slog.Attr) {
	if requestID := chimiddleware.GetReqID(r.Context()); requestID != "" {
		*attrs = append(*attrs, slog.String("request_id", requestID))
	}
	if routeContext := chi.RouteContext(r.Context()); routeContext != nil {
		if route := routeContext.RoutePattern(); route != "" {
			*attrs = append(*attrs, slog.String("route", route))
		}
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
	body   bytes.Buffer
}

func (w *statusWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	if w.body.Len() < maxLoggedBodySize {
		remaining := maxLoggedBodySize - w.body.Len()
		if len(body) > remaining {
			_, _ = w.body.Write(body[:remaining])
		} else {
			_, _ = w.body.Write(body)
		}
	}
	return n, err
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *statusWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	conn, readWriter, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		_ = conn.Close()
		return nil, nil, fmt.Errorf("clear hijacked connection deadline: %w", err)
	}
	return conn, readWriter, nil
}
