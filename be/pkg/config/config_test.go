package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("APP_NAME", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("GRPC_ADDR", "")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "")
	t.Setenv("HTTP_TRUST_PROXY_HEADERS", "")
	t.Setenv("POSTGRES_SSL_MODE", "")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("SEAWEEDFS_BASE_URL", "")
	t.Setenv("UPLOAD_MAX_SIZE", "")
	t.Setenv("ALLOWED_IMAGE_TYPES", "")
	t.Setenv("UPLOAD_RATE_LIMIT_PER_MIN", "")

	cfg := Load()

	if cfg.App.Name != "falzo-api" {
		t.Fatalf("expected default app name, got %q", cfg.App.Name)
	}

	if cfg.HTTP.Addr != ":8080" {
		t.Fatalf("expected default http addr, got %q", cfg.HTTP.Addr)
	}

	if cfg.GRPC.Addr != ":9090" {
		t.Fatalf("expected default grpc addr, got %q", cfg.GRPC.Addr)
	}

	if cfg.HTTP.ShutdownTimeout != 60*time.Second {
		t.Fatalf("expected default shutdown timeout, got %v", cfg.HTTP.ShutdownTimeout)
	}

	if cfg.HTTP.TrustProxyHeaders {
		t.Fatal("expected default trusted proxy headers to be disabled")
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

	if cfg.Engagement.ClaimMinIdle != 30*time.Second {
		t.Fatalf("expected default engagement claim min idle, got %v", cfg.Engagement.ClaimMinIdle)
	}

	if cfg.Engagement.MaxRetries != 10 {
		t.Fatalf("expected default engagement max retries, got %d", cfg.Engagement.MaxRetries)
	}

	if cfg.Postgres.SSLMode != "disable" {
		t.Fatalf("expected default postgres ssl mode disable, got %q", cfg.Postgres.SSLMode)
	}

	if cfg.Upload.SeaweedFSBaseURL != "http://127.0.0.1:8888" {
		t.Fatalf("expected default seaweedfs base url, got %q", cfg.Upload.SeaweedFSBaseURL)
	}

	if cfg.Upload.MaxSize != 10*1024*1024 {
		t.Fatalf("expected default upload max size, got %d", cfg.Upload.MaxSize)
	}

	if cfg.Upload.RateLimitPerMin != 20 {
		t.Fatalf("expected default upload rate limit, got %d", cfg.Upload.RateLimitPerMin)
	}
}

