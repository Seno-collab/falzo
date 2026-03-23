package httpapi

import (
	"net/http"
	"time"
)

type Option func(*Handler)

func WithPublicMiddlewares(middlewares ...func(http.Handler) http.Handler) Option {
	return func(h *Handler) {
		h.publicMiddlewares = append(h.publicMiddlewares, middlewares...)
	}
}

func WithProtector(protector *authProtector) Option {
	return func(h *Handler) {
		if protector != nil {
			h.protector = protector
		}
	}
}

func WithProtectorConfig(limitPerMinute int, failureThreshold int, cooldown time.Duration) Option {
	return WithProtector(newAuthProtector(limitPerMinute, failureThreshold, cooldown))
}
