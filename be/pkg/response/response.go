package response

import (
	"encoding/json"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Meta struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Envelope struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    any           `json:"data"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
	Meta    Meta          `json:"meta"`
}

func JSON(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Success(w http.ResponseWriter, status int, message string, data any, r *http.Request) {
	JSON(w, status, Envelope{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    buildMeta(r),
	})
}

func Error(w http.ResponseWriter, status int, message string, r *http.Request, errors ...ErrorDetail) {
	JSON(w, status, Envelope{
		Success: false,
		Message: message,
		Data:    nil,
		Errors:  errors,
		Meta:    buildMeta(r),
	})
}

func buildMeta(r *http.Request) Meta {
	meta := Meta{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if r == nil {
		return meta
	}

	meta.RequestID = chimiddleware.GetReqID(r.Context())
	return meta
}
