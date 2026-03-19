package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	App   AppConfig
	HTTP  HTTPConfig
	Auth  AuthConfig
	MySQL MySQLConfig
	Redis RedisConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Addr            string
	ShutdownTimeout time.Duration
}

type AuthConfig struct {
	JWTSecret     string
	TokenTTL      time.Duration
	AdminUsername string
	AdminPassword string
}

type MySQLConfig struct {
	Host            string
	Port            string
	Database        string
	User            string
	Password        string
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
	return Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "falzo-api"),
			Env:  getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPConfig{
			Addr:            getEnv("HTTP_ADDR", ":8080"),
			ShutdownTimeout: getDuration("HTTP_SHUTDOWN_TIMEOUT", 60*time.Second),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("AUTH_JWT_SECRET", "change-me-in-production"),
			TokenTTL:      getDuration("AUTH_TOKEN_TTL", 24*time.Hour),
			AdminUsername: getEnv("AUTH_ADMIN_USERNAME", "admin"),
			AdminPassword: getEnv("AUTH_ADMIN_PASSWORD", "admin123"),
		},
		MySQL: MySQLConfig{
			Host:            getEnv("MYSQL_HOST", "127.0.0.1"),
			Port:            getEnv("MYSQL_PORT", "3306"),
			Database:        getEnv("MYSQL_DATABASE", ""),
			User:            getEnv("MYSQL_USER", ""),
			Password:        getEnv("MYSQL_PASSWORD", ""),
			MaxOpenConns:    getInt("MYSQL_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("MYSQL_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getDuration("MYSQL_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getInt("REDIS_DB", 0),
		},
	}
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
