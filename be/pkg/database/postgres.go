package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"falzo-be/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Client defines the database client contract exposed to the application.
type Client interface {
	Pool() *pgxpool.Pool
	Close() error
}

// postgresClient implements Client using a PostgreSQL-backed pgxpool.Pool connection.
type postgresClient struct {
	pool *pgxpool.Pool
}

func New(cfg config.PostgresConfig) (Client, error) {
	sslMode := strings.TrimSpace(cfg.SSLMode)
	if sslMode == "" {
		sslMode = "disable"
	}

	dsnURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   "/" + cfg.Database,
	}
	query := dsnURL.Query()
	query.Set("sslmode", sslMode)
	// Force per-session timezone so NOW()/CURRENT_TIMESTAMP are UTC+0.
	query.Set("timezone", "UTC")
	dsnURL.RawQuery = query.Encode()
	dsn := dsnURL.String()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL DSN: %w", err)
	}
	config.MaxConns = int32(cfg.MaxOpenConns)
	config.MinConns = int32(cfg.MinOpenConns)
	config.MaxConnIdleTime = cfg.ConnMaxIdleTime
	config.MaxConnLifetime = cfg.ConnMaxLifetime
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
	}

	// Verify the connection by pinging the database.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}
	return &postgresClient{pool: pool}, nil
}

func (e *postgresClient) Pool() *pgxpool.Pool {
	if e == nil {
		return nil
	}

	return e.pool
}

func (e *postgresClient) Close() error {
	if e == nil || e.pool == nil {
		return nil
	}

	e.pool.Close()
	return nil
}
