package query

import "time"

type AuthenticatedUser struct {
	UserID    uint64
	Username  string
	Subject   string
	SessionID string
	ExpiresAt *time.Time
}
