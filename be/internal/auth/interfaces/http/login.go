package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/domain"
	httpresponse "falzo-be/pkg/response"
)

type LoginRequest struct {
	Username string `json:"username"`
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
		httpresponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpresponse.ErrorDetail{
			Code:    "INVALID_FORMAT",
			Message: "Invalid JSON payload",
		})
		return
	}

	if req.Username == "" || req.Password == "" {
		httpresponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpresponse.ErrorDetail{
			Code:    "REQUIRED_FIELD",
			Message: "Username and password are required",
		})
		return
	}

	key := authClientIP(r) + ":" + strings.ToLower(req.Username)
	if !h.protector.loginLimiter.allow(key, now) {
		httpresponse.Error(w, http.StatusTooManyRequests, "Too many requests", r, httpresponse.ErrorDetail{
			Code:    "RATE_LIMITED",
			Message: "Too many login attempts, please try again later",
		})
		return
	}

	tokens, err := h.service.Login(r.Context(), command.Login{
		Username: req.Username,
		Password: req.Password,
	})
	h.protector.observe(err, time.Now())
	if err != nil {
		writeAuthError(w, r, err, "login")
		return
	}

	httpresponse.Success(w, http.StatusOK, "Login successful", tokens, r)
}
