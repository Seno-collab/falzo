package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	authservice "falzo/internal/auth/application"
	authmiddleware "falzo/internal/auth/infrastructure/http"
	authpostgres "falzo/internal/auth/infrastructure/persistence/postgres"
	authbcrypt "falzo/internal/auth/infrastructure/security/bcrypt"
	authtoken "falzo/internal/auth/infrastructure/token"
	authhandler "falzo/internal/auth/interfaces/http"
	"falzo/pkg/cache"
	"falzo/pkg/config"
	"falzo/pkg/database"
	"falzo/pkg/dto"
	httpmiddleware "falzo/pkg/http/middleware"
	httpresponse "falzo/pkg/response"
	"falzo/pkg/shutdown"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func Run() {
	cfg := config.Load()
	dbClient, err := database.New(cfg.Postgres)
	if err != nil {
		log.Error().Err(err).Msg("postgres unavailable at startup")
		os.Exit(1)
	}

	redisClient, err := cache.New(cfg.Redis)
	if err != nil {
		log.Warn().Err(err).Str("addr", cfg.Redis.Addr).Msg("redis unavailable at startup")
	}

	r := chi.NewRouter()
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.RequestID)
	r.Use(httpmiddleware.Recover)
	r.Use(requestLogger)

	jwtManager := authtoken.NewJWTManager(cfg.Auth)
	authService := authservice.New(
		authpostgres.NewAccountRepository(dbClient),
		authbcrypt.NewPasswordHasher(),
		jwtManager,
		jwtManager,
	)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		httpresponse.JSON(w, http.StatusOK, dto.MessageResponse{Message: "Hello world!"})
	})
	r.Mount("/auth", authhandler.New(authService).Routes())
	r.With(authmiddleware.RequireAuth(authService)).Get("/profile", func(w http.ResponseWriter, r *http.Request) {
		principal, ok := authmiddleware.AuthenticatedUserFromContext(r.Context())
		if !ok {
			httpresponse.Error(w, http.StatusUnauthorized, "unauthorized", "missing auth context")
			return
		}

		httpresponse.JSON(w, http.StatusOK, map[string]string{
			"message":  "authenticated",
			"username": principal.Username,
		})
	})

	sm := shutdown.NewManager()
	srv := &http.Server{Addr: cfg.HTTP.Addr, Handler: r}
	sm.Register("http-stop-accepting", 5*time.Second, func(ctx context.Context) error {
		// Stop accepting new requests immediately
		return srv.Shutdown(ctx)
	})
	if redisClient != nil {
		sm.Register("redis-close", 5*time.Second, func(ctx context.Context) error {
			return redisClient.Close()
		})
	}
	if dbClient != nil {
		sm.Register("postgres-close", 5*time.Second, func(ctx context.Context) error {
			return dbClient.Close()
		})
	}
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

		if requestID := chimiddleware.GetReqID(r.Context()); requestID != "" {
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
