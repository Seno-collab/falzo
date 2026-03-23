package httpapi

import (
	"errors"
	"net/http"

	"falzo-be/internal/auth/domain"
	httpresponse "falzo-be/pkg/response"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type apiError struct {
	status  int
	message string
	code    string
	detail  string
	logErr  bool
}

func writeAuthError(w http.ResponseWriter, r *http.Request, err error, operation string) {
	mapped := mapAuthError(err)
	if mapped.logErr {
		log.Error().
			Err(err).
			Str("operation", operation).
			Str("request_id", chimiddleware.GetReqID(r.Context())).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("auth request failed")
	}

	httpresponse.Error(w, mapped.status, mapped.message, r, httpresponse.ErrorDetail{
		Code:    mapped.code,
		Message: mapped.detail,
	})
}

func mapAuthError(err error) apiError {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return apiError{
			status:  http.StatusUnauthorized,
			message: "Invalid credentials",
			code:    "UNAUTHORIZED",
			detail:  "Username or password is incorrect",
		}
	case errors.Is(err, domain.ErrInvalidToken):
		return apiError{
			status:  http.StatusUnauthorized,
			message: "Unauthorized",
			code:    "UNAUTHORIZED",
			detail:  "Token is invalid",
		}
	case errors.Is(err, domain.ErrSessionRevoked):
		return apiError{
			status:  http.StatusUnauthorized,
			message: "Unauthorized",
			code:    "UNAUTHORIZED",
			detail:  "Session has been revoked or expired",
		}
	case errors.Is(err, domain.ErrUserExists):
		return apiError{
			status:  http.StatusConflict,
			message: "Account already exists",
			code:    "ALREADY_EXISTS",
			detail:  "Username or email is already in use",
		}
	case errors.Is(err, domain.ErrAuthDependencyUnavailable):
		return apiError{
			status:  http.StatusServiceUnavailable,
			message: "Authentication service unavailable",
			code:    "SERVICE_UNAVAILABLE",
			detail:  "Authentication service is temporarily unavailable",
			logErr:  true,
		}
	case errors.Is(err, domain.ErrAuthTemporarilyUnavailable):
		return apiError{
			status:  http.StatusServiceUnavailable,
			message: "Authentication service unavailable",
			code:    "SERVICE_UNAVAILABLE",
			detail:  "Authentication service is temporarily unavailable",
		}
	case errors.Is(err, domain.ErrAuthInternal):
		return apiError{
			status:  http.StatusInternalServerError,
			message: "Internal server error",
			code:    "INTERNAL_ERROR",
			detail:  "An unexpected error occurred",
			logErr:  true,
		}
	default:
		return apiError{
			status:  http.StatusInternalServerError,
			message: "Internal server error",
			code:    "INTERNAL_ERROR",
			detail:  "An unexpected error occurred",
			logErr:  true,
		}
	}
}
