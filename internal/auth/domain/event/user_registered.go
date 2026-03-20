package event

import "time"

type UserRegistered struct {
	UserID     uint64
	Username   string
	Email      string
	OccurredAt time.Time
}
