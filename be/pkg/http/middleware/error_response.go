package middleware

import (
	"errors"

	"falzo-be/internal/share"
)

var (
	errRateLimited  = errors.New("rate limited")
	errRequestPanic = errors.New("request panicked")
)

func mapRateLimitError(error) share.ApiError {
	return share.TooManyRequests("Too many requests, please try again later")
}

func mapRecoverError(error) share.ApiError {
	return share.Internal()
}
