package request

import (
	"be/internal/shared/apperror"
	"be/internal/shared/validatorx"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxBodySize = 1 << 20 // 1MB
func DecodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var v T

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&v); err != nil {
		var zero T

		if errors.Is(err, io.EOF) {
			return zero, apperror.InvalidRequest("Request body is required")
		}

		return zero, apperror.InvalidRequest("Invalid JSON body")
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		var zero T
		return zero, apperror.InvalidRequest("Request body must contain only one JSON object")
	}

	if validationErrors := validatorx.Validate(v); len(validationErrors) > 0 {
		var zero T
		return zero, apperror.ValidationError(validationErrors)
	}

	return v, nil
}
