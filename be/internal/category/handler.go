package category

import (
	"context"
	"encoding/json"
	"falzo-be/internal/auth"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
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
	readMiddlewares []func(http.Handler) http.Handler
}

type HandlerOption func(*Handler)

func WithReadMiddlewares(middlewares ...func(http.Handler) http.Handler) HandlerOption {
	return func(h *Handler) {
		h.readMiddlewares = append(h.readMiddlewares, middlewares...)
	}
}

func NewHandler(service handlerService, authService interface {
	Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
}, options ...HandlerOption) *Handler {
	h := &Handler{service: service, authService: authService}
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
	r.With(h.readMiddlewares...).Get("/slug/{slug}", h.GetBySlug)
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth(h.authService))
		protected.Post("/", h.Create)
		protected.With(h.readMiddlewares...).Get("/{id}", h.GetByID)
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
	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "create_category", mapCategoryError)
		return
	}
	err := h.service.Create(r.Context(), req.Name, req.Slug)
	if err != nil {
		share.WriteError(w, r, err, "create_category", mapCategoryError)
		return
	}
	httpResponse.Success(w, http.StatusCreated, "Category created successfully", nil, r)
}

func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	category, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		share.WriteError(w, r, err, "get_category_by_slug", mapCategoryError)
		return
	}
	httpResponse.Success(w, http.StatusOK, "Category fetched successfully", category, r)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.List(r.Context())
	if err != nil {
		share.WriteError(w, r, err, "list_categories", mapCategoryError)
		return
	}
	httpResponse.Success(w, http.StatusOK, "Categories fetched successfully", categories, r)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idUint, err := parseID(id)
	if err != nil {
		share.WriteError(w, r, errorInvalidCategoryIDParam, "get_category_by_id", mapCategoryError)
		return
	}
	category, err := h.service.GetByID(r.Context(), idUint)
	if err != nil {
		share.WriteError(w, r, err, "get_category_by_id", mapCategoryError)
		return
	}
	httpResponse.Success(w, http.StatusOK, "Category fetched successfully", category, r)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "update_category", mapCategoryError)
		return
	}
	idUint, err := parseID(id)
	if err != nil {
		share.WriteError(w, r, errorInvalidCategoryIDParam, "update_category", mapCategoryError)
		return
	}
	updatedCategory, err := h.service.Update(r.Context(), idUint, req.Name, req.Slug)
	if err != nil {
		share.WriteError(w, r, err, "update_category", mapCategoryError)
		return
	}
	httpResponse.Success(w, http.StatusOK, "Category updated successfully", updatedCategory, r)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idUint, err := parseID(id)
	if err != nil {
		share.WriteError(w, r, errorInvalidCategoryIDParam, "delete_category", mapCategoryError)
		return
	}
	err = h.service.Delete(r.Context(), idUint)
	if err != nil {
		share.WriteError(w, r, err, "delete_category", mapCategoryError)
		return
	}
	httpResponse.Success(w, http.StatusOK, "Category deleted successfully", nil, r)
}

func parseID(idStr string) (uint64, error) {
	id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
	if err != nil || id == 0 {
		return 0, errorInvalidCategoryIDParam
	}
	return id, nil
}
