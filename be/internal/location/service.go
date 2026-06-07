package location

import (
	"context"
	"strings"
)

type Service struct {
	locations Repository
}

func NewService(locations Repository) *Service {
	return &Service{locations: locations}
}

type SearchInput struct {
	Query string
}

type NearbyInput struct {
	Latitude     float64
	Longitude    float64
	RadiusMeters float64
}

type GetPostsByLocationInput struct {
	LocationID string
}

type GetPlaceBySlugInput struct {
	Slug string
}

func (s *Service) Search(ctx context.Context, input SearchInput) ([]Location, error) {
	if s.locations == nil {
		return nil, ErrDependencyUnavailable
	}

	searchQuery := strings.TrimSpace(input.Query)
	if searchQuery == "" {
		return nil, ErrSearchQueryRequired
	}

	return s.locations.Search(ctx, searchQuery)
}

func (s *Service) Nearby(ctx context.Context, input NearbyInput) ([]NearbyLocation, error) {
	if s.locations == nil {
		return nil, ErrDependencyUnavailable
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

func (s *Service) GetPostsByLocation(ctx context.Context, input GetPostsByLocationInput) ([]LocationPost, error) {
	if s.locations == nil {
		return nil, ErrDependencyUnavailable
	}

	locationID := strings.TrimSpace(input.LocationID)
	if locationID == "" {
		return nil, ErrLocationIDRequired
	}

	return s.locations.GetPostsByLocationID(ctx, locationID)
}

func (s *Service) GetPlaceBySlug(ctx context.Context, input GetPlaceBySlugInput) (PlaceDetail, error) {
	if s.locations == nil {
		return PlaceDetail{}, ErrDependencyUnavailable
	}

	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		return PlaceDetail{}, ErrPlaceSlugRequired
	}

	return s.locations.FindPlaceBySlug(ctx, slug)
}
