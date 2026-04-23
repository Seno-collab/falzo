package httpapi

import (
	"falzo-be/internal/post/application"

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
	r.Post("/", h.CreatePost)
	r.Post("/{id}/like", h.LikePost)
	r.Post("/{id}/save", h.SavePost)
	r.Get("/", h.GetPosts)
	r.Get("/location", h.GetPostsByLocation)
	r.Get("/{id}", h.GetPostDetail)
	return r
}
