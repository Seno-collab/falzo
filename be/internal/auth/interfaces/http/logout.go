package httpapi

import (
	"net/http"

	httpresponse "falzo-be/pkg/response"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	httpresponse.JSON(w, http.StatusOK, map[string]string{
		"message": "logout acknowledged",
	})
}
