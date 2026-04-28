package entity

import (
	"time"

	"falzo-be/internal/post/domain/value_object"
)

type Post struct {
	ID           uint64
	UserID       uint64
	ImageURL     value_object.ImageURL
	Caption      value_object.Caption
	LocationName value_object.LocationName
	Latitude     float64
	Longitude    float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
