package post

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"falzo-be/internal/auth"
	"falzo-be/internal/notification"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type handlerService interface {
	CreatePost(ctx context.Context, input CreatePostInput) (PostView, error)
	UpdatePost(ctx context.Context, input UpdatePostInput) (PostView, error)
	DeletePost(ctx context.Context, input ModerationInput) error
	HidePost(ctx context.Context, input ModerationInput) error
	ReportPost(ctx context.Context, input ReportInput) error
	LikePost(ctx context.Context, input PostActionInput) error
	UnlikePost(ctx context.Context, input PostActionInput) error
	SavePost(ctx context.Context, input PostActionInput) error
	UnsavePost(ctx context.Context, input PostActionInput) error
	CreateSavedCollection(ctx context.Context, input CreateSavedCollectionInput) (SavedCollectionView, error)
	ListSavedCollections(ctx context.Context, input SavedCollectionInput) ([]SavedCollectionView, error)
	ListSavedPosts(ctx context.Context, input SavedCollectionInput) ([]PostView, error)
	AddPostToSavedCollection(ctx context.Context, input SavedCollectionPostInput) error
	RemovePostFromSavedCollection(ctx context.Context, input SavedCollectionPostInput) error
	DeleteSavedCollection(ctx context.Context, input SavedCollectionInput) error
	UpdateSavedCollectionVisibility(ctx context.Context, input UpdateSavedCollectionVisibilityInput) (SavedCollectionView, error)
	GetPublicSavedCollection(ctx context.Context, input PublicSavedCollectionInput) (*SavedCollectionView, error)
	CommentPost(ctx context.Context, input CommentPostInput) (CommentView, error)
	UpdateComment(ctx context.Context, input UpdateCommentInput) (CommentView, error)
	DeleteComment(ctx context.Context, input ModerationInput) error
	HideComment(ctx context.Context, input ModerationInput) error
	ReportComment(ctx context.Context, input ReportInput) error
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
	commentEvents      CommentEventSubscriber
	commentPublisher   CommentEventPublisher
	postEvents         PostEventSubscriber
	postPublisher      PostEventPublisher
	notifications      notification.Publisher
	followers          FollowerLister
	commentMiddlewares []func(http.Handler) http.Handler
	reportMiddlewares  []func(http.Handler) http.Handler
}

type HandlerOption func(*Handler)

type FollowerLister interface {
	ListFollowerIDs(ctx context.Context, userID uint64) ([]uint64, error)
}

func WithCommentEvents(subscriber CommentEventSubscriber, publisher CommentEventPublisher) HandlerOption {
	return func(h *Handler) {
		h.commentEvents = subscriber
		h.commentPublisher = publisher
	}
}

func WithPostEvents(subscriber PostEventSubscriber, publisher PostEventPublisher) HandlerOption {
	return func(h *Handler) {
		h.postEvents = subscriber
		h.postPublisher = publisher
	}
}

func WithNotifications(publisher notification.Publisher) HandlerOption {
	return func(h *Handler) {
		h.notifications = publisher
	}
}

func WithFollowers(followers FollowerLister) HandlerOption {
	return func(h *Handler) {
		h.followers = followers
	}
}

func WithCommentMiddlewares(middlewares ...func(http.Handler) http.Handler) HandlerOption {
	return func(h *Handler) {
		h.commentMiddlewares = append(h.commentMiddlewares, middlewares...)
	}
}

func WithReportMiddlewares(middlewares ...func(http.Handler) http.Handler) HandlerOption {
	return func(h *Handler) {
		h.reportMiddlewares = append(h.reportMiddlewares, middlewares...)
	}
}

