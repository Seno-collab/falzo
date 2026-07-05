package main

import (
	"be/internal/api"
	"be/internal/api/handler"
	authapp "be/internal/application/auth"
	"be/internal/config"
	"be/internal/insfrastructure/persistence/postgres"
	redisinfra "be/internal/insfrastructure/redis"
	"be/internal/insfrastructure/security"
	"be/internal/shared/clock"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load(".env")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()
	redisClient, err := redisinfra.NewRedisClient(ctx, cfg.Redis, cfg.Redis.DB)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer redisClient.Close()

	c := clock.NewSystemClock()
	users := postgres.NewUserRepository(db)
	hasher := security.NewBcryptPasswordHasher(12)
	tokens, err := security.NewJWTTokenManager(cfg.JWT.Secret, time.Duration(cfg.JWT.ExpiredMinutes)*time.Minute, time.Duration(cfg.JWT.RefreshExpiredHours)*time.Hour, time.Duration(cfg.JWT.ResetExpiredMinutes)*time.Minute)
	if err != nil {
		return fmt.Errorf("create token manager: %w", err)
	}
	sessions := redisinfra.NewTokenSessionStore(redisClient, cfg.Redis.KeyPrefix)
	register := authapp.NewRegisterUseCase(users, hasher, c)
	login := authapp.NewLoginUseCase(hasher, users, tokens, sessions, cfg.Auth.MaxLoginAttempts, time.Duration(cfg.Auth.LockMinutes)*time.Minute, c)
	refresh := authapp.NewRefreshTokenUseCase(users, tokens, sessions, c)
	forgot := authapp.NewForgotPasswordUseCase(users, tokens, sessions, c)
	reset := authapp.NewResetPasswordUseCase(users, hasher, tokens, sessions, c)
	logout := authapp.NewLogoutUseCase(tokens, sessions, c)
	authHandler := handler.NewAuthHandler(register, login, refresh, forgot, reset, logout, cfg.Server.Env != "production")

	server := &http.Server{Addr: ":" + cfg.Server.Port, Handler: api.NewRouter(authHandler), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()
	log.Printf("API listening on %s", server.Addr)
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
	return nil
}
