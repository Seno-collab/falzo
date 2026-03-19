package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	httpresponse "falzo/internal/http/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Chi
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func NewHandler(service Chi) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Group(func(protected chi.Router) {
		protected.Use(RequireAuth(h.service))
		protected.Get("/me", h.Me)
		protected.Post("/logout", h.Logout)
	})
	return r
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", "invalid json payload")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", "username, email and password are required")
		return
	}

	err := h.service.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrUserExists) {
			status = http.StatusConflict
		}
		if errors.Is(err, ErrAuthUnavailable) {
			status = http.StatusServiceUnavailable
		}
		httpresponse.Error(w, status, "register failed", err.Error())
		return
	}

	httpresponse.JSON(w, http.StatusCreated, map[string]string{
		"message": "account created",
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", "invalid json payload")
		return
	}

	if req.Username == "" || req.Password == "" {
		httpresponse.Error(w, http.StatusBadRequest, "invalid request", "username and password are required")
		return
	}

	token, err := h.service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		if errors.Is(err, ErrAuthUnavailable) {
			status = http.StatusServiceUnavailable
		}
		httpresponse.Error(w, status, "login failed", err.Error())
		return
	}

	httpresponse.JSON(w, http.StatusOK, LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, http.StatusOK, map[string]string{
		"message": "logout acknowledged",
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := ClaimsFromContext(r.Context())
	if !ok {
		httpresponse.Error(w, http.StatusUnauthorized, "unauthorized", "missing auth context")
		return
	}

	httpresponse.JSON(w, http.StatusOK, map[string]any{
		"username": claims.Username,
		"subject":  claims.Subject,
		"expires":  claims.ExpiresAt,
	})
}
