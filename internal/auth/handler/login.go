package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"falzo/internal/auth"
	httpresponse "falzo/pkg/http/response"
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

	token, err := h.service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, auth.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
		}
		if errors.Is(err, auth.ErrAuthUnavailable) {
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
