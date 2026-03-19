package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"falzo/internal/auth"
	httpresponse "falzo/pkg/http/response"
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

	err := h.service.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, auth.ErrUserExists) {
			status = http.StatusConflict
		}
		if errors.Is(err, auth.ErrAuthUnavailable) {
			status = http.StatusServiceUnavailable
		}
		httpresponse.Error(w, status, "register failed", err.Error())
		return
	}

	httpresponse.JSON(w, http.StatusCreated, map[string]string{
		"message": "account created",
	})
}
