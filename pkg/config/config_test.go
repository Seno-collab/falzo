package config

import (
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("APP_NAME", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "")
	t.Setenv("REDIS_ADDR", "")

	cfg := Load()

	if cfg.App.Name != "falzo-api" {
		t.Fatalf("expected default app name, got %q", cfg.App.Name)
	}

	if cfg.HTTP.Addr != ":8080" {
		t.Fatalf("expected default http addr, got %q", cfg.HTTP.Addr)
	}

	if cfg.HTTP.ShutdownTimeout != 60*time.Second {
		t.Fatalf("expected default shutdown timeout, got %v", cfg.HTTP.ShutdownTimeout)
	}

	if cfg.Redis.Addr != "127.0.0.1:6379" {
		t.Fatalf("expected default redis addr, got %q", cfg.Redis.Addr)
	}
}

func TestLoadUsesEnvOverrides(t *testing.T) {
	t.Setenv("APP_NAME", "falzo-test")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "15s")
	t.Setenv("REDIS_DB", "4")

	cfg := Load()

	if cfg.App.Name != "falzo-test" {
		t.Fatalf("expected env app name, got %q", cfg.App.Name)
	}

	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("expected env http addr, got %q", cfg.HTTP.Addr)
	}

	if cfg.HTTP.ShutdownTimeout != 15*time.Second {
		t.Fatalf("expected env shutdown timeout, got %v", cfg.HTTP.ShutdownTimeout)
	}

	if cfg.Redis.DB != 4 {
		t.Fatalf("expected env redis db, got %d", cfg.Redis.DB)
	}
}
