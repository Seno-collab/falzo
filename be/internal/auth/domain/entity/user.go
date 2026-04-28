package entity

import (
	"falzo-be/internal/auth/domain/value_object"
	"time"
)

type User struct {
	ID           uint64
	Username     value_object.Username
	Email        value_object.Email
	PasswordHash value_object.PasswordHash
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
