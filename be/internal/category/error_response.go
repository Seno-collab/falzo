package category

import (
	"errors"
	"net/http"

	"falzo-be/internal/share"
)

var (
	errInvalidJSONPayload            = errors.New("invalid JSON payload")
	errorCategoryQueryParamsRequired = errors.New("category query params required")
	errorInvalidCategoryIDParam      = errors.New("invalid category id param")
	errorInvalidCategoryNameParam    = errors.New("invalid category name param")
	errorInvalidCategorySlugParam    = errors.New("invalid category slug param")
)

func mapCategoryError(err error) share.ApiError {
	switch {
	case errors.Is(err, errInvalidJSONPayload):
		return share.ApiError{
			Status:  http.StatusBadRequest,
			Message: share.ValidationField,
			Code:    share.INVALID_FORMAT,
			Detail:  "Invalid JSON payload",
		}
	case errors.Is(err, errorCategoryQueryParamsRequired):
		return share.Required("", "name, slug are required")
	case errors.Is(err, errorInvalidCategoryIDParam):
		return share.BadRequest("id", "id must be a valid integer")
	case errors.Is(err, errorInvalidCategoryNameParam):
		return share.BadRequest("name", "name must be a valid string")
	case errors.Is(err, errorInvalidCategorySlugParam):
		return share.BadRequest("slug", "slug must be a valid string")
	case errors.Is(err, ErrNameRequired):
		return share.Required("name", "name is required")
	case errors.Is(err, ErrSlugRequired):
		return share.Required("slug", "slug is required")
	case errors.Is(err, ErrNameTooLong):
		return share.BadRequest("name", "name cannot exceed 255 characters")
	case errors.Is(err, ErrSlugTooLong):
		return share.BadRequest("slug", "slug cannot exceed 255 characters")
	case errors.Is(err, ErrNotFound):
		return share.NotFound("Category not found", "Requested category does not exist")
	case errors.Is(err, ErrAlreadyExists):
		return share.ApiError{
			Status:  http.StatusConflict,
			Message: "Category already exists",
			Code:    "ALREADY_EXISTS",
			Detail:  "Category name or slug is already in use",
		}
	case errors.Is(err, ErrDependencyUnavailable):
		return share.ServiceUnavailable("Category service unavailable", "Category service is temporarily unavailable")
	default:
		return share.Internal()
	}
}
