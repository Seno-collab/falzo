package location

import (
	"errors"

	"falzo-be/internal/share"
)

var (
	errLocationQueryParamsRequired = errors.New("location query params required")
	errInvalidLatitudeParam        = errors.New("invalid latitude param")
	errInvalidLongitudeParam       = errors.New("invalid longitude param")
	errInvalidRadiusParam          = errors.New("invalid radius param")
)

func mapLocationError(err error) share.ApiError {
	switch {
	case errors.Is(err, errLocationQueryParamsRequired):
		return share.Required("", "lat, lng and radius are required")
	case errors.Is(err, errInvalidLatitudeParam):
		return share.BadRequest("lat", "lat must be a valid float64")
	case errors.Is(err, errInvalidLongitudeParam):
		return share.BadRequest("lng", "lng must be a valid float64")
	case errors.Is(err, errInvalidRadiusParam):
		return share.BadRequest("radius", "radius must be a valid float64")
	case errors.Is(err, ErrSearchQueryRequired):
		return share.Required("q", "q is required")
	case errors.Is(err, ErrLatitudeOutOfRange):
		return share.BadRequest("lat", "lat must be between -90 and 90")
	case errors.Is(err, ErrLongitudeOutOfRange):
		return share.BadRequest("lng", "lng must be between -180 and 180")
	case errors.Is(err, ErrRadiusMustBePositive):
		return share.BadRequest("radius", "radius must be greater than 0")
	case errors.Is(err, ErrLocationIDRequired):
		return share.Required("id", "location id is required")
	case errors.Is(err, ErrPlaceSlugRequired):
		return share.Required("slug", "place slug is required")
	case errors.Is(err, ErrPlaceNotFound):
		return share.NotFound("Place not found", "Requested place does not exist")
	case errors.Is(err, ErrDependencyUnavailable):
		return share.ServiceUnavailable("Location service unavailable", "Location service is temporarily unavailable")
	default:
		return share.Internal()
	}
}
