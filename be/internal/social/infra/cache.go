package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"falzo-be/internal/social"
	pkgcache "falzo-be/pkg/cache"

	goredis "github.com/redis/go-redis/v9"
)

type CachedRepository struct {
	next       social.Repository
	redis      *goredis.Client
	profileTTL time.Duration
}

func NewCachedRepository(next social.Repository, cache pkgcache.Client, profileTTL time.Duration) social.Repository {
	if next == nil || cache == nil || cache.Client() == nil || profileTTL <= 0 {
		return next
	}

	return &CachedRepository{next: next, redis: cache.Client(), profileTTL: profileTTL}
}

func (r *CachedRepository) GetPublicProfile(ctx context.Context, userID uint64, viewerUserID uint64) (social.PublicProfile, error) {
	key := publicProfileCacheKey(userID, viewerUserID)
	value, err := r.redis.Get(ctx, key).Result()
	if err == nil {
		var profile social.PublicProfile
		if decodeErr := json.Unmarshal([]byte(value), &profile); decodeErr == nil {
			return profile, nil
		}
		_ = r.redis.Del(ctx, key).Err()
	}

	profile, err := r.next.GetPublicProfile(ctx, userID, viewerUserID)
	if err != nil {
		return social.PublicProfile{}, err
	}
	payload, err := json.Marshal(profile)
	if err == nil {
		_ = r.redis.Set(ctx, key, payload, r.profileTTL).Err()
	}
	return profile, nil
}

func (r *CachedRepository) Follow(ctx context.Context, followerID uint64, followingID uint64) (bool, error) {
	created, err := r.next.Follow(ctx, followerID, followingID)
	if err != nil {
		return false, err
	}
	r.invalidateProfiles(ctx, followerID, followingID)
	return created, nil
}

func (r *CachedRepository) Unfollow(ctx context.Context, followerID uint64, followingID uint64) error {
	if err := r.next.Unfollow(ctx, followerID, followingID); err != nil {
		return err
	}
	r.invalidateProfiles(ctx, followerID, followingID)
	return nil
}

func (r *CachedRepository) Block(ctx context.Context, blockerID uint64, blockedID uint64) error {
	if err := r.next.Block(ctx, blockerID, blockedID); err != nil {
		return err
	}
	r.invalidateProfiles(ctx, blockerID, blockedID)
	return nil
}

func (r *CachedRepository) Unblock(ctx context.Context, blockerID uint64, blockedID uint64) error {
	if err := r.next.Unblock(ctx, blockerID, blockedID); err != nil {
		return err
	}
	r.invalidateProfiles(ctx, blockerID, blockedID)
	return nil
}

func (r *CachedRepository) ListFollowerIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	return r.next.ListFollowerIDs(ctx, userID)
}

func (r *CachedRepository) InvalidatePublicProfile(ctx context.Context, userID uint64) {
	r.invalidateProfiles(ctx, userID)
}

func (r *CachedRepository) invalidateProfiles(ctx context.Context, userIDs ...uint64) {
	for _, userID := range userIDs {
		if userID == 0 {
			continue
		}
		iter := r.redis.Scan(ctx, 0, publicProfileCachePattern(userID), 0).Iterator()
		for iter.Next(ctx) {
			_ = r.redis.Del(ctx, iter.Val()).Err()
		}
		if iter.Err() != nil {
			_ = r.redis.Del(ctx, publicProfileCacheKey(userID, 0)).Err()
		}
	}
}

func publicProfileCacheKey(userID uint64, viewerUserID uint64) string {
	return fmt.Sprintf("social:profile:v2:%d:%d", userID, viewerUserID)
}

func publicProfileCachePattern(userID uint64) string {
	return fmt.Sprintf("social:profile:v2:%d:*", userID)
}

var _ social.Repository = (*CachedRepository)(nil)
