package application

import (
	"context"
	"errors"

	"falzo-be/internal/location/application/query"
	"falzo-be/internal/location/domain"
	"falzo-be/internal/location/domain/entity"
)

var ErrRadiusMustBePositive = errors.New("radius must be greater than 0")
var ErrLatitudeOutOfRange = errors.New("latitude must be between -90 and 90")
var ErrLongitudeOutOfRange = errors.New("longitude must be between -180 and 180")

func (s *service) Nearby(ctx context.Context, input query.NearbyLocation) ([]entity.NearbyLocation, error) {
	if s.locations == nil {
		return nil, domain.ErrLocationDependencyUnavailable
	}

	if input.Latitude < -90 || input.Latitude > 90 {
		return nil, ErrLatitudeOutOfRange
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		return nil, ErrLongitudeOutOfRange
	}
	if input.RadiusMeters <= 0 {
		return nil, ErrRadiusMustBePositive
	}

	return s.locations.Nearby(ctx, input.Latitude, input.Longitude, input.RadiusMeters)
}
