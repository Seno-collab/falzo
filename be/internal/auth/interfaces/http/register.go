package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/domain"
	httpresponse "falzo-be/pkg/response"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
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

	err := h.service.Register(r.Context(), command.Register{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrUserExists) {
			status = http.StatusConflict
		}
		if errors.Is(err, domain.ErrAuthUnavailable) {
			status = http.StatusServiceUnavailable
		}
		httpresponse.Error(w, status, "register failed", err.Error())
		return
	}

	httpresponse.Success(w, map[string]string{
		"message": "account created",
	})
}
