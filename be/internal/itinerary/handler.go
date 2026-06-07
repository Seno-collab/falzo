package itinerary

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
	ListPublic(ctx context.Context, input ListInput) (ListPage, error)
	GetPublicBySlug(ctx context.Context, input GetBySlugInput) (Detail, error)
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
	r.With(h.readMiddlewares...).Get("/", h.List)
	r.With(h.readMiddlewares...).Get("/{slug}", h.GetBySlug)
	return r
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	input, err := parseListInput(r)
	if err != nil {
		share.WriteError(w, r, err, "list_itineraries", mapItineraryError)
		return
	}

	page, err := h.service.ListPublic(r.Context(), input)
	if err != nil {
		share.WriteError(w, r, err, "list_itineraries", mapItineraryError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Itineraries fetched successfully", page, r)
}

func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSpace(chi.URLParam(r, "slug"))

	detail, err := h.service.GetPublicBySlug(r.Context(), GetBySlugInput{Slug: slug})
	if err != nil {
		share.WriteError(w, r, err, "get_itinerary_by_slug", mapItineraryError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Itinerary fetched successfully", detail, r)
}

func parseListInput(r *http.Request) (ListInput, error) {
	query := r.URL.Query()

	page, err := parseOptionalInt(query.Get("page"), errInvalidPageParam)
	if err != nil {
		return ListInput{}, err
	}
	if strings.TrimSpace(query.Get("page")) != "" && page <= 0 {
		return ListInput{}, ErrPageMustBePositive
	}
	limit, err := parseOptionalInt(query.Get("limit"), errInvalidLimitParam)
	if err != nil {
		return ListInput{}, err
	}
	if strings.TrimSpace(query.Get("limit")) != "" && limit <= 0 {
		return ListInput{}, ErrLimitMustBePositive
	}
	durationDays, err := parseOptionalInt(query.Get("durationDays"), errInvalidDurationDaysParam)
	if err != nil {
		return ListInput{}, err
	}
	if strings.TrimSpace(query.Get("durationDays")) != "" && durationDays <= 0 {
		return ListInput{}, ErrInvalidDurationDays
	}
	budgetMax, err := parseOptionalInt(query.Get("budgetMax"), errInvalidBudgetMaxParam)
	if err != nil {
		return ListInput{}, err
	}

	return ListInput{
		Province:     query.Get("province"),
		DurationDays: durationDays,
		BudgetMax:    budgetMax,
		TravelStyle:  query.Get("travelStyle"),
		Page:         page,
		Limit:        limit,
	}, nil
}

func parseOptionalInt(raw string, parseErr error) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, parseErr
	}
	return parsed, nil
}
