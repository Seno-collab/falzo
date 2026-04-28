package repository

import (
	"context"

	"falzo-be/internal/post/domain/aggregate"
	"falzo-be/internal/post/domain/entity"
	"falzo-be/internal/post/domain/value_object"
)

type PostRepository interface {
	Create(ctx context.Context, post *aggregate.Post) error
	Like(ctx context.Context, postID uint64, userID uint64) error
	Save(ctx context.Context, postID uint64, userID uint64) error
	GetPosts(ctx context.Context, page int, limit int) ([]entity.Post, error)
	GetPostDetail(ctx context.Context, postID uint64) (*entity.Post, error)
	GetPostsByLocation(ctx context.Context, locationName value_object.LocationName) ([]entity.Post, error)
}
