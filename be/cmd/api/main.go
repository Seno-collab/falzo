package main

import (
	"be/internal/config"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})))
	if err := run(); err != nil {
		slog.Error("api stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() (runErr error) {
	cfg, err := config.Load(".env")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, cleanup, err := initializeApplication(ctx, cfg)
	if err != nil {
		return fmt.Errorf("initialize application: %w", err)
	}
	defer cleanup()

	bootstrapLogger := slog.Default()
	slog.SetDefault(app.logger)
	defer slog.SetDefault(bootstrapLogger)
	defer func() {
		if runErr != nil {
			app.logger.Error("api stopped", slog.Any("error", runErr))
		}
	}()

	if err := app.realtimeHub.Start(ctx); err != nil {
		return fmt.Errorf("start realtime hub: %w", err)
	}
	defer app.realtimeHub.Close()
	app.phaseScheduler.Start(ctx)

	slog.Info("api starting", slog.String("address", app.server.Addr))
	errCh := make(chan error, 1)
	go func() { errCh <- app.server.ListenAndServe() }()
	slog.Info("api listening", slog.String("address", app.server.Addr))

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		slog.Info("api shutting down", slog.String("reason", ctx.Err().Error()))
		app.realtimeHub.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return app.server.Shutdown(shutdownCtx)
	}
	return nil
}
