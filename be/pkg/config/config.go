package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App        AppConfig
	HTTP       HTTPConfig
	Auth       AuthConfig
	Postgres   PostgresConfig
	Redis      RedisConfig
	Engagement EngagementConfig
	Upload     UploadConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Addr                 string
	ShutdownTimeout      time.Duration
	TrustProxyHeaders    bool
	CORSAllowedOrigins   []string
	CORSAllowedMethods   []string
	CORSAllowedHeaders   []string
	CORSAllowCredentials bool
	CORSMaxAgeSeconds    int
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
	MinOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type EngagementConfig struct {
	ClaimMinIdle time.Duration
	MaxRetries   int
}

type UploadConfig struct {
	SeaweedFSBaseURL string
	SeaweedFSTimeout time.Duration
	MaxSize          int64
	AllowedTypes     []string
}

func Load() Config {
	loadDotEnvFromWorkingDir()

	return Config{
		App: AppConfig{
			Name: GetEnv("APP_NAME", "falzo-api"),
			Env:  GetEnv("APP_ENV", "development"),
		},
		HTTP: HTTPConfig{
			Addr:                 GetEnv("HTTP_ADDR", ":8080"),
			ShutdownTimeout:      GetDuration("HTTP_SHUTDOWN_TIMEOUT", 60*time.Second),
			TrustProxyHeaders:    GetBool("HTTP_TRUST_PROXY_HEADERS", false),
			CORSAllowedOrigins:   getCSV("HTTP_CORS_ALLOWED_ORIGINS", []string{"*"}),
			CORSAllowedMethods:   getCSV("HTTP_CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			CORSAllowedHeaders:   getCSV("HTTP_CORS_ALLOWED_HEADERS", []string{"Accept", "Authorization", "Content-Type", "Origin", "X-Requested-With"}),
			CORSAllowCredentials: GetBool("HTTP_CORS_ALLOW_CREDENTIALS", false),
			CORSMaxAgeSeconds:    GetInt("HTTP_CORS_MAX_AGE_SECONDS", 600),
		},
		Auth: AuthConfig{
			JWTSecret:                  GetEnv("AUTH_JWT_SECRET", "change-me-in-production"),
			TokenTTL:                   GetDuration("AUTH_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL:            GetDuration("AUTH_REFRESH_TOKEN_TTL", 168*time.Hour),
			RateLimitPerMin:            GetInt("AUTH_RATE_LIMIT_PER_MIN", 60),
			DependencyFailureThreshold: GetInt("AUTH_DEPENDENCY_FAILURE_THRESHOLD", 5),
			DependencyCoolDown:         GetDuration("AUTH_DEPENDENCY_COOLDOWN", 15*time.Second),
		},
		Postgres: PostgresConfig{
			Host:            GetEnv("POSTGRES_HOST", "127.0.0.1"),
			Port:            GetEnv("POSTGRES_PORT", "5432"),
			Database:        GetEnv("POSTGRES_DB", ""),
			User:            GetEnv("POSTGRES_USER", ""),
			Password:        GetEnv("POSTGRES_PASSWORD", ""),
			SSLMode:         GetEnv("POSTGRES_SSL_MODE", "disable"),
			MaxOpenConns:    GetInt("POSTGRES_MAX_OPEN_CONNS", 25),
			MinOpenConns:    GetInt("POSTGRES_MIN_OPEN_CONNS", 5),
			MaxIdleConns:    GetInt("POSTGRES_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: GetDuration("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: GetDuration("POSTGRES_CONN_MAX_IDLE_TIME", 30*time.Minute),
		},
		Redis: RedisConfig{
			Addr:     GetEnv("REDIS_ADDR", "127.0.0.1:6379"),
			Password: GetEnv("REDIS_PASSWORD", ""),
			DB:       GetInt("REDIS_DB", 0),
		},
		Engagement: EngagementConfig{
			ClaimMinIdle: GetDuration("ENGAGEMENT_CLAIM_MIN_IDLE", 30*time.Second),
			MaxRetries:   GetInt("ENGAGEMENT_MAX_RETRIES", 10),
		},
		Upload: UploadConfig{
			SeaweedFSBaseURL: GetEnv("SEAWEEDFS_BASE_URL", "http://127.0.0.1:8888"),
			SeaweedFSTimeout: GetDuration("SEAWEEDFS_TIMEOUT", 10*time.Second),
			MaxSize:          GetInt64("UPLOAD_MAX_SIZE", 10*1024*1024),
			AllowedTypes:     getCSV("ALLOWED_IMAGE_TYPES", []string{"image/jpeg", "image/png", "image/webp"}),
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

func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func GetInt(key string, fallback int) int {
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

func GetInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func GetDuration(key string, fallback time.Duration) time.Duration {
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

func GetBool(key string, fallback bool) bool {
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

func getCSV(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return append([]string(nil), fallback...)
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		items = append(items, trimmed)
	}

	if len(items) == 0 {
		return append([]string(nil), fallback...)
	}

	return items
}
