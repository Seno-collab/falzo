package domainuser

import "errors"

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrUserLocked                = errors.New("user is locked")
	ErrUserDisabled              = errors.New("user is disabled")
	ErrInvalidPassword           = errors.New("invalid password")
	ErrInvalidUsernameOrPassword = errors.New("invalid user name or password")
	ErrUserNameAlreadyExists     = errors.New("username already exists")
	ErrInvalidToken              = errors.New("invalid or expired token")
)
