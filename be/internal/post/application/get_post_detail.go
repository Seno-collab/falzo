package application

import (
	"context"

	"falzo-be/internal/post/application/query"
	"falzo-be/internal/post/domain"
)

func (s *service) GetPostDetail(ctx context.Context, input query.GetPostDetail) (*query.Post, error) {
	if s.posts == nil {
		return nil, domain.ErrPostDependencyUnavailable
	}
	if input.PostID == 0 {
		return nil, ErrPostIDRequired
	}

	item, err := s.posts.GetPostDetail(ctx, input.PostID)
	if err != nil {
		return nil, err
	}

	mapped := mapPostEntity(*item)
	return &mapped, nil
}
