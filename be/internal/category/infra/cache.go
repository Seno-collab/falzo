package categoryInfra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"falzo-be/internal/category"
	"falzo-be/internal/i18n"
	pkgcache "falzo-be/pkg/cache"

	goredis "github.com/redis/go-redis/v9"
)

const categoriesListCacheKeyPrefix = "categories:list:v2"

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
	cacheKey := categoriesListCacheKey(i18n.LocaleFromContext(ctx))
	value, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var items []category.Category
		if decodeErr := json.Unmarshal([]byte(value), &items); decodeErr == nil {
			return items, nil
		}
		_ = r.redis.Del(ctx, cacheKey).Err()
	}

	items, err := r.next.List(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(items)
	if err == nil {
		_ = r.redis.Set(ctx, cacheKey, payload, r.listTTL).Err()
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
	_ = r.redis.Del(
		ctx,
		categoriesListCacheKey(i18n.English),
		categoriesListCacheKey(i18n.Vietnamese),
	).Err()
}

var _ category.Repository = (*CachedRepository)(nil)

func categoriesListCacheKey(locale string) string {
	return fmt.Sprintf("%s:%s", categoriesListCacheKeyPrefix, locale)
}
