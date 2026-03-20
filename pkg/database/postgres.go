package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"falzo/pkg/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Client defines the database client contract exposed to the application.
type Client interface {
	DB() *sql.DB
	Close() error
}

// postgresClient implements Client using a PostgreSQL-backed sql.DB connection.
type postgresClient struct {
	db *sql.DB
}

func New(cfg config.PostgresConfig) (Client, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &postgresClient{db: db}, nil
}

func (e *postgresClient) DB() *sql.DB {
	if e == nil {
		return nil
	}

	return e.db
}

func (e *postgresClient) Close() error {
	if e == nil || e.db == nil {
		return nil
	}

	return e.db.Close()
}
