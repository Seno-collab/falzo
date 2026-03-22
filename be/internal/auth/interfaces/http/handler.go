package httpapi

import (
	"falzo-be/internal/auth/application"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service application.Service
}

func New(service application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Group(func(protected chi.Router) {
		protected.Use(requireAuth(h.service))
		protected.Get("/me", h.Me)
		protected.Post("/logout", h.Logout)
	})
	return r
}
