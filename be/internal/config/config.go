package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Game     GameConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	UserName string
	Password string
	Database string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret         string
	ExpiredMinutes int
}

type GameConfig struct {
	MaxPlayersPerRoom  int
	RoomTimeoutSeconds int
}

func Load(envPath string) (*Config, error) {
	if envPath == "" {
		envPath = ".env"
	}

	// Load .env cho local/dev.
	// Production có thể không có file .env, mà truyền ENV trực tiếp từ Docker/K8s/CI-CD.
	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			return nil, fmt.Errorf("load env file failed: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("check env file failed: %w", err)
	}

	v := viper.New()
	v.AutomaticEnv()

	setDefaults(v)

	cfg := &Config{
		Server: ServerConfig{
			Port: v.GetString("SERVER_PORT"),
			Env:  v.GetString("SERVER_ENV"),
		},
		Database: DatabaseConfig{
			Host:     v.GetString("DATABASE_HOST"),
			Port:     v.GetInt("DATABASE_PORT"),
			UserName: v.GetString("DATABASE_USER_NAME"),
			Password: v.GetString("DATABASE_PASSWORD"),
			Database: v.GetString("DATABASE_NAME"),
			SSLMode:  v.GetString("DATABASE_SSL_MODE"),
		},
		Redis: RedisConfig{
			Host:     v.GetString("REDIS_HOST"),
			Port:     v.GetInt("REDIS_PORT"),
			Password: v.GetString("REDIS_PASSWORD"),
			DB:       v.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:         v.GetString("JWT_SECRET"),
			ExpiredMinutes: v.GetInt("JWT_EXPIRED_MINUTES"),
		},
		Game: GameConfig{
			MaxPlayersPerRoom:  v.GetInt("GAME_MAX_PLAYERS_PER_ROOM"),
			RoomTimeoutSeconds: v.GetInt("GAME_ROOM_TIMEOUT_SECONDS"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("SERVER_ENV", "development")

	v.SetDefault("DATABASE_HOST", "localhost")
	v.SetDefault("DATABASE_PORT", 5432)
	v.SetDefault("DATABASE_SSL_MODE", "disable")

	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_DB", 0)

	v.SetDefault("JWT_EXPIRED_MINUTES", 60)

	v.SetDefault("GAME_MAX_PLAYERS_PER_ROOM", 4)
	v.SetDefault("GAME_ROOM_TIMEOUT_SECONDS", 300)
}

func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("DATABASE_HOST is required")
	}

	if c.Database.UserName == "" {
		return fmt.Errorf("DATABASE_USER_NAME is required")
	}

	if c.Database.Password == "" {
		return fmt.Errorf("DATABASE_PASSWORD is required")
	}

	if c.Database.Database == "" {
		return fmt.Errorf("DATABASE_NAME is required")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	return nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.UserName,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}

func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
