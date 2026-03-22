package httpapi

import (
	"net/http"

	httpresponse "falzo-be/pkg/response"
)

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := authenticatedUserFromContext(r.Context())
	if !ok {
		httpresponse.Error(w, http.StatusUnauthorized, "Unauthorized", r, httpresponse.ErrorDetail{
			Code:    "UNAUTHORIZED",
			Message: "Missing auth context",
		})
		return
	}

	httpresponse.Success(w, http.StatusOK, "Authenticated user fetched successfully", map[string]any{
		"username": claims.Username,
		"subject":  claims.Subject,
		"expires":  claims.ExpiresAt,
	}, r)
}
