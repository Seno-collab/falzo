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
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type handlerService interface {
	CreatePost(ctx context.Context, input CreatePostInput) (PostView, error)
	LikePost(ctx context.Context, input PostActionInput) error
	UnlikePost(ctx context.Context, input PostActionInput) error
	SavePost(ctx context.Context, input PostActionInput) error
	UnsavePost(ctx context.Context, input PostActionInput) error
	CommentPost(ctx context.Context, input CommentPostInput) (CommentView, error)
	UpdateComment(ctx context.Context, input UpdateCommentInput) (CommentView, error)
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
	commentEvents    CommentEventSubscriber
	commentPublisher CommentEventPublisher
	postEvents       PostEventSubscriber
	postPublisher    PostEventPublisher
}

type HandlerOption func(*Handler)

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
	r.Get("/{id}/comments/events", h.StreamComments)
	r.Get("/{id}/comments", h.GetComments)
	r.Get("/{id}", h.GetPostDetail)
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth(h.authService))
		protected.Post("/", h.CreatePost)
		protected.Post("/{id}/like", h.LikePost)
		protected.Delete("/{id}/like", h.UnlikePost)
		protected.Post("/{id}/save", h.SavePost)
		protected.Delete("/{id}/save", h.UnsavePost)
		protected.Post("/{id}/comments", h.CommentPost)
		protected.Put("/{id}/comments/{commentID}", h.UpdateComment)
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
	Content          string `json:"content"`
	ReplyToCommentID uint64 `json:"reply_to_comment_id"`
}

type UpdateCommentRequest struct {
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

	if h.postPublisher != nil {
		if err := h.postPublisher.PublishPostCreated(r.Context(), post); err != nil {
			log.Warn().Err(err).Uint64("post_id", post.ID).Msg("post event publish failed")
		}
	}

	httpResponse.Success(w, http.StatusCreated, "Post created successfully", post, r)
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
			if comment.PostID != postID {
				continue
			}
			if err := writeSSE(w, flusher, "comment.created", comment); err != nil {
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

	posts, err := h.service.GetPosts(r.Context(), ListPostsInput{
		Page:         page,
		Limit:        limit,
		ViewerUserID: h.viewerUserID(r),
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
		case post, ok := <-posts:
			if !ok {
				return
			}
			if err := writeSSE(w, flusher, "post.created", post); err != nil {
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
		if err := h.commentPublisher.PublishCommentCreated(r.Context(), comment); err != nil {
			log.Warn().Err(err).Uint64("post_id", comment.PostID).Uint64("comment_id", comment.ID).Msg("comment event publish failed")
		}
	}

	httpResponse.Success(w, http.StatusOK, "Comment updated successfully", comment, r)
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
