package httpapi

import (
	"errors"
	"net/http"

	"falzo-be/internal/post/application"
	"falzo-be/internal/post/domain"
	"falzo-be/internal/post/domain/value_object"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type apiError struct {
	status  int
	message string
	code    string
	field   string
	detail  string
	logErr  bool
}

func writePostError(w http.ResponseWriter, r *http.Request, err error, operation string) {
	mapped := mapPostError(err)
	if mapped.logErr {
		log.Error().
			Err(err).
			Str("operation", operation).
			Str("request_id", chimiddleware.GetReqID(r.Context())).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("post request failed")
	}

	httpResponse.Error(w, mapped.status, mapped.message, r, httpResponse.ErrorDetail{
		Code:    mapped.code,
		Field:   mapped.field,
		Message: mapped.detail,
	})
}

func mapPostError(err error) apiError {
	switch {
	case errors.Is(err, application.ErrUserIDRequired):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    "REQUIRED_FIELD",
			field:   "user_id",
			detail:  "user_id is required",
		}
	case errors.Is(err, application.ErrPostIDRequired):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    "REQUIRED_FIELD",
			field:   "id",
			detail:  "post id is required",
		}
	case errors.Is(err, application.ErrPageMustBePositive):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    share.INVALID_FIELD,
			field:   "page",
			detail:  "page must be greater than 0",
		}
	case errors.Is(err, application.ErrLimitMustBePositive):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    share.INVALID_FIELD,
			field:   "limit",
			detail:  "limit must be greater than 0",
		}
	case errors.Is(err, application.ErrLocationNameRequired):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    share.REQUIRED_FIELD,
			field:   "location_name",
			detail:  "location_name is required",
		}
	case errors.Is(err, application.ErrLatitudeOutOfRange):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    share.INVALID_FIELD,
			field:   "latitude",
			detail:  "latitude must be between -90 and 90",
		}
	case errors.Is(err, application.ErrLongitudeOutOfRange):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    share.INVALID_FIELD,
			field:   "longitude",
			detail:  "longitude must be between -180 and 180",
		}
	case errors.Is(err, value_object.ErrImageURLRequired):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    share.REQUIRED_FIELD,
			field:   "image_url",
			detail:  "image_url is required",
		}
	case errors.Is(err, value_object.ErrInvalidImageURL):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    share.INVALID_FIELD,
			field:   "image_url",
			detail:  "image_url must be a valid URL",
		}
	case errors.Is(err, value_object.ErrCaptionTooLong):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    share.INVALID_FIELD,
			field:   "caption",
			detail:  "caption exceeds max length",
		}
	case errors.Is(err, value_object.ErrLocationNameTooLong):
		return apiError{
			status:  http.StatusBadRequest,
			message: share.ValidationField,
			code:    share.INVALID_FIELD,
			field:   "location_name",
			detail:  "location_name exceeds max length",
		}
	case errors.Is(err, domain.ErrPostNotFound):
		return apiError{
			status:  http.StatusNotFound,
			message: "Post not found",
			code:    share.NOT_FOUND,
			detail:  "Requested post does not exist",
		}
	case errors.Is(err, domain.ErrPostDependencyUnavailable):
		return apiError{
			status:  http.StatusServiceUnavailable,
			message: "Post service unavailable",
			code:    share.SERVICE_UNAVAILABLE,
			detail:  "Post service is temporarily unavailable",
			logErr:  true,
		}
	default:
		return apiError{
			status:  http.StatusInternalServerError,
			message: "Internal server error",
			code:    share.INTERNAL_ERROR,
			detail:  "An unexpected error occurred",
			logErr:  true,
		}
	}
}
