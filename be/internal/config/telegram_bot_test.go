package config

import "testing"

func TestLoadTelegramBot(t *testing.T) {
	t.Setenv("NATS_URL", " nats://nats:4222 ")
	t.Setenv("NATS_ALERT_STREAM", " TEST_ALERTS ")
	t.Setenv("NATS_ALERT_SUBJECT", " test.alerts.error ")
	t.Setenv("NATS_ALERT_DURABLE", " test-telegram-bot ")
	t.Setenv("TELEGRAM_HEALTH_PORT", " 9091 ")
	t.Setenv("TELEGRAM_BOT_TOKEN", " token ")
	t.Setenv("TELEGRAM_CHAT_ID", " chat-id ")
	t.Setenv("TELEGRAM_API_BASE_URL", " https://telegram.example.test ")

	cfg, err := LoadTelegramBot(t.TempDir() + "/missing.env")
	if err != nil {
		t.Fatalf("LoadTelegramBot() error = %v", err)
	}

	if cfg.NATS.URL != "nats://nats:4222" || cfg.NATS.Stream != "TEST_ALERTS" ||
		cfg.NATS.Subject != "test.alerts.error" || cfg.NATS.Durable != "test-telegram-bot" {
		t.Fatalf("unexpected NATS config: %+v", cfg.NATS)
	}
	if cfg.Telegram.HealthPort != "9091" || cfg.Telegram.BotToken != "token" ||
		cfg.Telegram.ChatID != "chat-id" || cfg.Telegram.APIBaseURL != "https://telegram.example.test" {
		t.Fatalf("unexpected Telegram config: %+v", cfg.Telegram)
	}
}

func TestLoadTelegramBotUsesDefaults(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_CHAT_ID", "chat-id")

	cfg, err := LoadTelegramBot(t.TempDir() + "/missing.env")
	if err != nil {
		t.Fatalf("LoadTelegramBot() error = %v", err)
	}

	if cfg.NATS.URL != "nats://localhost:4222" || cfg.NATS.Stream != "FALZO_ALERTS" ||
		cfg.NATS.Subject != "falzo.alerts.error" || cfg.NATS.Durable != "falzo-telegram-error-bot" {
		t.Fatalf("unexpected default NATS config: %+v", cfg.NATS)
	}
	if cfg.Telegram.HealthPort != "8081" || cfg.Telegram.APIBaseURL != "https://api.telegram.org" {
		t.Fatalf("unexpected default Telegram config: %+v", cfg.Telegram)
	}
}

func TestLoadTelegramBotRequiresCredentials(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")

	if _, err := LoadTelegramBot(t.TempDir() + "/missing.env"); err == nil {
		t.Fatal("LoadTelegramBot() error = nil, want missing credential error")
	}
}
