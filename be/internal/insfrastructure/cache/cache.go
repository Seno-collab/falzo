package cache

import (
	cacheports "be/internal/application/ports/cache"
	"context"
	"encoding/json"
	"errors"
	"time"
)

type LoaderFunc[T any] func(ctx context.Context) (*T, error)

type GenericCache[T any] struct {
	store  cacheports.Store
	key    string
	ttl    time.Duration
	loader LoaderFunc[T]
}

func NewGenericCache[T any](store cacheports.Store, key string, ttl time.Duration, loader LoaderFunc[T]) *GenericCache[T] {
	return &GenericCache[T]{
		store:  store,
		key:    key,
		ttl:    ttl,
		loader: loader,
	}
}

func (c *GenericCache[T]) Get(ctx context.Context) (*T, bool, error) {
	bytes, found, err := c.store.Get(ctx, c.key)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	var result T
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, false, err
	}
	return &result, true, nil
}

func (c *GenericCache[T]) Set(ctx context.Context, value *T) error {
	if value == nil {
		return errors.New("cache value is nil")
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.store.Set(ctx, c.key, bytes, c.ttl)
}

func (c *GenericCache[T]) Delete(ctx context.Context) error {
	return c.store.Delete(ctx, c.key)
}

func (c *GenericCache[T]) Reload(ctx context.Context) (*T, error) {
	if c.loader == nil {
		return nil, errors.New("cache loader is nil")
	}

	value, err := c.loader(ctx)
	if err != nil {
		return nil, err
	}

	if value == nil {
		_ = c.Delete(ctx)
		return nil, nil
	}

	if err := c.Set(ctx, value); err != nil {
		return value, err
	}

	return value, nil
}
