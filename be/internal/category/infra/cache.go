package categoryInfra

import (
	"context"
	"encoding/json"
	"time"

	"falzo-be/internal/category"
	pkgcache "falzo-be/pkg/cache"

	goredis "github.com/redis/go-redis/v9"
)

const categoriesListCacheKey = "categories:list:v1"

type CachedRepository struct {
	next    category.Repository
	redis   *goredis.Client
	listTTL time.Duration
}

func NewCachedRepository(next category.Repository, cache pkgcache.Client, listTTL time.Duration) category.Repository {
	if next == nil || cache == nil || cache.Client() == nil || listTTL <= 0 {
		return next
	}

	return &CachedRepository{next: next, redis: cache.Client(), listTTL: listTTL}
}

func (r *CachedRepository) Create(ctx context.Context, input category.CategoryCreateInput) error {
	if err := r.next.Create(ctx, input); err != nil {
		return err
	}
	r.invalidateList(ctx)
	return nil
}

func (r *CachedRepository) GetByID(ctx context.Context, id uint64) (category.Category, error) {
	return r.next.GetByID(ctx, id)
}

func (r *CachedRepository) GetBySlug(ctx context.Context, slug string) (category.Category, error) {
	return r.next.GetBySlug(ctx, slug)
}

func (r *CachedRepository) List(ctx context.Context) ([]category.Category, error) {
	value, err := r.redis.Get(ctx, categoriesListCacheKey).Result()
	if err == nil {
		var items []category.Category
		if decodeErr := json.Unmarshal([]byte(value), &items); decodeErr == nil {
			return items, nil
		}
		_ = r.redis.Del(ctx, categoriesListCacheKey).Err()
	}

	items, err := r.next.List(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(items)
	if err == nil {
		_ = r.redis.Set(ctx, categoriesListCacheKey, payload, r.listTTL).Err()
	}
	return items, nil
}

func (r *CachedRepository) Update(ctx context.Context, id uint64, name, slug string) (category.Category, error) {
	item, err := r.next.Update(ctx, id, name, slug)
	if err != nil {
		return category.Category{}, err
	}
	r.invalidateList(ctx)
	return item, nil
}

func (r *CachedRepository) Delete(ctx context.Context, id uint64) error {
	if err := r.next.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidateList(ctx)
	return nil
}

func (r *CachedRepository) invalidateList(ctx context.Context) {
	_ = r.redis.Del(ctx, categoriesListCacheKey).Err()
}

var _ category.Repository = (*CachedRepository)(nil)
