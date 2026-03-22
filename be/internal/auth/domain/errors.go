package domain

import "errors"

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid token")
var ErrSessionRevoked = errors.New("session revoked or expired")
var ErrUserExists = errors.New("user already exists")
var ErrAuthDependencyUnavailable = errors.New("auth dependency unavailable")
var ErrAuthInternal = errors.New("auth internal error")
