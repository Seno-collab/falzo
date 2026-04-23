package application

import (
	"context"
	"errors"

	"falzo-be/internal/post/application/query"
	"falzo-be/internal/post/domain"
)

var ErrPageMustBePositive = errors.New("page must be greater than 0")
var ErrLimitMustBePositive = errors.New("limit must be greater than 0")

func (s *service) GetPosts(ctx context.Context, input query.GetPosts) ([]query.Post, error) {
	if s.posts == nil {
		return nil, domain.ErrPostDependencyUnavailable
	}
	if input.Page <= 0 {
		return nil, ErrPageMustBePositive
	}
	if input.Limit <= 0 {
		return nil, ErrLimitMustBePositive
	}

	items, err := s.posts.GetPosts(ctx, input.Page, input.Limit)
	if err != nil {
		return nil, err
	}

	posts := make([]query.Post, 0, len(items))
	for _, item := range items {
		posts = append(posts, mapPostEntity(item))
	}

	return posts, nil
}
