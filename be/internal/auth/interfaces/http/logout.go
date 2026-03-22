package httpapi

import (
	"net/http"

	httpresponse "falzo-be/pkg/response"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	httpresponse.Success(w, http.StatusOK, "Logout acknowledged", map[string]string{
		"message": "logout acknowledged",
	}, r)
}
