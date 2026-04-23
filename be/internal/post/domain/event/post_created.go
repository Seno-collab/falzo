package event

import "time"

type PostCreated struct {
	PostID        uint64
	UserID        uint64
	ImageURL      string
	LocationName  string
	OccurredAtUTC time.Time
}
