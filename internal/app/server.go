package app

import (
	"context"
	"falzo/pkg/lib"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func Run() {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(requestLogger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})
	sm := lib.NewShutdownManager(60 * time.Second)
	srv := &http.Server{Addr: ":8080", Handler: r}
	sm.Register("http-stop-acceting", 5*time.Second, func(ctx context.Context) error {
		// Stop accepting new requests immediately
		return srv.Shutdown(ctx)
	})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Str("addr", srv.Addr).Msg("server failed")
			os.Exit(1)
		}
	}()
	log.Info().Str("addr", srv.Addr).Msg("server started")
	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("shutdown initiated")

	if err := sm.Shutdown(); err != nil {
		log.Error().Err(err).Msg("shutdown completed with errors")
		os.Exit(1)
	}

	log.Info().Msg("shutdown complete")
}

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

func requestLogger(next http.Handler) http.Handler {
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