func NewHandler(
	service handlerService,
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	},
	options ...HandlerOption,
) *Handler {
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
	r.Get("/", h.GetPosts)
	r.Get("/events", h.StreamPosts)
	r.Get("/location", h.GetPostsByLocation)
	r.Get("/saved-collections/public/{shareSlug}", h.GetPublicSavedCollection)
	r.Get("/{id}/comments/events", h.StreamComments)
	r.Get("/{id}/comments", h.GetComments)
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth(h.authService))
		protected.Get("/saved", h.ListSavedPosts)
		protected.Get("/saved-collections", h.ListSavedCollections)
		protected.Post("/saved-collections", h.CreateSavedCollection)
		protected.Patch("/saved-collections/{collectionID}", h.UpdateSavedCollectionVisibility)
		protected.Delete("/saved-collections/{collectionID}", h.DeleteSavedCollection)
		protected.Post("/saved-collections/{collectionID}/posts/{postID}", h.AddPostToSavedCollection)
		protected.Delete("/saved-collections/{collectionID}/posts/{postID}", h.RemovePostFromSavedCollection)
		protected.Post("/", h.CreatePost)
		protected.Put("/{id}", h.UpdatePost)
		protected.Delete("/{id}", h.DeletePost)
		protected.Post("/{id}/hide", h.HidePost)
		protected.With(h.reportMiddlewares...).Post("/{id}/report", h.ReportPost)
		protected.Post("/{id}/like", h.LikePost)
		protected.Delete("/{id}/like", h.UnlikePost)
		protected.Post("/{id}/save", h.SavePost)
		protected.Delete("/{id}/save", h.UnsavePost)
		protected.With(h.commentMiddlewares...).Post("/{id}/comments", h.CommentPost)
		protected.With(h.commentMiddlewares...).Put("/{id}/comments/{commentID}", h.UpdateComment)
		protected.Delete("/{id}/comments/{commentID}", h.DeleteComment)
		protected.Post("/{id}/comments/{commentID}/hide", h.HideComment)
		protected.With(h.reportMiddlewares...).Post("/{id}/comments/{commentID}/report", h.ReportComment)
	})
	r.Get("/{id}", h.GetPostDetail)
	return r
}

