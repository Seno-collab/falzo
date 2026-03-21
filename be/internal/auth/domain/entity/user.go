package entity

import (
	"time"

	"falzo-be/internal/auth/domain/valueobject"
)

type User struct {
	ID           uint64
	Username     valueobject.Username
	Email        valueobject.Email
	PasswordHash valueobject.PasswordHash
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
