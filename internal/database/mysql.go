package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"falzo/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

// Client defines the database client contract exposed to the application.
type Client interface {
	DB() *sql.DB
	Close() error
}

// mysqlClient implements Client using a MySQL-backed sql.DB connection.
type mysqlClient struct {
	db *sql.DB
}

func New(cfg config.MySQLConfig) (Client, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	db, err := sql.Open("mysql", dsn)
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

	return &mysqlClient{db: db}, nil
}

func (e *mysqlClient) DB() *sql.DB {
	if e == nil {
		return nil
	}

	return e.db
}

func (e *mysqlClient) Close() error {
	if e == nil || e.db == nil {
		return nil
	}

	return e.db.Close()
}
