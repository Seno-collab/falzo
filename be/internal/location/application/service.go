package application

import (
	"context"

	"falzo-be/internal/location/application/query"
	"falzo-be/internal/location/domain/entity"
	"falzo-be/internal/location/domain/repository"
)

type Service interface {
	Search(ctx context.Context, input query.SearchLocation) ([]entity.Location, error)
	Nearby(ctx context.Context, input query.NearbyLocation) ([]entity.NearbyLocation, error)
	GetPostsByLocation(ctx context.Context, input query.GetPostsByLocation) ([]entity.LocationPost, error)
}

type service struct {
	locations repository.LocationRepository
}

func New(locations repository.LocationRepository) Service {
	return &service{
		locations: locations,
	}
}
