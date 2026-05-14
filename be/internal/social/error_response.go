package social

import (
	"errors"
	"net/http"

	"falzo-be/internal/share"
)

var errInvalidUserIDParam = errors.New("invalid user id param")

func mapSocialError(err error) share.ApiError {
	switch {
	case errors.Is(err, errInvalidUserIDParam):
		return share.BadRequest("id", "id must be a valid positive integer")
	case errors.Is(err, ErrUserIDRequired):
		return share.Required("user_id", "user_id is required")
	case errors.Is(err, ErrTargetUserIDRequired):
		return share.Required("target_user_id", "target user id is required")
	case errors.Is(err, ErrCannotFollowSelf):
		return share.ApiError{
			Status:  http.StatusBadRequest,
			Message: share.ValidationField,
			Code:    share.INVALID_FORMAT,
			Detail:  "You cannot follow yourself",
			Field:   "target_user_id",
		}
	case errors.Is(err, ErrUserNotFound):
		return share.NotFound("User not found", "Requested user does not exist")
	case errors.Is(err, ErrDependencyUnavailable):
		return share.ServiceUnavailable("Social service unavailable", "Social service is temporarily unavailable")
	default:
		return share.Internal()
	}
}
