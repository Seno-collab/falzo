package cache

import (
	"context"
	"time"

	"falzo/internal/config"

	"github.com/redis/go-redis/v9"
)

// Client defines the cache client contract used by the application.
type Client interface {
	Client() *redis.Client
	Close() error
}

// redisClient implements Client using a Redis client.
type redisClient struct {
	client *redis.Client
}

func New(cfg config.RedisConfig) (Client, error) {
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

	return &redisClient{client: client}, nil
}

func (r *redisClient) Client() *redis.Client {
	if r == nil {
		return nil
	}

	return r.client
}

func (r *redisClient) Close() error {
	if r == nil || r.client == nil {
		return nil
	}

	return r.client.Close()
}
