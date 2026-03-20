package cache

import (
	"context"
	"time"

	"falzo/pkg/config"

	goredis "github.com/redis/go-redis/v9"
)

// Client defines the cache client contract used by the application.
type Client interface {
	Client() *goredis.Client
	Close() error
}

// redisClient implements Client using a Redis client.
type redisClient struct {
	client *goredis.Client
}

func New(cfg config.RedisConfig) (Client, error) {
	client := goredis.NewClient(&goredis.Options{
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

func (r *redisClient) Client() *goredis.Client {
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
