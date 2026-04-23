package application

import (
	"context"
	"errors"
	"strings"

	"falzo-be/internal/location/application/query"
	"falzo-be/internal/location/domain"
	"falzo-be/internal/location/domain/entity"
)

var ErrLocationIDRequired = errors.New("location id is required")

func (s *service) GetPostsByLocation(ctx context.Context, input query.GetPostsByLocation) ([]entity.LocationPost, error) {
	if s.locations == nil {
		return nil, domain.ErrLocationDependencyUnavailable
	}

	locationID := strings.TrimSpace(input.LocationID)
	if locationID == "" {
		return nil, ErrLocationIDRequired
	}

	return s.locations.GetPostsByLocationID(ctx, locationID)
}
