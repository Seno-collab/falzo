package app

import (
	"context"
	"falzo-be/internal/auth/application"
	authCache "falzo-be/internal/auth/infrastructure/persistence/cache"
	authPostgres "falzo-be/internal/auth/infrastructure/persistence/postgres"
	"falzo-be/internal/auth/infrastructure/security/bcrypt"
	"falzo-be/internal/auth/infrastructure/token"
	authHTTP "falzo-be/internal/auth/interfaces/http"
	"falzo-be/pkg/cache"
	"falzo-be/pkg/config"
	"falzo-be/pkg/database"
	httpMiddleware "falzo-be/pkg/http/middleware"
	"falzo-be/pkg/logger"
	"falzo-be/pkg/shutdown"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func Run() {
	// Keep application-wide local time aligned to UTC (UTC+0).
	time.Local = time.UTC

	config.BootstrapEnv()
	logger.SetupLogger()
	cfg := config.Load()
	if err := config.Validate(cfg); err != nil {
		log.Fatal().Err(err).Msg("invalid configuration")
	}

	db, err := database.New(cfg.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect postgres")
	}

	accounts := authPostgres.NewAccountRepository(db)
	sessions := authPostgres.NewSessionRepository(db)
	redisClient, err := cache.New(cfg.Redis)
	if err != nil {
		log.Warn().Err(err).Msg("redis unavailable, continuing without session cache")
	} else {
		sessions = authCache.NewSessionRepository(sessions, redisClient, cfg.Auth.TokenTTL)
	}
	passwords := bcrypt.NewPasswordHasher()
	jwtManager := token.NewJWTManager(cfg.Auth)
	authService := application.New(accounts, sessions, passwords, jwtManager, jwtManager, cfg.Auth.RefreshTokenTTL)
	authRateLimit := httpMiddleware.NewIPRateLimiter(cfg.Auth.RateLimitPerMin, time.Minute)
	authProtector := authHTTP.WithProtectorConfig(cfg.Auth.RateLimitPerMin, cfg.Auth.DependencyFailureThreshold, cfg.Auth.DependencyCoolDown)
	authHandler := authHTTP.New(authService, authProtector, authHTTP.WithPublicMiddlewares(authRateLimit))

	r := chi.NewRouter()
	if cfg.HTTP.TrustProxyHeaders {
		r.Use(middleware.RealIP)
	}
	r.Use(httpMiddleware.CORS(httpMiddleware.CORSConfig{
		AllowedOrigins:   cfg.HTTP.CORSAllowedOrigins,
		AllowedMethods:   cfg.HTTP.CORSAllowedMethods,
		AllowedHeaders:   cfg.HTTP.CORSAllowedHeaders,
		AllowCredentials: cfg.HTTP.CORSAllowCredentials,
		MaxAgeSeconds:    cfg.HTTP.CORSMaxAgeSeconds,
	}))
	r.Use(middleware.RequestID)
	r.Use(httpMiddleware.Recover)
	r.Use(logger.RequestLogger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	})
	r.Mount("/auth", authHandler.Routes())

	sm := shutdown.NewManager()
	srv := &http.Server{Addr: cfg.HTTP.Addr, Handler: r}
	sm.Register("http-stop-accepting", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
		// Stop accepting new requests immediately
		return srv.Shutdown(ctx)
	})
	sm.Register("postgres-close", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
		return db.Close()
	})
	if redisClient != nil {
		sm.Register("redis-close", cfg.HTTP.ShutdownTimeout, func(ctx context.Context) error {
			return redisClient.Close()
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
