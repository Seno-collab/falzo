package httpapi

import (
	"net/http"
	"strings"

	"falzo-be/internal/auth/application/command"
	httpresponse "falzo-be/pkg/response"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || token == "" {
		httpresponse.Error(w, http.StatusUnauthorized, "Unauthorized", r, httpresponse.ErrorDetail{
			Code:    "UNAUTHORIZED",
			Message: "Invalid authorization header",
		})
		return
	}

	if err := h.service.Logout(r.Context(), command.Logout{Token: token}); err != nil {
		writeAuthError(w, r, err, "logout")
		return
	}

	httpresponse.Success(w, http.StatusOK, "Logout acknowledged", map[string]string{
		"message": "logout acknowledged",
	}, r)
}
