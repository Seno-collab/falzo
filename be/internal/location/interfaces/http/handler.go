package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"falzo-be/internal/location/application"
	"falzo-be/internal/location/application/query"
	"falzo-be/internal/location/domain"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	service application.Service
}

func New(service application.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/search", h.Search)
	r.Get("/nearby", h.Nearby)
	r.Get("/{id}/posts", h.GetPostsByLocation)
	return r
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	locations, err := h.service.Search(r.Context(), query.SearchLocation{
		Query: q,
	})
	if err != nil {
		writeLocationError(w, r, err, "search_location")
		return
	}

	httpResponse.Success(w, http.StatusOK, "Locations fetched successfully", locations, r)
}

func (h *Handler) Nearby(w http.ResponseWriter, r *http.Request) {
	latRaw := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngRaw := strings.TrimSpace(r.URL.Query().Get("lng"))
	radiusRaw := strings.TrimSpace(r.URL.Query().Get("radius"))

	if latRaw == "" || lngRaw == "" || radiusRaw == "" {
		httpResponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpResponse.ErrorDetail{
			Code:    "REQUIRED_FIELD",
			Message: "lat, lng and radius are required",
		})
		return
	}

	lat, err := strconv.ParseFloat(latRaw, 64)
	if err != nil {
		httpResponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpResponse.ErrorDetail{
			Code:    "INVALID_FIELD",
			Field:   "lat",
			Message: "lat must be a valid float64",
		})
		return
	}

	lng, err := strconv.ParseFloat(lngRaw, 64)
	if err != nil {
		httpResponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpResponse.ErrorDetail{
			Code:    "INVALID_FIELD",
			Field:   "lng",
			Message: "lng must be a valid float64",
		})
		return
	}

	radius, err := strconv.ParseFloat(radiusRaw, 64)
	if err != nil {
		httpResponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpResponse.ErrorDetail{
			Code:    "INVALID_FIELD",
			Field:   "radius",
			Message: "radius must be a valid float64",
		})
		return
	}

	locations, err := h.service.Nearby(r.Context(), query.NearbyLocation{
		Latitude:     lat,
		Longitude:    lng,
		RadiusMeters: radius,
	})
	if err != nil {
		writeLocationError(w, r, err, "nearby_location")
		return
	}

	httpResponse.Success(w, http.StatusOK, "Nearby locations fetched successfully", locations, r)
}

func (h *Handler) GetPostsByLocation(w http.ResponseWriter, r *http.Request) {
	locationID := strings.TrimSpace(chi.URLParam(r, "id"))

	posts, err := h.service.GetPostsByLocation(r.Context(), query.GetPostsByLocation{
		LocationID: locationID,
	})
	if err != nil {
		writeLocationError(w, r, err, "get_posts_by_location")
		return
	}

	httpResponse.Success(w, http.StatusOK, "Posts by location fetched successfully", posts, r)
}

type apiError struct {
	status  int
	message string
	code    string
	field   string
	detail  string
	logErr  bool
}

func writeLocationError(w http.ResponseWriter, r *http.Request, err error, operation string) {
	mapped := mapLocationError(err)
	if mapped.logErr {
		log.Error().
			Err(err).
			Str("operation", operation).
			Str("request_id", chimiddleware.GetReqID(r.Context())).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("location request failed")
	}

	httpResponse.Error(w, mapped.status, mapped.message, r, httpResponse.ErrorDetail{
		Code:    mapped.code,
		Field:   mapped.field,
		Message: mapped.detail,
	})
}

func mapLocationError(err error) apiError {
	switch {
	case errors.Is(err, application.ErrSearchQueryRequired):
		return apiError{
			status:  http.StatusBadRequest,
			message: "Validation failed",
			code:    "REQUIRED_FIELD",
			field:   "q",
			detail:  "q is required",
		}
	case errors.Is(err, application.ErrLatitudeOutOfRange):
		return apiError{
			status:  http.StatusBadRequest,
			message: "Validation failed",
			code:    "INVALID_FIELD",
			field:   "lat",
			detail:  "lat must be between -90 and 90",
		}
	case errors.Is(err, application.ErrLongitudeOutOfRange):
		return apiError{
			status:  http.StatusBadRequest,
			message: "Validation failed",
			code:    "INVALID_FIELD",
			field:   "lng",
			detail:  "lng must be between -180 and 180",
		}
	case errors.Is(err, application.ErrRadiusMustBePositive):
		return apiError{
			status:  http.StatusBadRequest,
			message: "Validation failed",
			code:    "INVALID_FIELD",
			field:   "radius",
			detail:  "radius must be greater than 0",
		}
	case errors.Is(err, application.ErrLocationIDRequired):
		return apiError{
			status:  http.StatusBadRequest,
			message: "Validation failed",
			code:    "REQUIRED_FIELD",
			field:   "id",
			detail:  "location id is required",
		}
	case errors.Is(err, domain.ErrLocationDependencyUnavailable):
		return apiError{
			status:  http.StatusServiceUnavailable,
			message: "Location service unavailable",
			code:    "SERVICE_UNAVAILABLE",
			detail:  "Location service is temporarily unavailable",
			logErr:  true,
		}
	case errors.Is(err, domain.ErrLocationInternal):
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
