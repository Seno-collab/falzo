package location

import (
	"context"
	"errors"
)

var (
	ErrDependencyUnavailable = errors.New("location dependency unavailable")
	ErrInternal              = errors.New("location internal error")
	ErrSearchQueryRequired   = errors.New("search query is required")
	ErrLocationIDRequired    = errors.New("location id is required")
	ErrRadiusMustBePositive  = errors.New("radius must be greater than 0")
	ErrLatitudeOutOfRange    = errors.New("latitude must be between -90 and 90")
	ErrLongitudeOutOfRange   = errors.New("longitude must be between -180 and 180")
)

type Repository interface {
	Search(ctx context.Context, query string) ([]Location, error)
	Nearby(ctx context.Context, latitude, longitude, radiusMeters float64) ([]NearbyLocation, error)
	GetPostsByLocationID(ctx context.Context, locationID string) ([]LocationPost, error)
}

type Location struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type NearbyLocation struct {
	Location       Location `json:"location"`
	DistanceMeters float64  `json:"distance_meters"`
}

type LocationPost struct {
	ID           string  `json:"id"`
	UserID       uint64  `json:"user_id"`
	ImageURL     string  `json:"image_url"`
	Caption      string  `json:"caption"`
	LocationName string  `json:"location_name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}
