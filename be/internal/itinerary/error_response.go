package itinerary

import (
	"errors"

	"falzo-be/internal/share"
)

var (
	errInvalidPageParam         = errors.New("invalid page param")
	errInvalidLimitParam        = errors.New("invalid limit param")
	errInvalidDurationDaysParam = errors.New("invalid durationDays param")
	errInvalidBudgetMaxParam    = errors.New("invalid budgetMax param")
)

func mapItineraryError(err error) share.ApiError {
	switch {
	case errors.Is(err, errInvalidPageParam):
		return share.BadRequest("page", "page must be an integer")
	case errors.Is(err, errInvalidLimitParam):
		return share.BadRequest("limit", "limit must be an integer")
	case errors.Is(err, errInvalidDurationDaysParam):
		return share.BadRequest("durationDays", "durationDays must be an integer")
	case errors.Is(err, errInvalidBudgetMaxParam):
		return share.BadRequest("budgetMax", "budgetMax must be an integer")
	case errors.Is(err, ErrSlugRequired):
		return share.Required("slug", "itinerary slug is required")
	case errors.Is(err, ErrPageMustBePositive):
		return share.BadRequest("page", "page must be greater than 0")
	case errors.Is(err, ErrLimitMustBePositive):
		return share.BadRequest("limit", "limit must be greater than 0")
	case errors.Is(err, ErrLimitTooLarge):
		return share.BadRequest("limit", "limit must not exceed 50")
	case errors.Is(err, ErrInvalidDurationDays):
		return share.BadRequest("durationDays", "durationDays must be between 1 and 14")
	case errors.Is(err, ErrInvalidBudgetMax):
		return share.BadRequest("budgetMax", "budgetMax must not be negative")
	case errors.Is(err, ErrNotFound):
		return share.NotFound("Itinerary not found", "Requested itinerary does not exist")
	case errors.Is(err, ErrDependencyUnavailable):
		return share.ServiceUnavailable("Itinerary service unavailable", "Itinerary service is temporarily unavailable")
	default:
		return share.Internal()
	}
}
