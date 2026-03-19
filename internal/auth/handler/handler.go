package handler

import (
	"falzo/internal/auth"
	authmiddleware "falzo/internal/auth/middleware"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service auth.Service
}

func New(service auth.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Group(func(protected chi.Router) {
		protected.Use(authmiddleware.RequireAuth(h.service))
		protected.Get("/me", h.Me)
		protected.Post("/logout", h.Logout)
	})
	return r
}
