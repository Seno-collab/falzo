package lib

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(data)
	r.bytes += n
	return n, err
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		event := log.Info()
		if r.URL.Path == "/favicon.ico" {
			event = log.Debug()
		}

		durationMs := float64(time.Since(start).Microseconds()) / 1000
		event = event.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rec.status).
			Float64("duration_ms", durationMs).
			Int("bytes", rec.bytes).
			Str("remote_ip", r.RemoteAddr)

		if requestID := middleware.GetReqID(r.Context()); requestID != "" {
			event = event.Str("request_id", sanitizeRequestID(requestID))
		}

		if r.ContentLength >= 0 {
			event = event.Int64("content_length", r.ContentLength)
		}

		event.Msg("request completed")
	})
}

func sanitizeRequestID(requestID string) string {
	if _, suffix, found := strings.Cut(requestID, "/"); found {
		return suffix
	}

	return requestID
}
