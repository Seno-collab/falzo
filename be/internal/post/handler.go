package post

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"falzo-be/internal/auth"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

type handlerService interface {
	CreatePost(ctx context.Context, input CreatePostInput) (PostView, error)
	LikePost(ctx context.Context, input PostActionInput) error
	SavePost(ctx context.Context, input PostActionInput) error
	CommentPost(ctx context.Context, input CommentPostInput) (CommentView, error)
	GetPosts(ctx context.Context, input ListPostsInput) ([]PostView, error)
	GetPostDetail(ctx context.Context, input GetPostDetailInput) (*PostView, error)
	GetPostsByLocation(ctx context.Context, input GetPostsByLocationInput) ([]PostView, error)
	GetComments(ctx context.Context, input ListCommentsInput) ([]CommentView, error)
}

type Handler struct {
	service     handlerService
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	}
}

func NewHandler(
	service handlerService,
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	},
) *Handler {
	return &Handler{service: service, authService: authService}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetPosts)
	r.Get("/location", h.GetPostsByLocation)
	r.Get("/{id}/comments", h.GetComments)
	r.Get("/{id}", h.GetPostDetail)
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth(h.authService))
		protected.Post("/", h.CreatePost)
		protected.Post("/{id}/like", h.LikePost)
		protected.Post("/{id}/save", h.SavePost)
		protected.Post("/{id}/comments", h.CommentPost)
	})
	return r
}

type CreatePostRequest struct {
	ImageURL     string  `json:"image_url"`
	Caption      string  `json:"caption"`
	LocationName string  `json:"location_name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

type CommentPostRequest struct {
	Content string `json:"content"`
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "create_post", mapPostError)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "create_post", mapPostError)
		return
	}

	post, err := h.service.CreatePost(r.Context(), CreatePostInput{
		UserID:       principal.UserID,
		ImageURL:     req.ImageURL,
		Caption:      req.Caption,
		LocationName: req.LocationName,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
	})
	if err != nil {
		share.WriteError(w, r, err, "create_post", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusCreated, "Post created successfully", post, r)
}

func (h *Handler) GetPostDetail(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || postID == 0 {
		share.WriteError(w, r, errInvalidPostIDParam, "get_post_detail", mapPostError)
		return
	}

	post, err := h.service.GetPostDetail(r.Context(), GetPostDetailInput{PostID: postID})
	if err != nil {
		share.WriteError(w, r, err, "get_post_detail", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Post detail fetched successfully", post, r)
}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || postID == 0 {
		share.WriteError(w, r, errInvalidPostIDParam, "get_comments", mapPostError)
		return
	}

	page, limit, ok := h.parsePageLimit(w, r, "get_comments")
	if !ok {
		return
	}

	comments, err := h.service.GetComments(r.Context(), ListCommentsInput{PostID: postID, Page: page, Limit: limit})
	if err != nil {
		share.WriteError(w, r, err, "get_comments", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Comments fetched successfully", comments, r)
}

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	page, limit, ok := h.parsePageLimit(w, r, "get_posts")
	if !ok {
		return
	}

	posts, err := h.service.GetPosts(r.Context(), ListPostsInput{Page: page, Limit: limit})
	if err != nil {
		share.WriteError(w, r, err, "get_posts", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Posts fetched successfully", posts, r)
}

func (h *Handler) parsePageLimit(w http.ResponseWriter, r *http.Request, operation string) (int, int, bool) {
	page := 1
	limit := 12

	pageRaw := strings.TrimSpace(r.URL.Query().Get("page"))
	if pageRaw != "" {
		parsedPage, err := strconv.Atoi(pageRaw)
		if err != nil {
			share.WriteError(w, r, errInvalidPageParam, operation, mapPostError)
			return 0, 0, false
		}
		page = parsedPage
	}

	limitRaw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if limitRaw != "" {
		parsedLimit, err := strconv.Atoi(limitRaw)
		if err != nil {
			share.WriteError(w, r, errInvalidLimitParam, operation, mapPostError)
			return 0, 0, false
		}
		limit = parsedLimit
	}

	return page, limit, true
}

func (h *Handler) GetPostsByLocation(w http.ResponseWriter, r *http.Request) {
	locationName := strings.TrimSpace(r.URL.Query().Get("location_name"))

	posts, err := h.service.GetPostsByLocation(r.Context(), GetPostsByLocationInput{LocationName: locationName})
	if err != nil {
		share.WriteError(w, r, err, "get_posts_by_location", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Posts by location fetched successfully", posts, r)
}

func (h *Handler) LikePost(w http.ResponseWriter, r *http.Request) {
	h.handlePostAction(w, r, "like_post", "Post liked successfully", "post liked", h.service.LikePost)
}

func (h *Handler) SavePost(w http.ResponseWriter, r *http.Request) {
	h.handlePostAction(w, r, "save_post", "Post saved successfully", "post saved", h.service.SavePost)
}

func (h *Handler) CommentPost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || postID == 0 {
		share.WriteError(w, r, errInvalidPostIDParam, "comment_post", mapPostError)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "comment_post", mapPostError)
		return
	}

	var req CommentPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "comment_post", mapPostError)
		return
	}

	comment, err := h.service.CommentPost(r.Context(), CommentPostInput{
		PostID:  postID,
		UserID:  principal.UserID,
		Content: req.Content,
	})
	if err != nil {
		share.WriteError(w, r, err, "comment_post", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusCreated, "Comment created successfully", comment, r)
}

func (h *Handler) handlePostAction(
	w http.ResponseWriter,
	r *http.Request,
	operation string,
	successMessage string,
	payloadMessage string,
	action func(context.Context, PostActionInput) error,
) {
	postID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || postID == 0 {
		share.WriteError(w, r, errInvalidPostIDParam, operation, mapPostError)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, operation, mapPostError)
		return
	}

	if err := action(r.Context(), PostActionInput{PostID: postID, UserID: principal.UserID}); err != nil {
		share.WriteError(w, r, err, operation, mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, successMessage, map[string]string{"message": payloadMessage}, r)
}
