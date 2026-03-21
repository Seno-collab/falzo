package httpapi

import (
	"net/http"

	authhttp "falzo-be/internal/auth/infrastructure/http"
	httpresponse "falzo-be/pkg/response"
)

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := authhttp.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpresponse.Error(w, http.StatusUnauthorized, "unauthorized", "missing auth context")
		return
	}

	httpresponse.Success(w, map[string]any{
		"username": claims.Username,
		"subject":  claims.Subject,
		"expires":  claims.ExpiresAt,
	})
}
