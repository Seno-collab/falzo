package httpapi

import (
	"net/http"
	"strings"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || token == "" {
		httpResponse.Error(w, http.StatusUnauthorized, "Unauthorized", r, httpResponse.ErrorDetail{
			Code:    share.UNAUTHORIZED,
			Message: "Invalid authorization header",
		})
		return
	}

	if err := h.service.Logout(r.Context(), command.Logout{Token: token}); err != nil {
		writeAuthError(w, r, err, "logout")
		return
	}

	httpResponse.Success(w, http.StatusOK, "Logout acknowledged", map[string]string{
		"message": "logout acknowledged",
	}, r)
}
