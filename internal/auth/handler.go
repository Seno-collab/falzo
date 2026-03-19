package auth

import (
	"net/http"

	httpresponse "falzo/internal/http/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	return r
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, http.StatusNotImplemented, map[string]string{
		"message": "login not implemented",
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, http.StatusNotImplemented, map[string]string{
		"message": "logout not implemented",
	})
}
