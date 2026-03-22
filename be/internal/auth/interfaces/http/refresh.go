package httpapi

import (
	"encoding/json"
	"net/http"

	"falzo-be/internal/auth/application/command"
	httpresponse "falzo-be/pkg/response"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
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

	tokens, err := h.service.Refresh(r.Context(), command.Refresh{RefreshToken: req.RefreshToken})
	if err != nil {
		writeAuthError(w, r, err, "refresh")
		return
	}

	httpresponse.Success(w, http.StatusOK, "Token refreshed successfully", tokens, r)
}
