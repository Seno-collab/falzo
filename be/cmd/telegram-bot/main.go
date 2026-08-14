package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"be/internal/alerting"
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
	natsURL := envOrDefault("NATS_URL", "nats://localhost:4222")
	stream := envOrDefault("NATS_ALERT_STREAM", "FALZO_ALERTS")
	subject := envOrDefault("NATS_ALERT_SUBJECT", "falzo.alerts.error")
	durable := envOrDefault("NATS_ALERT_DURABLE", "falzo-telegram-error-bot")
	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	chatID := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID"))
	if token == "" || chatID == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID are required")
	}

	telegramClient, err := telegram.NewClient(
		envOrDefault("TELEGRAM_API_BASE_URL", "https://api.telegram.org"),
		token,
		chatID,
		nil,
	)
	if err != nil {
		return fmt.Errorf("create Telegram client: %w", err)
	}
	consumer, err := natsinfra.NewAlertConsumer(natsURL, stream, subject, durable)
	if err != nil {
		return fmt.Errorf("create alert consumer: %w", err)
	}
	defer consumer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	slog.Info("telegram alert bot listening", slog.String("subject", subject), slog.String("durable", durable))
	return consumer.Run(ctx, func(ctx context.Context, event alerting.Event) error {
		if err := telegramClient.SendAlert(ctx, event); err != nil {
			slog.ErrorContext(ctx, "send Telegram alert", slog.String("event_id", event.ID), slog.Any("error", err))
			return err
		}
		slog.InfoContext(ctx, "Telegram alert delivered", slog.String("event_id", event.ID))
		return nil
	})
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
