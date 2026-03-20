package domain

import "errors"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid token")
var ErrUserExists = errors.New("user already exists")
var ErrAuthUnavailable = errors.New("auth database unavailable")