type CreatePostRequest struct {
	ImageURL     string  `json:"image_url"`
	Caption      string  `json:"caption"`
	LocationName string  `json:"location_name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	CategoryID   uint64  `json:"category_id"`
}

type CommentPostRequest struct {
	Content          string `json:"content"`
	ReplyToCommentID uint64 `json:"reply_to_comment_id"`
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}

type ModerateContentRequest struct {
	Reason string `json:"reason"`
}

type ReportContentRequest struct {
	Reason string `json:"reason"`
}

type CreateSavedCollectionRequest struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

type UpdateSavedCollectionRequest struct {
	IsPublic *bool `json:"is_public"`
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
		CategoryID:   req.CategoryID,
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

	if h.postPublisher != nil {
		if err := h.postPublisher.PublishPostCreated(r.Context(), post); err != nil {
			log.Warn().Err(err).Uint64("post_id", post.ID).Msg("post event publish failed")
		}
	}
	h.publishFollowerPostNotifications(r.Context(), principal, post)

	httpResponse.Success(w, http.StatusCreated, "Post created successfully", post, r)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	postID, err := parseRouteUint(r, "id")
	if err != nil {
		share.WriteError(w, r, errInvalidPostIDParam, "update_post", mapPostError)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "update_post", mapPostError)
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "update_post", mapPostError)
		return
	}

	item, err := h.service.UpdatePost(r.Context(), UpdatePostInput{
		PostID:       postID,
		UserID:       principal.UserID,
		CategoryID:   req.CategoryID,
		Caption:      req.Caption,
		LocationName: req.LocationName,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
	})
	if err != nil {
		share.WriteError(w, r, err, "update_post", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Post updated successfully", item, r)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	h.handlePostModerationAction(w, r, "delete_post", "Post deleted successfully", false, h.service.DeletePost, h.publishPostDeleted)
}

func (h *Handler) HidePost(w http.ResponseWriter, r *http.Request) {
	h.handlePostModerationAction(w, r, "hide_post", "Post hidden successfully", true, h.service.HidePost, h.publishPostDeleted)
}

func (h *Handler) ReportPost(w http.ResponseWriter, r *http.Request) {
	postID, err := parseRouteUint(r, "id")
	if err != nil {
		share.WriteError(w, r, errInvalidPostIDParam, "report_post", mapPostError)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "report_post", mapPostError)
		return
	}

	var req ReportContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "report_post", mapPostError)
		return
	}

	if err := h.service.ReportPost(r.Context(), ReportInput{
		PostID:         postID,
		ReporterUserID: principal.UserID,
		Reason:         req.Reason,
	}); err != nil {
		share.WriteError(w, r, err, "report_post", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusCreated, "Post reported successfully", nil, r)
}

func (h *Handler) GetPostDetail(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || postID == 0 {
		share.WriteError(w, r, errInvalidPostIDParam, "get_post_detail", mapPostError)
		return
	}

	post, err := h.service.GetPostDetail(r.Context(), GetPostDetailInput{
		PostID:       postID,
		ViewerUserID: h.viewerUserID(r),
	})
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

func (h *Handler) StreamComments(w http.ResponseWriter, r *http.Request) {
	if h.commentEvents == nil {
		share.WriteError(w, r, ErrDependencyUnavailable, "stream_comments", mapPostError)
		return
	}

	postID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || postID == 0 {
		share.WriteError(w, r, errInvalidPostIDParam, "stream_comments", mapPostError)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	comments, unsubscribe := h.commentEvents.SubscribeComments(r.Context(), postID)
	defer unsubscribe()

	if _, err := w.Write([]byte(": connected\n\n")); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case comment, ok := <-comments:
			if !ok {
				return
			}
			if comment.Comment.PostID != postID {
				continue
			}
			if err := writeSSE(w, flusher, comment.Type, comment.Comment); err != nil {
				return
			}
		case <-heartbeat.C:
			if _, err := w.Write([]byte(": keep-alive\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	page, limit, ok := h.parsePageLimit(w, r, "get_posts")
	if !ok {
		return
	}

	sort := r.URL.Query().Get("sort")
	latRaw := r.URL.Query().Get("lat")
	lngRaw := r.URL.Query().Get("lng")
	latitude := parseOptionalFloat(latRaw)
	longitude := parseOptionalFloat(lngRaw)
	if strings.TrimSpace(sort) == postSortNearby && (strings.TrimSpace(latRaw) == "" || strings.TrimSpace(lngRaw) == "") {
		latitude = 999
	}

	posts, err := h.service.GetPosts(r.Context(), ListPostsInput{
		Page:         page,
		Limit:        limit,
		ViewerUserID: h.viewerUserID(r),
		Search:       r.URL.Query().Get("search"),
		CategorySlug: r.URL.Query().Get("category_slug"),
		Feed:         r.URL.Query().Get("feed"),
		Sort:         sort,
		Latitude:     latitude,
		Longitude:    longitude,
		RadiusMeters: parseOptionalInt(r.URL.Query().Get("radius_meters")),
	})
	if err != nil {
		share.WriteError(w, r, err, "get_posts", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Posts fetched successfully", posts, r)
}

func (h *Handler) StreamPosts(w http.ResponseWriter, r *http.Request) {
	if h.postEvents == nil {
		share.WriteError(w, r, ErrDependencyUnavailable, "stream_posts", mapPostError)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	posts, unsubscribe := h.postEvents.SubscribePosts(r.Context())
	defer unsubscribe()

	if _, err := w.Write([]byte(": connected\n\n")); err != nil {
		return
	}
	flusher.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-posts:
			if !ok {
				return
			}
			if err := writeSSE(w, flusher, event.Type, event.Post); err != nil {
				return
			}
		case <-heartbeat.C:
			if _, err := w.Write([]byte(": keep-alive\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
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

func (h *Handler) UnlikePost(w http.ResponseWriter, r *http.Request) {
	h.handlePostAction(w, r, "unlike_post", "Post unliked successfully", "post unliked", h.service.UnlikePost)
}

func (h *Handler) SavePost(w http.ResponseWriter, r *http.Request) {
	h.handlePostAction(w, r, "save_post", "Post saved successfully", "post saved", h.service.SavePost)
}

func (h *Handler) UnsavePost(w http.ResponseWriter, r *http.Request) {
	h.handlePostAction(w, r, "unsave_post", "Post unsaved successfully", "post unsaved", h.service.UnsavePost)
}

func (h *Handler) CreateSavedCollection(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "create_saved_collection", mapPostError)
		return
	}

	var req CreateSavedCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "create_saved_collection", mapPostError)
		return
	}

	collection, err := h.service.CreateSavedCollection(r.Context(), CreateSavedCollectionInput{
		UserID:   principal.UserID,
		Name:     req.Name,
		IsPublic: req.IsPublic,
	})
	if err != nil {
		share.WriteError(w, r, err, "create_saved_collection", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusCreated, "Saved collection created successfully", collection, r)
}

func (h *Handler) ListSavedCollections(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "list_saved_collections", mapPostError)
		return
	}

	collections, err := h.service.ListSavedCollections(r.Context(), SavedCollectionInput{UserID: principal.UserID})
	if err != nil {
		share.WriteError(w, r, err, "list_saved_collections", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Saved collections fetched successfully", collections, r)
}

func (h *Handler) ListSavedPosts(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "list_saved_posts", mapPostError)
		return
	}

	posts, err := h.service.ListSavedPosts(r.Context(), SavedCollectionInput{UserID: principal.UserID})
	if err != nil {
		share.WriteError(w, r, err, "list_saved_posts", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Saved posts fetched successfully", posts, r)
}

func (h *Handler) UpdateSavedCollectionVisibility(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "update_saved_collection", mapPostError)
		return
	}

	collectionID, err := parseRouteUint(r, "collectionID")
	if err != nil {
		share.WriteError(w, r, ErrCollectionIDRequired, "update_saved_collection", mapPostError)
		return
	}

	var req UpdateSavedCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "update_saved_collection", mapPostError)
		return
	}
	if req.IsPublic == nil {
		share.WriteError(w, r, errInvalidJSONPayload, "update_saved_collection", mapPostError)
		return
	}

	collection, err := h.service.UpdateSavedCollectionVisibility(r.Context(), UpdateSavedCollectionVisibilityInput{
		CollectionID: collectionID,
		UserID:       principal.UserID,
		IsPublic:     *req.IsPublic,
	})
	if err != nil {
		share.WriteError(w, r, err, "update_saved_collection", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Saved collection updated successfully", collection, r)
}

func (h *Handler) GetPublicSavedCollection(w http.ResponseWriter, r *http.Request) {
	collection, err := h.service.GetPublicSavedCollection(r.Context(), PublicSavedCollectionInput{
		ShareSlug:    strings.TrimSpace(chi.URLParam(r, "shareSlug")),
		ViewerUserID: h.viewerUserID(r),
	})
	if err != nil {
		share.WriteError(w, r, err, "get_public_saved_collection", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Public saved collection fetched successfully", collection, r)
}

func (h *Handler) AddPostToSavedCollection(w http.ResponseWriter, r *http.Request) {
	h.handleSavedCollectionPostAction(
		w,
		r,
		"add_saved_collection_post",
		"Post added to saved collection successfully",
		h.service.AddPostToSavedCollection,
	)
}

func (h *Handler) RemovePostFromSavedCollection(w http.ResponseWriter, r *http.Request) {
	h.handleSavedCollectionPostAction(
		w,
		r,
		"remove_saved_collection_post",
		"Post removed from saved collection successfully",
		h.service.RemovePostFromSavedCollection,
	)
}

func (h *Handler) DeleteSavedCollection(w http.ResponseWriter, r *http.Request) {
	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "delete_saved_collection", mapPostError)
		return
	}

	collectionID, err := parseRouteUint(r, "collectionID")
	if err != nil {
		share.WriteError(w, r, ErrCollectionIDRequired, "delete_saved_collection", mapPostError)
		return
	}

	if err := h.service.DeleteSavedCollection(r.Context(), SavedCollectionInput{
		CollectionID: collectionID,
		UserID:       principal.UserID,
	}); err != nil {
		share.WriteError(w, r, err, "delete_saved_collection", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Saved collection deleted successfully", nil, r)
}

func (h *Handler) handleSavedCollectionPostAction(
	w http.ResponseWriter,
	r *http.Request,
	operation string,
	successMessage string,
	action func(context.Context, SavedCollectionPostInput) error,
) {
	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, operation, mapPostError)
		return
	}

	collectionID, err := parseRouteUint(r, "collectionID")
	if err != nil {
		share.WriteError(w, r, ErrCollectionIDRequired, operation, mapPostError)
		return
	}

	postID, err := parseRouteUint(r, "postID")
	if err != nil {
		share.WriteError(w, r, errInvalidPostIDParam, operation, mapPostError)
		return
	}

	if err := action(r.Context(), SavedCollectionPostInput{
		CollectionID: collectionID,
		PostID:       postID,
		UserID:       principal.UserID,
	}); err != nil {
		share.WriteError(w, r, err, operation, mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, successMessage, nil, r)
}

func parseRouteUint(r *http.Request, name string) (uint64, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, name)), 10, 64)
	if err != nil || value == 0 {
		return 0, errInvalidPostIDParam
	}

	return value, nil
}

func parseOptionalFloat(raw string) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0
	}

	return value
}

func parseOptionalInt(raw string) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}

	return value
}

func moderationActorFromPrincipal(principal *auth.AuthenticatedUser) ModerationActor {
	actor := ModerationActor{UserID: principal.UserID}
	for _, role := range principal.Roles {
		if role == "admin" || role == "moderator" {
			actor.IsAdmin = true
			break
		}
	}

	return actor
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
		PostID:           postID,
		UserID:           principal.UserID,
		Content:          req.Content,
		ReplyToCommentID: req.ReplyToCommentID,
	})
	if err != nil {
		share.WriteError(w, r, err, "comment_post", mapPostError)
		return
	}

	if h.commentPublisher != nil {
		if err := h.commentPublisher.PublishCommentCreated(r.Context(), comment); err != nil {
			log.Warn().Err(err).Uint64("post_id", comment.PostID).Uint64("comment_id", comment.ID).Msg("comment event publish failed")
		}
	}

	h.publishCommentNotification(r.Context(), principal, comment)

	httpResponse.Success(w, http.StatusCreated, "Comment created successfully", comment, r)
}

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || postID == 0 {
		share.WriteError(w, r, errInvalidPostIDParam, "update_comment", mapPostError)
		return
	}

	commentID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "commentID")), 10, 64)
	if err != nil || commentID == 0 {
		share.WriteError(w, r, ErrCommentNotFound, "update_comment", mapPostError)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "update_comment", mapPostError)
		return
	}

	var req UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "update_comment", mapPostError)
		return
	}

	comment, err := h.service.UpdateComment(r.Context(), UpdateCommentInput{
		PostID:    postID,
		CommentID: commentID,
		UserID:    principal.UserID,
		Content:   req.Content,
	})
	if err != nil {
		share.WriteError(w, r, err, "update_comment", mapPostError)
		return
	}

	if h.commentPublisher != nil {
		if err := h.commentPublisher.PublishCommentUpdated(r.Context(), comment); err != nil {
			log.Warn().Err(err).Uint64("post_id", comment.PostID).Uint64("comment_id", comment.ID).Msg("comment event publish failed")
		}
	}

	httpResponse.Success(w, http.StatusOK, "Comment updated successfully", comment, r)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	h.handleCommentModerationAction(w, r, "delete_comment", "Comment deleted successfully", false, h.service.DeleteComment)
}

func (h *Handler) HideComment(w http.ResponseWriter, r *http.Request) {
	h.handleCommentModerationAction(w, r, "hide_comment", "Comment hidden successfully", true, h.service.HideComment)
}

func (h *Handler) ReportComment(w http.ResponseWriter, r *http.Request) {
	postID, err := parseRouteUint(r, "id")
	if err != nil {
		share.WriteError(w, r, errInvalidPostIDParam, "report_comment", mapPostError)
		return
	}
	commentID, err := parseRouteUint(r, "commentID")
	if err != nil {
		share.WriteError(w, r, ErrCommentNotFound, "report_comment", mapPostError)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "report_comment", mapPostError)
		return
	}

	var req ReportContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "report_comment", mapPostError)
		return
	}

	if err := h.service.ReportComment(r.Context(), ReportInput{
		PostID:         postID,
		CommentID:      commentID,
		ReporterUserID: principal.UserID,
		Reason:         req.Reason,
	}); err != nil {
		share.WriteError(w, r, err, "report_comment", mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusCreated, "Comment reported successfully", nil, r)
}

func (h *Handler) handlePostModerationAction(
	w http.ResponseWriter,
	r *http.Request,
	operation string,
	successMessage string,
	needsReason bool,
	action func(context.Context, ModerationInput) error,
	afterSuccess func(context.Context, uint64),
) {
	postID, err := parseRouteUint(r, "id")
	if err != nil {
		share.WriteError(w, r, errInvalidPostIDParam, operation, mapPostError)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, operation, mapPostError)
		return
	}

	reason, ok := h.readModerationReason(w, r, operation, needsReason)
	if !ok {
		return
	}

	if err := action(r.Context(), ModerationInput{
		PostID: postID,
		Actor:  moderationActorFromPrincipal(principal),
		Reason: reason,
	}); err != nil {
		share.WriteError(w, r, err, operation, mapPostError)
		return
	}
	if afterSuccess != nil {
		afterSuccess(r.Context(), postID)
	}

	httpResponse.Success(w, http.StatusOK, successMessage, nil, r)
}

func (h *Handler) handleCommentModerationAction(
	w http.ResponseWriter,
	r *http.Request,
	operation string,
	successMessage string,
	needsReason bool,
	action func(context.Context, ModerationInput) error,
) {
	postID, err := parseRouteUint(r, "id")
	if err != nil {
		share.WriteError(w, r, errInvalidPostIDParam, operation, mapPostError)
		return
	}
	commentID, err := parseRouteUint(r, "commentID")
	if err != nil {
		share.WriteError(w, r, ErrCommentNotFound, operation, mapPostError)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, operation, mapPostError)
		return
	}

	reason, ok := h.readModerationReason(w, r, operation, needsReason)
	if !ok {
		return
	}

	if err := action(r.Context(), ModerationInput{
		PostID:    postID,
		CommentID: commentID,
		Actor:     moderationActorFromPrincipal(principal),
		Reason:    reason,
	}); err != nil {
		share.WriteError(w, r, err, operation, mapPostError)
		return
	}

	httpResponse.Success(w, http.StatusOK, successMessage, nil, r)
}

func (h *Handler) readModerationReason(w http.ResponseWriter, r *http.Request, operation string, required bool) (string, bool) {
	if !required {
		return "", true
	}

	var req ModerateContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, operation, mapPostError)
		return "", false
	}

	return req.Reason, true
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
		return err
	}
	if _, err := w.Write([]byte("data: ")); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if _, err := w.Write([]byte("\n\n")); err != nil {
		return err
	}

	flusher.Flush()
	return nil
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

func (h *Handler) viewerUserID(r *http.Request) uint64 {
	if h.authService == nil {
		return 0
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return 0
	}

	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return 0
	}

	principal, err := h.authService.Authenticate(r.Context(), strings.TrimSpace(token))
	if err != nil || principal == nil {
		return 0
	}

	return principal.UserID
}

func (h *Handler) publishCommentNotification(ctx context.Context, principal *auth.AuthenticatedUser, comment CommentView) {
	if h.notifications == nil || principal == nil || principal.UserID == 0 {
		return
	}

	item, err := h.service.GetPostDetail(ctx, GetPostDetailInput{PostID: comment.PostID})
	if err != nil || item == nil {
		log.Warn().Err(err).Uint64("post_id", comment.PostID).Uint64("comment_id", comment.ID).Msg("comment notification post lookup failed")
		return
	}
	if item.UserID == 0 || item.UserID == principal.UserID {
		item.UserID = 0
	}

	actorName := strings.TrimSpace(principal.Username)
	if actorName == "" {
		actorName = "Someone"
	}

	recipients := make(map[uint64]string)
	if item.UserID != 0 {
		recipients[item.UserID] = "New comment"
	}
	if comment.ReplyToUserID != 0 && comment.ReplyToUserID != principal.UserID {
		recipients[comment.ReplyToUserID] = "New reply"
	}

	for userID, title := range recipients {
		if err := h.notifications.Publish(ctx, notification.Notification{
			UserID:      userID,
			ActorUserID: principal.UserID,
			ActorName:   actorName,
			Type:        notification.TypePostCommented,
			Title:       title,
			Body:        actorName + ": " + commentNotificationSnippet(comment.Content),
			Resource:    notification.ResourceComment,
			ResourceID:  notification.ResourceIDUint64(comment.ID),
			PostID:      comment.PostID,
		}); err != nil {
			log.Warn().Err(err).Uint64("post_id", comment.PostID).Uint64("comment_id", comment.ID).Uint64("user_id", userID).Msg("comment notification publish failed")
		}
	}
}

func (h *Handler) publishFollowerPostNotifications(ctx context.Context, principal *auth.AuthenticatedUser, item PostView) {
	if h.notifications == nil || h.followers == nil || principal == nil || principal.UserID == 0 || item.ID == 0 {
		return
	}

	followerIDs, err := h.followers.ListFollowerIDs(ctx, principal.UserID)
	if err != nil {
		log.Warn().Err(err).Uint64("user_id", principal.UserID).Uint64("post_id", item.ID).Msg("post follower lookup failed")
		return
	}

	actorName := strings.TrimSpace(principal.Username)
	if actorName == "" {
		actorName = strings.TrimSpace(item.UserName)
	}
	if actorName == "" {
		actorName = "Someone"
	}

	detail := strings.TrimSpace(item.Caption)
	if detail == "" {
		detail = strings.TrimSpace(item.LocationName)
	}
	body := actorName + " uploaded a new post."
	if detail != "" {
		body = actorName + " uploaded " + detail + "."
	}

	for _, userID := range followerIDs {
		if userID == 0 || userID == principal.UserID {
			continue
		}
		if err := h.notifications.Publish(ctx, notification.Notification{
			UserID:      userID,
			ActorUserID: principal.UserID,
			ActorName:   actorName,
			Type:        notification.TypePostCreated,
			Title:       "New upload",
			Body:        body,
			Resource:    notification.ResourcePost,
			ResourceID:  notification.ResourceIDUint64(item.ID),
			PostID:      item.ID,
		}); err != nil {
			log.Warn().Err(err).Uint64("post_id", item.ID).Uint64("user_id", userID).Msg("post follower notification publish failed")
		}
	}
}

func (h *Handler) publishPostDeleted(ctx context.Context, postID uint64) {
	if h.postPublisher == nil || postID == 0 {
		return
	}
	if err := h.postPublisher.PublishPostDeleted(ctx, postID); err != nil {
		log.Warn().Err(err).Uint64("post_id", postID).Msg("post deleted event publish failed")
	}
}

func commentNotificationSnippet(content string) string {
	value := strings.Join(strings.Fields(content), " ")
	if len(value) <= 120 {
		return value
	}

	return value[:117] + "..."
}
