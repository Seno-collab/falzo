package location

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

var (
	errLocationQueryParamsRequired = errors.New("location query params required")
	errInvalidLatitudeParam        = errors.New("invalid latitude param")
	errInvalidLongitudeParam       = errors.New("invalid longitude param")
	errInvalidRadiusParam          = errors.New("invalid radius param")
)

type handlerService interface {
	Search(ctx context.Context, input SearchInput) ([]Location, error)
	Nearby(ctx context.Context, input NearbyInput) ([]NearbyLocation, error)
	GetPostsByLocation(ctx context.Context, input GetPostsByLocationInput) ([]LocationPost, error)
}

type Handler struct {
	service handlerService
}

func NewHandler(service handlerService) *Handler {
	return &Handler{service: service}
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

	locations, err := h.service.Search(r.Context(), SearchInput{Query: q})
	if err != nil {
		share.WriteError(w, r, err, "search_location", mapLocationError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Locations fetched successfully", locations, r)
}

func (h *Handler) Nearby(w http.ResponseWriter, r *http.Request) {
	latRaw := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngRaw := strings.TrimSpace(r.URL.Query().Get("lng"))
	radiusRaw := strings.TrimSpace(r.URL.Query().Get("radius"))

	if latRaw == "" || lngRaw == "" || radiusRaw == "" {
		share.WriteError(w, r, errLocationQueryParamsRequired, "nearby_location", mapLocationError)
		return
	}

	lat, err := strconv.ParseFloat(latRaw, 64)
	if err != nil {
		share.WriteError(w, r, errInvalidLatitudeParam, "nearby_location", mapLocationError)
		return
	}

	lng, err := strconv.ParseFloat(lngRaw, 64)
	if err != nil {
		share.WriteError(w, r, errInvalidLongitudeParam, "nearby_location", mapLocationError)
		return
	}

	radius, err := strconv.ParseFloat(radiusRaw, 64)
	if err != nil {
		share.WriteError(w, r, errInvalidRadiusParam, "nearby_location", mapLocationError)
		return
	}

	locations, err := h.service.Nearby(r.Context(), NearbyInput{
		Latitude:     lat,
		Longitude:    lng,
		RadiusMeters: radius,
	})
	if err != nil {
		share.WriteError(w, r, err, "nearby_location", mapLocationError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Nearby locations fetched successfully", locations, r)
}

func (h *Handler) GetPostsByLocation(w http.ResponseWriter, r *http.Request) {
	locationID := strings.TrimSpace(chi.URLParam(r, "id"))

	posts, err := h.service.GetPostsByLocation(r.Context(), GetPostsByLocationInput{LocationID: locationID})
	if err != nil {
		share.WriteError(w, r, err, "get_posts_by_location", mapLocationError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Posts by location fetched successfully", posts, r)
}

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
	case errors.Is(err, ErrDependencyUnavailable):
		return share.ServiceUnavailable("Location service unavailable", "Location service is temporarily unavailable")
	default:
		return share.Internal()
	}
}
