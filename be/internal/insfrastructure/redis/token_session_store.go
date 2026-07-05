package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type TokenSessionStore struct {
	client *goredis.Client
	prefix string
}

func NewTokenSessionStore(client *goredis.Client, prefix string) *TokenSessionStore {
	return &TokenSessionStore{client: client, prefix: prefix}
}

func (s *TokenSessionStore) SaveRefresh(ctx context.Context, tokenID string, userID int64, ttl time.Duration) error {
	return s.save(ctx, "refresh", tokenID, userID, ttl)
}
func (s *TokenSessionStore) ConsumeRefresh(ctx context.Context, tokenID string) (bool, error) {
	return s.consume(ctx, "refresh", tokenID)
}
func (s *TokenSessionStore) SavePasswordReset(ctx context.Context, tokenID string, userID int64, ttl time.Duration) error {
	return s.save(ctx, "password-reset", tokenID, userID, ttl)
}
func (s *TokenSessionStore) ConsumePasswordReset(ctx context.Context, tokenID string) (bool, error) {
	return s.consume(ctx, "password-reset", tokenID)
}
func (s *TokenSessionStore) save(ctx context.Context, kind, tokenID string, userID int64, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("cannot store expired %s token", kind)
	}
	return s.client.Set(ctx, s.key(kind, tokenID), userID, ttl).Err()
}
func (s *TokenSessionStore) consume(ctx context.Context, kind, tokenID string) (bool, error) {
	_, err := s.client.GetDel(ctx, s.key(kind, tokenID)).Result()
	if err == goredis.Nil {
		return false, nil
	}
	return err == nil, err
}
func (s *TokenSessionStore) key(kind, tokenID string) string {
	return s.prefix + ":auth:" + kind + ":" + tokenID
}
