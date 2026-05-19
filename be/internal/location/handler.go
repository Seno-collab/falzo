package location

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

type handlerService interface {
	Search(ctx context.Context, input SearchInput) ([]Location, error)
	Nearby(ctx context.Context, input NearbyInput) ([]NearbyLocation, error)
	GetPostsByLocation(ctx context.Context, input GetPostsByLocationInput) ([]LocationPost, error)
}

type Handler struct {
	service         handlerService
	readMiddlewares []func(http.Handler) http.Handler
}

type HandlerOption func(*Handler)

func WithReadMiddlewares(middlewares ...func(http.Handler) http.Handler) HandlerOption {
	return func(h *Handler) {
		h.readMiddlewares = append(h.readMiddlewares, middlewares...)
	}
}

func NewHandler(service handlerService, options ...HandlerOption) *Handler {
	h := &Handler{service: service}
	for _, option := range options {
		if option != nil {
			option(h)
		}
	}
	return h
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(h.readMiddlewares...).Get("/search", h.Search)
	r.With(h.readMiddlewares...).Get("/nearby", h.Nearby)
	r.With(h.readMiddlewares...).Get("/{id}/posts", h.GetPostsByLocation)
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
