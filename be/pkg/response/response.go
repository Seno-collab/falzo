package response

import (
	"encoding/json"
	"falzo-be/internal/i18n"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Meta struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

type ErrorDetail struct {
	Field      string `json:"field,omitempty"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	MessageKey string `json:"message_key,omitempty"`
	Debug      *Debug `json:"debug,omitempty"`
}

type Debug struct {
	Module       string            `json:"module,omitempty"`
	Operation    string            `json:"operation,omitempty"`
	SourceFile   string            `json:"source_file,omitempty"`
	SourceLine   int               `json:"source_line,omitempty"`
	SourceFunc   string            `json:"source_func,omitempty"`
	AppCode      string            `json:"app_code,omitempty"`
	AppOperation string            `json:"app_operation,omitempty"`
	AppMetadata  map[string]string `json:"app_metadata,omitempty"`
	RootError    string            `json:"root_error,omitempty"`
}

type Envelope struct {
	Success    bool          `json:"success"`
	Message    string        `json:"message"`
	MessageKey string        `json:"message_key,omitempty"`
	Data       any           `json:"data"`
	Errors     []ErrorDetail `json:"errors,omitempty"`
	Meta       Meta          `json:"meta"`
}

func JSON(w http.ResponseWriter, status int, payload Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func Success(w http.ResponseWriter, status int, message string, data any, r *http.Request) {
	translation := i18n.ResolveRequest(r, message)
	JSON(w, status, Envelope{
		Success:    true,
		Message:    translation.Value,
		MessageKey: translation.Key,
		Data:       data,
		Meta:       buildMeta(r),
	})
}

func Error(w http.ResponseWriter, status int, message string, r *http.Request, errors ...ErrorDetail) {
	translation := i18n.ResolveRequest(r, message)
	localizedErrors := make([]ErrorDetail, 0, len(errors))
	for _, detail := range errors {
		detailTranslation := i18n.ResolveRequest(r, detail.Message)
		detail.Message = detailTranslation.Value
		detail.MessageKey = detailTranslation.Key
		localizedErrors = append(localizedErrors, detail)
	}

	JSON(w, status, Envelope{
		Success:    false,
		Message:    translation.Value,
		MessageKey: translation.Key,
		Data:       nil,
		Errors:     localizedErrors,
		Meta:       buildMeta(r),
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
