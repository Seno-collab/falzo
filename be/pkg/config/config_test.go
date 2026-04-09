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

	if cfg.Auth.RateLimitPerMin != 60 {
		t.Fatalf("expected default auth rate limit, got %d", cfg.Auth.RateLimitPerMin)
	}

	if cfg.Auth.DependencyFailureThreshold != 5 {
		t.Fatalf("expected default dependency failure threshold, got %d", cfg.Auth.DependencyFailureThreshold)
	}

	if cfg.Auth.DependencyCoolDown != 15*time.Second {
		t.Fatalf("expected default dependency cooldown, got %v", cfg.Auth.DependencyCoolDown)
	}
}

func TestLoadUsesEnvOverrides(t *testing.T) {
	t.Setenv("APP_NAME", "falzo-test")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "15s")
	t.Setenv("REDIS_DB", "4")
	t.Setenv("AUTH_RATE_LIMIT_PER_MIN", "15")
	t.Setenv("AUTH_DEPENDENCY_FAILURE_THRESHOLD", "7")
	t.Setenv("AUTH_DEPENDENCY_COOLDOWN", "20s")

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

	if cfg.Auth.RateLimitPerMin != 15 {
		t.Fatalf("expected env auth rate limit, got %d", cfg.Auth.RateLimitPerMin)
	}

	if cfg.Auth.DependencyFailureThreshold != 7 {
		t.Fatalf("expected env dependency failure threshold, got %d", cfg.Auth.DependencyFailureThreshold)
	}

	if cfg.Auth.DependencyCoolDown != 20*time.Second {
		t.Fatalf("expected env dependency cooldown, got %v", cfg.Auth.DependencyCoolDown)
	}
}

func TestValidateRejectsWeakJWTSecretOutsideDevelopment(t *testing.T) {
	cfg := Config{
		App: AppConfig{Env: "production"},
		Auth: AuthConfig{
			JWTSecret: "change-me-in-production",
		},
	}

	if err := Validate(cfg); err == nil {
		t.Fatal("expected validation error for weak jwt secret")
	}
}
