package httpapi

import (
	"net/http"

	authhttp "falzo/internal/auth/infrastructure/http"
	httpresponse "falzo/pkg/response"
)

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := authhttp.AuthenticatedUserFromContext(r.Context())
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
