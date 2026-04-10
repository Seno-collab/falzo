package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Auth     AuthConfig
	Postgres PostgresConfig
	Redis    RedisConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Addr              string
	ShutdownTimeout   time.Duration
	TrustProxyHeaders bool
}

type AuthConfig struct {
	JWTSecret                  string
	TokenTTL                   time.Duration
	RefreshTokenTTL            time.Duration
	RateLimitPerMin            int
	DependencyFailureThreshold int
	DependencyCoolDown         time.Duration
}

type PostgresConfig struct {
	Host            string
	Port            string
	Database        string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func Load() Config {
	loadDotEnvFromWorkingDir()

	return Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "falzo-api"),
			Env:  getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPConfig{
			Addr:              getEnv("HTTP_ADDR", ":8080"),
			ShutdownTimeout:   getDuration("HTTP_SHUTDOWN_TIMEOUT", 60*time.Second),
			TrustProxyHeaders: getBool("HTTP_TRUST_PROXY_HEADERS", false),
		},
		Auth: AuthConfig{
			JWTSecret:                  getEnv("AUTH_JWT_SECRET", "change-me-in-production"),
			TokenTTL:                   getDuration("AUTH_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL:            getDuration("AUTH_REFRESH_TOKEN_TTL", 168*time.Hour),
			RateLimitPerMin:            getInt("AUTH_RATE_LIMIT_PER_MIN", 60),
			DependencyFailureThreshold: getInt("AUTH_DEPENDENCY_FAILURE_THRESHOLD", 5),
			DependencyCoolDown:         getDuration("AUTH_DEPENDENCY_COOLDOWN", 15*time.Second),
		},
		Postgres: PostgresConfig{
			Host:            getEnv("POSTGRES_HOST", "127.0.0.1"),
			Port:            getEnv("POSTGRES_PORT", "5432"),
			Database:        getEnv("POSTGRES_DATABASE", ""),
			User:            getEnv("POSTGRES_USER", ""),
			Password:        getEnv("POSTGRES_PASSWORD", ""),
			SSLMode:         getEnv("POSTGRES_SSL_MODE", "disable"),
			MaxOpenConns:    getInt("POSTGRES_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("POSTGRES_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getDuration("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getInt("REDIS_DB", 0),
		},
	}
}

func Validate(cfg Config) error {
	if strings.EqualFold(cfg.App.Env, "development") {
		return nil
	}

	secret := strings.TrimSpace(cfg.Auth.JWTSecret)
	if secret == "" || secret == "change-me-in-production" || len(secret) < 32 {
		return errors.New("AUTH_JWT_SECRET must be set to a strong value outside development")
	}
	if strings.EqualFold(strings.TrimSpace(cfg.Postgres.SSLMode), "disable") || strings.TrimSpace(cfg.Postgres.SSLMode) == "" {
		return errors.New("POSTGRES_SSL_MODE must not be disable outside development")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
