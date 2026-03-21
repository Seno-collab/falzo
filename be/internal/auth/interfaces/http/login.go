package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/domain"
	httpresponse "falzo-be/pkg/response"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
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

	token, err := h.service.Login(r.Context(), command.Login{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		if errors.Is(err, domain.ErrAuthUnavailable) {
			status = http.StatusServiceUnavailable
		}
		httpresponse.Error(w, status, "login failed", err.Error())
		return
	}

	httpresponse.Success(w, LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	})
}
