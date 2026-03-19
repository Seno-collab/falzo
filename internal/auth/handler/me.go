package handler

import (
	"net/http"

	"falzo/internal/auth"
	httpresponse "falzo/pkg/http/response"
)

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
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
