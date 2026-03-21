package response

import (
	"encoding/json"
	"net/http"
)

type ErrorPayload struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, status int, message string, err string) {
	JSON(w, status, ErrorPayload{
		Message: message,
		Error:   err,
	})
}

func Success(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}
