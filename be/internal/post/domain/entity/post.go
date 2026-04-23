package entity

import (
	"time"

	"falzo-be/internal/post/domain/valueobject"
)

type Post struct {
	ID           uint64
	UserID       uint64
	ImageURL     valueobject.ImageURL
	Caption      valueobject.Caption
	LocationName valueobject.LocationName
	Latitude     float64
	Longitude    float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
