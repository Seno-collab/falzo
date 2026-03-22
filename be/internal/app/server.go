package app

import (
	"context"
	"falzo-be/internal/auth/application"
	authpostgres "falzo-be/internal/auth/infrastructure/persistence/postgres"
	"falzo-be/internal/auth/infrastructure/security/bcrypt"
	"falzo-be/internal/auth/infrastructure/token"
	authhttp "falzo-be/internal/auth/interfaces/http"
	"falzo-be/pkg/config"
	"falzo-be/pkg/database"
	httpmw "falzo-be/pkg/http/middleware"
	"falzo-be/pkg/logger"
	"falzo-be/pkg/shutdown"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func Run() {
	logger.SetupLogger()
	cfg := config.Load()

	db, err := database.New(cfg.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect postgres")
	}

	accounts := authpostgres.NewAccountRepository(db)
	sessions := authpostgres.NewSessionRepository(db)
	passwords := bcrypt.NewPasswordHasher()
	jwtManager := token.NewJWTManager(cfg.Auth)
	authService := application.New(accounts, sessions, passwords, jwtManager, jwtManager, cfg.Auth.RefreshTokenTTL)
	authHandler := authhttp.New(authService)

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(httpmw.Recover)
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
