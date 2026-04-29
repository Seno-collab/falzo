package share

import (
	httpResponse "falzo-be/pkg/response"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/rs/zerolog/log"
)

type ApiError struct {
	Status  int
	Message string
	Code    string
	Field   string
	Detail  string
	LogErr  bool
}

func WriteError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	operation string,
	mapper func(error) ApiError,
) {
	mapped := mapper(err)

	if mapped.LogErr {
		log.Error().
			Err(err).
			Str("operation", operation).
			Str("request_id", chimiddleware.GetReqID(r.Context())).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("request failed")
	}

	httpResponse.Error(w, mapped.Status, mapped.Message, r, httpResponse.ErrorDetail{
		Code:    mapped.Code,
		Field:   mapped.Field,
		Message: mapped.Detail,
	})

}

func BadRequest(field, detail string) ApiError {
	return ApiError{
		Status:  http.StatusBadRequest,
		Message: ValidationField,
		Code:    INVALID_FIELD,
		Field:   field,
		Detail:  detail,
	}
}

func Required(field, detail string) ApiError {
	return ApiError{
		Status:  http.StatusBadRequest,
		Message: ValidationField,
		Code:    REQUIRED_FIELD,
		Field:   field,
		Detail:  detail,
	}
}

func NotFound(msg, detail string) ApiError {
	return ApiError{
		Status:  http.StatusNotFound,
		Message: msg,
		Code:    NOT_FOUND,
		Detail:  detail,
	}
}

func Internal() ApiError {
	return ApiError{
		Status:  http.StatusInternalServerError,
		Message: InternalServerError,
		Code:    INTERNAL_ERROR,
		Detail:  UnexpectedError,
		LogErr:  true,
	}
}

func ServiceUnavailable(message, detail string) ApiError {
	return ApiError{
		Status:  http.StatusServiceUnavailable,
		Message: message,
		Code:    SERVICE_UNAVAILABLE,
		Detail:  detail,
		LogErr:  true,
	}
}

func TooManyRequests(detail string) ApiError {
	return ApiError{
		Status:  http.StatusTooManyRequests,
		Message: "Too many requests",
		Code:    RATE_LIMITED,
		Detail:  detail,
	}
}

func UnauthorizedCredentials(message, detail string) ApiError {
	if message == "" {
		message = Unauthorized
	}
	return ApiError{
		Status:  http.StatusUnauthorized,
		Message: message,
		Code:    UNAUTHORIZED,
		Detail:  detail,
	}
}
