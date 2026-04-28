package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/domain"
	"falzo-be/internal/auth/domain/value_object"
	httpResponse "falzo-be/pkg/response"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if !h.protector.allowOperation(now) {
		writeAuthError(w, r, domain.ErrAuthTemporarilyUnavailable, "login")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpResponse.Error(w, http.StatusBadRequest, "ValidationField", r, httpResponse.ErrorDetail{
			Code:    "INVALID_FORMAT",
			Message: "Invalid JSON payload",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		httpResponse.Error(w, http.StatusBadRequest, "ValidationField", r, httpResponse.ErrorDetail{
			Code:    "REQUIRED_FIELD",
			Message: "email and password are required",
		})
		return
	}

	if _, err := value_object.NewEmail(req.Email); err != nil {
		httpResponse.Error(w, http.StatusBadRequest, "ValidationField", r, httpResponse.ErrorDetail{
			Code:    "INVALID_FIELD",
			Message: "email must be a valid email",
		})
		return
	}

	key := authClientIP(r) + ":" + strings.ToLower(strings.TrimSpace(req.Email))
	if !h.protector.loginLimiter.allow(key, now) {
		httpResponse.Error(w, http.StatusTooManyRequests, "Too many requests", r, httpResponse.ErrorDetail{
			Code:    "RATE_LIMITED",
			Message: "Too many login attempts, please try again later",
		})
		return
	}

	tokens, err := h.service.Login(r.Context(), command.Login{
		Email:    req.Email,
		Password: req.Password,
	})
	h.protector.observe(err, time.Now())
	if err != nil {
		writeAuthError(w, r, err, "login")
		return
	}

	httpResponse.Success(w, http.StatusOK, "Login successful", tokens, r)
}
