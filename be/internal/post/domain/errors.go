package domain

import "errors"

var ErrPostNotFound = errors.New("post not found")
var ErrPostDependencyUnavailable = errors.New("post dependency unavailable")
var ErrPostInternal = errors.New("post internal error")
