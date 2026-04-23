package repository

import (
	"context"

	"falzo-be/internal/location/domain/entity"
)

type LocationRepository interface {
	Search(ctx context.Context, query string) ([]entity.Location, error)
	Nearby(ctx context.Context, latitude, longitude, radiusMeters float64) ([]entity.NearbyLocation, error)
	GetPostsByLocationID(ctx context.Context, locationID string) ([]entity.LocationPost, error)
}
