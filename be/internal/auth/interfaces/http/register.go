package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/domain"
	httpResponse "falzo-be/pkg/response"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if !h.protector.allowOperation(now) {
		writeAuthError(w, r, domain.ErrAuthTemporarilyUnavailable, "register")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpResponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpResponse.ErrorDetail{
			Code:    "INVALID_FORMAT",
			Message: "Invalid JSON payload",
		})
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		httpResponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpResponse.ErrorDetail{
			Code:    "REQUIRED_FIELD",
			Message: "Username, email and password are required",
		})
		return
	}

	if !h.protector.registerLimiter.allow(authClientIP(r), now) {
		httpResponse.Error(w, http.StatusTooManyRequests, "Too many requests", r, httpResponse.ErrorDetail{
			Code:    "RATE_LIMITED",
			Message: "Too many registration attempts, please try again later",
		})
		return
	}

	err := h.service.Register(r.Context(), command.Register{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	h.protector.observe(err, time.Now())
	if err != nil {
		writeAuthError(w, r, err, "register")
		return
	}

	httpResponse.Success(w, http.StatusCreated, "Account created successfully", map[string]string{
		"message": "account created",
	}, r)
}
