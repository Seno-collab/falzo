package domain

import "errors"

var ErrLocationDependencyUnavailable = errors.New("location dependency unavailable")
var ErrLocationInternal = errors.New("location internal error")
