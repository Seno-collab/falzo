package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"be/internal/alerting"
	"be/internal/config"
	natsinfra "be/internal/insfrastructure/nats"
	"be/internal/insfrastructure/telegram"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true})).With(
		slog.String("service", "falzo-telegram-bot"),
	)
	slog.SetDefault(logger)
	if err := run(); err != nil {
		logger.Error("telegram bot stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadTelegramBot(".env")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	telegramClient, err := telegram.NewClient(
		cfg.Telegram.APIBaseURL,
		cfg.Telegram.BotToken,
		cfg.Telegram.ChatID,
		nil,
	)
	if err != nil {
		return fmt.Errorf("create Telegram client: %w", err)
	}
	consumer, err := natsinfra.NewAlertConsumer(cfg.NATS.URL, cfg.NATS.Stream, cfg.NATS.Subject, cfg.NATS.Durable)
	if err != nil {
		return fmt.Errorf("create alert consumer: %w", err)
	}
	defer func(consumer *natsinfra.AlertConsumer) {
		_ = consumer.Close()
	}(consumer)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	healthServer := &http.Server{
		Addr:              ":" + cfg.Telegram.HealthPort,
		Handler:           newHealthHandler(),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	errCh := make(chan error, 2)
	go func() {
		slog.Info("telegram health server listening", slog.String("address", healthServer.Addr))
		errCh <- healthServer.ListenAndServe()
	}()

	slog.Info("telegram alert bot listening", slog.String("subject", cfg.NATS.Subject), slog.String("durable", cfg.NATS.Durable))
	go func() {
		errCh <- consumer.Run(ctx, func(ctx context.Context, event alerting.Event) error {
			if err := telegramClient.SendAlert(ctx, event); err != nil {
				slog.ErrorContext(ctx, "send Telegram alert", slog.String("event_id", event.ID), slog.Any("error", err))
				return err
			}
			slog.InfoContext(ctx, "Telegram alert delivered", slog.String("event_id", event.ID))
			return nil
		})
	}()

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-errCh:
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shutdownErr := healthServer.Shutdown(shutdownCtx)
	if runErr != nil && !errors.Is(runErr, http.ErrServerClosed) {
		return runErr
	}
	if shutdownErr != nil && !errors.Is(shutdownErr, http.ErrServerClosed) {
		return fmt.Errorf("shutdown Telegram health server: %w", shutdownErr)
	}
	return nil
}

func newHealthHandler() http.Handler {
	mux := http.NewServeMux()
	respond := func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			response.Header().Set("Allow", "GET, HEAD")
			http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusOK)
		if request.Method == http.MethodGet {
			_, _ = response.Write([]byte(`{"status":"ok","service":"falzo-telegram-bot"}`))
		}
	}
	mux.HandleFunc("/healthz", respond)
	mux.HandleFunc("/health/live", respond)
	mux.HandleFunc("/health/ready", respond)
	return mux
}
