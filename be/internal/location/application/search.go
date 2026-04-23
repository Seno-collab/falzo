package application

import (
	"context"
	"errors"
	"strings"

	"falzo-be/internal/location/application/query"
	"falzo-be/internal/location/domain"
	"falzo-be/internal/location/domain/entity"
)

var ErrSearchQueryRequired = errors.New("search query is required")

func (s *service) Search(ctx context.Context, input query.SearchLocation) ([]entity.Location, error) {
	if s.locations == nil {
		return nil, domain.ErrLocationDependencyUnavailable
	}

	searchQuery := strings.TrimSpace(input.Query)
	if searchQuery == "" {
		return nil, ErrSearchQueryRequired
	}

	return s.locations.Search(ctx, searchQuery)
}
