package share

import (
	"errors"
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
		entry := log.Error().
			Err(err).
			Str("operation", operation).
			Str("request_id", chimiddleware.GetReqID(r.Context())).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", mapped.Status).
			Str("api_code", mapped.Code).
			Str("api_detail", mapped.Detail)

		chain := errorChain(err)
		if len(chain) > 0 {
			entry = entry.Strs("error_chain", chain).Str("root_error", chain[len(chain)-1])
		}

		entry.Msg("request failed")
	}

	httpResponse.Error(w, mapped.Status, mapped.Message, r, httpResponse.ErrorDetail{
		Code:    mapped.Code,
		Field:   mapped.Field,
		Message: mapped.Detail,
	})

}

func errorChain(err error) []string {
	if err == nil {
		return nil
	}

	chain := make([]string, 0, 4)
	seen := map[string]struct{}{}
	collectErrorChain(err, &chain, seen, 8)
	return chain
}

func collectErrorChain(err error, chain *[]string, seen map[string]struct{}, remaining int) {
	if err == nil || remaining <= 0 {
		return
	}

	message := err.Error()
	if message != "" {
		if _, ok := seen[message]; !ok {
			seen[message] = struct{}{}
			*chain = append(*chain, message)
		}
	}

	var multi interface {
		Unwrap() []error
	}
	if errors.As(err, &multi) {
		for _, inner := range multi.Unwrap() {
			collectErrorChain(inner, chain, seen, remaining-1)
		}
		return
	}

	collectErrorChain(errors.Unwrap(err), chain, seen, remaining-1)
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
