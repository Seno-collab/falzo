package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	NATS     NATSConfig
	JWT      JWTConfig
	Auth     AuthConfig
	Google   GoogleConfig
	Game     GameConfig
}

type ServerConfig struct {
	Port                    string
	Env                     string
	WebSocketOriginPatterns []string
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
	Host                string
	Port                int
	Password            string
	DB                  int
	KeyPrefix           string
	PoolSize            int
	MinIdleConns        int
	DialTimeoutSeconds  int
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
}

type NATSConfig struct {
	Enabled bool
	URL     string
	Stream  string
	Subject string
}

type JWTConfig struct {
	Secret              string
	ExpiredMinutes      int
	RefreshExpiredHours int
	ResetExpiredMinutes int
}

type AuthConfig struct {
	MaxLoginAttempts int
	LockMinutes      int
}

type GoogleConfig struct {
	ClientID string
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
	serverEnv := v.GetString("SERVER_ENV")
	webSocketOriginPatterns := splitCommaSeparated(v.GetString("WEBSOCKET_ORIGIN_PATTERNS"))
	if len(webSocketOriginPatterns) == 0 && serverEnv != "production" {
		webSocketOriginPatterns = []string{"localhost:3000", "127.0.0.1:3000"}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:                    v.GetString("SERVER_PORT"),
			Env:                     serverEnv,
			WebSocketOriginPatterns: webSocketOriginPatterns,
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
			Host:                v.GetString("REDIS_HOST"),
			Port:                v.GetInt("REDIS_PORT"),
			Password:            v.GetString("REDIS_PASSWORD"),
			DB:                  v.GetInt("REDIS_DB"),
			KeyPrefix:           v.GetString("REDIS_KEY_PREFIX"),
			PoolSize:            v.GetInt("REDIS_POOL_SIZE"),
			MinIdleConns:        v.GetInt("REDIS_MIN_IDLE_CONNS"),
			DialTimeoutSeconds:  v.GetInt("REDIS_DIAL_TIMEOUT_SECONDS"),
			ReadTimeoutSeconds:  v.GetInt("REDIS_READ_TIMEOUT_SECONDS"),
			WriteTimeoutSeconds: v.GetInt("REDIS_WRITE_TIMEOUT_SECONDS"),
		},
		NATS: NATSConfig{
			Enabled: v.GetBool("NATS_ENABLED"),
			URL:     v.GetString("NATS_URL"),
			Stream:  v.GetString("NATS_ALERT_STREAM"),
			Subject: v.GetString("NATS_ALERT_SUBJECT"),
		},
		JWT: JWTConfig{
			Secret:              v.GetString("JWT_SECRET"),
			ExpiredMinutes:      v.GetInt("JWT_EXPIRED_MINUTES"),
			RefreshExpiredHours: v.GetInt("JWT_REFRESH_EXPIRED_HOURS"),
			ResetExpiredMinutes: v.GetInt("JWT_RESET_EXPIRED_MINUTES"),
		},
		Auth: AuthConfig{MaxLoginAttempts: v.GetInt("AUTH_MAX_LOGIN_ATTEMPTS"), LockMinutes: v.GetInt("AUTH_LOCK_MINUTES")},
		Google: GoogleConfig{
			ClientID: v.GetString("GOOGLE_CLIENT_ID"),
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
	v.SetDefault("REDIS_POOL_SIZE", 20)
	v.SetDefault("REDIS_MIN_IDLE_CONNS", 5)
	v.SetDefault("REDIS_DIAL_TIMEOUT_SECONDS", 5)
	v.SetDefault("REDIS_READ_TIMEOUT_SECONDS", 3)
	v.SetDefault("REDIS_WRITE_TIMEOUT_SECONDS", 3)
	v.SetDefault("NATS_ENABLED", true)
	v.SetDefault("NATS_URL", "nats://localhost:4222")
	v.SetDefault("NATS_ALERT_STREAM", "FALZO_ALERTS")
	v.SetDefault("NATS_ALERT_SUBJECT", "falzo.alerts.error")

	v.SetDefault("JWT_EXPIRED_MINUTES", 60)
	v.SetDefault("JWT_REFRESH_EXPIRED_HOURS", 168)
	v.SetDefault("JWT_RESET_EXPIRED_MINUTES", 15)
	v.SetDefault("AUTH_MAX_LOGIN_ATTEMPTS", 5)
	v.SetDefault("AUTH_LOCK_MINUTES", 15)
	v.SetDefault("REDIS_KEY_PREFIX", "falzo")

	v.SetDefault("GAME_MAX_PLAYERS_PER_ROOM", 12)
	v.SetDefault("GAME_ROOM_TIMEOUT_SECONDS", 3600)
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
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.Auth.MaxLoginAttempts <= 0 || c.Auth.LockMinutes <= 0 {
		return fmt.Errorf("auth limits must be positive")
	}
	if c.Redis.PoolSize <= 0 || c.Redis.MinIdleConns < 0 || c.Redis.MinIdleConns > c.Redis.PoolSize {
		return fmt.Errorf("redis pool settings are invalid")
	}
	if c.Redis.DialTimeoutSeconds <= 0 || c.Redis.ReadTimeoutSeconds <= 0 || c.Redis.WriteTimeoutSeconds <= 0 {
		return fmt.Errorf("redis timeouts must be positive")
	}
	if c.NATS.Enabled && (c.NATS.URL == "" || c.NATS.Stream == "" || c.NATS.Subject == "") {
		return fmt.Errorf("NATS_URL, NATS_ALERT_STREAM and NATS_ALERT_SUBJECT are required when NATS is enabled")
	}
	if c.Google.ClientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	if c.Game.MaxPlayersPerRoom < 4 || c.Game.MaxPlayersPerRoom > 12 {
		return fmt.Errorf("GAME_MAX_PLAYERS_PER_ROOM must be between 4 and 12")
	}
	if c.Game.RoomTimeoutSeconds <= 0 {
		return fmt.Errorf("GAME_ROOM_TIMEOUT_SECONDS must be positive")
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

func splitCommaSeparated(value string) []string {
	values := make([]string, 0)
	for item := range strings.SplitSeq(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}
