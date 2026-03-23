package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/domain"
	httpresponse "falzo-be/pkg/response"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if !h.protector.allowOperation(now) {
		writeAuthError(w, r, domain.ErrAuthTemporarilyUnavailable, "refresh")
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpresponse.ErrorDetail{
			Code:    "INVALID_FORMAT",
			Message: "Invalid JSON payload",
		})
		return
	}

	if req.RefreshToken == "" {
		httpresponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpresponse.ErrorDetail{
			Code:    "REQUIRED_FIELD",
			Field:   "refresh_token",
			Message: "Refresh token is required",
		})
		return
	}

	key := authClientIP(r)
	if !h.protector.refreshLimiter.allow(key, now) {
		httpresponse.Error(w, http.StatusTooManyRequests, "Too many requests", r, httpresponse.ErrorDetail{
			Code:    "RATE_LIMITED",
			Message: "Too many refresh attempts, please try again later",
		})
		return
	}

	tokens, err := h.service.Refresh(r.Context(), command.Refresh{RefreshToken: req.RefreshToken})
	h.protector.observe(err, time.Now())
	if err != nil {
		writeAuthError(w, r, err, "refresh")
		return
	}

	httpresponse.Success(w, http.StatusOK, "Token refreshed successfully", tokens, r)
}