func TestLoadUsesEnvOverrides(t *testing.T) {
	t.Setenv("APP_NAME", "falzo-test")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("GRPC_ADDR", ":9091")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "15s")
	t.Setenv("HTTP_TRUST_PROXY_HEADERS", "true")
	t.Setenv("POSTGRES_SSL_MODE", "require")
	t.Setenv("REDIS_DB", "4")
	t.Setenv("AUTH_RATE_LIMIT_PER_MIN", "15")
	t.Setenv("AUTH_DEPENDENCY_FAILURE_THRESHOLD", "7")
	t.Setenv("AUTH_DEPENDENCY_COOLDOWN", "20s")
	t.Setenv("ENGAGEMENT_CLAIM_MIN_IDLE", "45s")
	t.Setenv("ENGAGEMENT_MAX_RETRIES", "5")
	t.Setenv("SEAWEEDFS_BASE_URL", "http://seaweed:8888")
	t.Setenv("UPLOAD_MAX_SIZE", "2048")
	t.Setenv("ALLOWED_IMAGE_TYPES", "image/png,image/webp")
	t.Setenv("UPLOAD_RATE_LIMIT_PER_MIN", "3")

	cfg := Load()

	if cfg.App.Name != "falzo-test" {
		t.Fatalf("expected env app name, got %q", cfg.App.Name)
	}

	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("expected env http addr, got %q", cfg.HTTP.Addr)
	}

	if cfg.GRPC.Addr != ":9091" {
		t.Fatalf("expected env grpc addr, got %q", cfg.GRPC.Addr)
	}

	if cfg.HTTP.ShutdownTimeout != 15*time.Second {
		t.Fatalf("expected env shutdown timeout, got %v", cfg.HTTP.ShutdownTimeout)
	}

	if !cfg.HTTP.TrustProxyHeaders {
		t.Fatal("expected env trusted proxy headers to be enabled")
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

	if cfg.Engagement.ClaimMinIdle != 45*time.Second {
		t.Fatalf("expected env engagement claim min idle, got %v", cfg.Engagement.ClaimMinIdle)
	}

	if cfg.Engagement.MaxRetries != 5 {
		t.Fatalf("expected env engagement max retries, got %d", cfg.Engagement.MaxRetries)
	}

	if cfg.Postgres.SSLMode != "require" {
		t.Fatalf("expected env postgres ssl mode, got %q", cfg.Postgres.SSLMode)
	}

	if cfg.Upload.SeaweedFSBaseURL != "http://seaweed:8888" {
		t.Fatalf("expected env seaweedfs base url, got %q", cfg.Upload.SeaweedFSBaseURL)
	}

	if cfg.Upload.MaxSize != 2048 {
		t.Fatalf("expected env upload max size, got %d", cfg.Upload.MaxSize)
	}

	if len(cfg.Upload.AllowedTypes) != 2 || cfg.Upload.AllowedTypes[0] != "image/png" || cfg.Upload.AllowedTypes[1] != "image/webp" {
		t.Fatalf("expected env upload allowed types, got %v", cfg.Upload.AllowedTypes)
	}

	if cfg.Upload.RateLimitPerMin != 3 {
		t.Fatalf("expected env upload rate limit, got %d", cfg.Upload.RateLimitPerMin)
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

func TestValidateRejectsDisabledPostgresTLSOutsideDevelopment(t *testing.T) {
	cfg := Config{
		App: AppConfig{Env: "production"},
		Auth: AuthConfig{
			JWTSecret: "01234567890123456789012345678901",
		},
		Postgres: PostgresConfig{
			SSLMode: "disable",
		},
	}

	if err := Validate(cfg); err == nil {
		t.Fatal("expected validation error for disabled postgres ssl mode")
	}
}

func TestLoadReadsValuesFromDotEnvFile(t *testing.T) {
	tempDir := t.TempDir()
	dotenvPath := filepath.Join(tempDir, ".env")
	content := "APP_NAME=from-dotenv\nHTTP_ADDR=:9999\nAUTH_RATE_LIMIT_PER_MIN=21\n"
	if err := os.WriteFile(dotenvPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWD) })

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	t.Setenv("APP_NAME", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("AUTH_RATE_LIMIT_PER_MIN", "")

	cfg := Load()

	if cfg.App.Name != "from-dotenv" {
		t.Fatalf("expected app name from .env, got %q", cfg.App.Name)
	}

	if cfg.HTTP.Addr != ":9999" {
		t.Fatalf("expected http addr from .env, got %q", cfg.HTTP.Addr)
	}

	if cfg.Auth.RateLimitPerMin != 21 {
		t.Fatalf("expected auth rate limit from .env, got %d", cfg.Auth.RateLimitPerMin)
	}
}

func TestLoadKeepsExistingEnvOverDotEnv(t *testing.T) {
	tempDir := t.TempDir()
	dotenvPath := filepath.Join(tempDir, ".env")
	content := "APP_NAME=from-dotenv\nHTTP_ADDR=:7777\n"
	if err := os.WriteFile(dotenvPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalWD) })

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	t.Setenv("APP_NAME", "from-env")
	t.Setenv("HTTP_ADDR", ":8888")

	cfg := Load()

	if cfg.App.Name != "from-env" {
		t.Fatalf("expected app name from env to win, got %q", cfg.App.Name)
	}

	if cfg.HTTP.Addr != ":8888" {
		t.Fatalf("expected http addr from env to win, got %q", cfg.HTTP.Addr)
	}
}
