package application

import (
	"context"
	"errors"
	"strings"

	"falzo-be/internal/post/application/query"
	"falzo-be/internal/post/domain"
	"falzo-be/internal/post/domain/valueobject"
)

var ErrLocationNameRequired = errors.New("location name is required")

func (s *service) GetPostsByLocation(ctx context.Context, input query.GetPostsByLocation) ([]query.Post, error) {
	if s.posts == nil {
		return nil, domain.ErrPostDependencyUnavailable
	}

	if strings.TrimSpace(input.LocationName) == "" {
		return nil, ErrLocationNameRequired
	}

	locationName, err := valueobject.NewLocationName(input.LocationName)
	if err != nil {
		return nil, err
	}

	items, err := s.posts.GetPostsByLocation(ctx, locationName)
	if err != nil {
		return nil, err
	}

	posts := make([]query.Post, 0, len(items))
	for _, item := range items {
		posts = append(posts, mapPostEntity(item))
	}

	return posts, nil
}
