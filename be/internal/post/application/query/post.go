package query

import "time"

type Post struct {
	ID           uint64    `json:"id"`
	UserID       uint64    `json:"user_id"`
	ImageURL     string    `json:"image_url"`
	Caption      string    `json:"caption"`
	LocationName string    `json:"location_name"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	CreatedAt    time.Time `json:"created_at"`
}
