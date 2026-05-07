package category

import (
	"context"
	"encoding/json"
	"errors"
	"falzo-be/internal/auth"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var (
	errorCategoryQueryParamsRequired = errors.New("category query params required")
	errorInvalidCategoryIDParam      = errors.New("invalid category id param")
	errorInvalidCategoryNameParam    = errors.New("invalid category name param")
	errorInvalidCategorySlugParam    = errors.New("invalid category slug param")
)

type handlerService interface {
	Create(ctx context.Context, name, slug string) error
	GetByID(ctx context.Context, id uint64) (Category, error)
	GetBySlug(ctx context.Context, slug string) (Category, error)
	List(ctx context.Context) ([]Category, error)
	Update(ctx context.Context, id uint64, name, slug string) (Category, error)
	Delete(ctx context.Context, id uint64) error
}

type Handler struct {
	service     handlerService
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	}
}

func NewHandler(service handlerService, authService interface {
	Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
}) *Handler {
	return &Handler{service: service, authService: authService}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/slug/{slug}", h.GetBySlug)
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth(h.authService))
		protected.Post("/", h.Create)
		protected.Get("/{id}", h.GetByID)
		protected.Put("/{id}", h.Update)
		protected.Delete("/{id}", h.Delete)
	})
	return r
}

type createCategoryRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// Implementation for creating a category
	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mapCategoryError(err)
		return
	}
	err := h.service.Create(r.Context(), req.Name, req.Slug)
	if err != nil {
		mapCategoryError(err)
		return
	}
	httpResponse.Success(w, http.StatusCreated, "Category created successfully", nil, r)
}

func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	category, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		mapCategoryError(err)
		return
	}
	httpResponse.Success(w, http.StatusOK, "Category fetched successfully", category, r)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.List(r.Context())
	if err != nil {
		mapCategoryError(err)
		return
	}
	httpResponse.Success(w, http.StatusOK, "Categories fetched successfully", categories, r)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idUint, err := parseID(id)
	if err != nil {
		mapCategoryError(errorInvalidCategoryIDParam)
		return
	}
	category, err := h.service.GetByID(r.Context(), idUint)
	if err != nil {
		mapCategoryError(err)
		return
	}
	httpResponse.Success(w, http.StatusOK, "Category fetched successfully", category, r)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	// Implementation for updating a category
	id := chi.URLParam(r, "id")
	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mapCategoryError(err)
		return
	}
	idUint, err := parseID(id)
	if err != nil {
		mapCategoryError(errorInvalidCategoryIDParam)
		return
	}
	updatedCategory, err := h.service.Update(r.Context(), idUint, req.Name, req.Slug)
	if err != nil {
		mapCategoryError(err)
		return
	}
	httpResponse.Success(w, http.StatusOK, "Category updated successfully", updatedCategory, r)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// Implementation for deleting a category
	id := chi.URLParam(r, "id")
	idUint, err := parseID(id)
	if err != nil {
		mapCategoryError(errorInvalidCategoryIDParam)
		return
	}
	err = h.service.Delete(r.Context(), idUint)
	if err != nil {
		mapCategoryError(err)
		return
	}
	httpResponse.Success(w, http.StatusOK, "Category deleted successfully", nil, r)
}

func parseID(idStr string) (uint64, error) {
	var id uint64
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		return 0, errorInvalidCategoryIDParam
	}
	return id, nil
}

func mapCategoryError(err error) share.ApiError {
	switch {
	case errors.Is(err, errorCategoryQueryParamsRequired):
		return share.Required("", "name, slug are required")
	case errors.Is(err, errorInvalidCategoryIDParam):
		return share.BadRequest("id", "id must be a valid integer")
	case errors.Is(err, errorInvalidCategoryNameParam):
		return share.BadRequest("name", "name must be a valid string")
	case errors.Is(err, errorInvalidCategorySlugParam):
		return share.BadRequest("slug", "slug must be a valid string")
	default:
		return share.Internal()
	}
}
