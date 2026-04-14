package httpapi

import (
	"falzo-be/internal/auth/application"
	"falzo-be/pkg/config"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service           application.Service
	publicMiddlewares []func(http.Handler) http.Handler
	protector         *authProtector
}

func New(service application.Service, opts ...Option) *Handler {
	h := &Handler{
		service: service,
	}
	for _, opt := range opts {
		opt(h)
	}

	if h.protector == nil {
		cfg := config.Load()
		h.protector = newAuthProtector(
			cfg.Auth.RateLimitPerMin,
			cfg.Auth.DependencyFailureThreshold,
			cfg.Auth.DependencyCoolDown,
		)
	}

	return h
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(h.publicMiddlewares...).Post("/register", h.Register)
	r.With(h.publicMiddlewares...).Post("/login", h.Login)
	r.With(h.publicMiddlewares...).Post("/refresh", h.Refresh)
	r.Group(func(protected chi.Router) {
		protected.Use(requireAuth(h.service))
		protected.Get("/me", h.Me)
		protected.Post("/logout", h.Logout)
	})
	return r
}
