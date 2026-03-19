package cache

import (
	"context"
	"time"

	"falzo/internal/config"

	"github.com/redis/go-redis/v9"
)

// Chi defines the cache client contract used by the application.
type Chi interface {
	Client() *redis.Client
	Close() error
}

// em implements Chi using a Redis client.
type em struct {
	client *redis.Client
}

func New(cfg config.RedisConfig) (Chi, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &em{client: client}, nil
}

func (r *em) Client() *redis.Client {
	if r == nil {
		return nil
	}

	return r.client
}

func (r *em) Close() error {
	if r == nil || r.client == nil {
		return nil
	}

	return r.client.Close()
}
