package domainuser

import (
	"time"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusLocked   UserStatus = "LOCKED"
	UserStatusDisabled UserStatus = "DISABLED"
)

type User struct {
	ID             int64
	UserName       string
	Email          string
	PasswordHash   string
	FailedAttempts int
	Status         UserStatus
	LockUntil      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (u *User) CanLogin(now time.Time) error {
	if u.Status == UserStatusDisabled {
		return ErrUserDisabled
	}
	if u.Status == UserStatusLocked {
		if u.LockUntil != nil && u.LockUntil.After(now) {
			return ErrUserLocked
		}
		u.RecordSuccessfulLogin(now)
	}
	return nil
}

func (u *User) RecordFailedLogin(now time.Time, maxAttempts int, lockDuration time.Duration) {
	u.FailedAttempts++
	u.UpdatedAt = now
	if u.FailedAttempts >= maxAttempts {
		lockedUnitl := now.Add(lockDuration)
		u.LockUntil = &lockedUnitl
		u.Status = UserStatusLocked
	}
}

func (u *User) RecordSuccessfulLogin(now time.Time) {
	u.FailedAttempts = 0
	u.LockUntil = nil
	u.Status = UserStatusActive
	u.UpdatedAt = now
}
