package httpapi

import (
	"net/http"

	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"
)

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := authenticatedUserFromContext(r.Context())
	if !ok {
		httpResponse.Error(w, http.StatusUnauthorized, "Unauthorized", r, httpResponse.ErrorDetail{
			Code:    share.UNAUTHORIZED,
			Message: "Missing auth context",
		})
		return
	}

	httpResponse.Success(w, http.StatusOK, "Authenticated user fetched successfully", map[string]any{
		"user_name": claims.Username,
		"subject":   claims.Subject,
		"expires":   claims.ExpiresAt,
	}, r)
}
