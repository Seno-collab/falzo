package social

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"falzo-be/internal/auth"
	"falzo-be/internal/notification"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type handlerService interface {
	GetPublicProfile(ctx context.Context, input PublicProfileInput) (PublicProfile, error)
	Follow(ctx context.Context, input FollowInput) (bool, error)
	Unfollow(ctx context.Context, input FollowInput) error
	Block(ctx context.Context, input FollowInput) error
	Unblock(ctx context.Context, input FollowInput) error
}

type Handler struct {
	service     handlerService
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	}
	notifications notification.Publisher
}

type HandlerOption func(*Handler)

func WithNotifications(publisher notification.Publisher) HandlerOption {
	return func(h *Handler) {
		h.notifications = publisher
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
	r.Get("/{id}", h.GetPublicProfile)
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth(h.authService))
		protected.Post("/{id}/follow", h.Follow)
		protected.Delete("/{id}/follow", h.Unfollow)
		protected.Post("/{id}/block", h.Block)
		protected.Delete("/{id}/block", h.Unblock)
	})
	return r
}

func (h *Handler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.parseUserID(w, r, "get_public_profile")
	if !ok {
		return
	}

	profile, err := h.service.GetPublicProfile(r.Context(), PublicProfileInput{
		UserID:       userID,
		ViewerUserID: h.viewerUserID(r),
	})
	if err != nil {
		share.WriteError(w, r, err, "get_public_profile", mapSocialError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "User profile fetched successfully", profile, r)
}

func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
	targetUserID, ok := h.parseUserID(w, r, "follow_user")
	if !ok {
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "follow_user", mapSocialError)
		return
	}

	created, err := h.service.Follow(r.Context(), FollowInput{
		FollowerID:  principal.UserID,
		FollowingID: targetUserID,
	})
	if err != nil {
		share.WriteError(w, r, err, "follow_user", mapSocialError)
		return
	}
	if created {
		h.publishFollowNotification(r.Context(), principal, targetUserID)
	}

	httpResponse.Success(w, http.StatusOK, "User followed successfully", map[string]bool{"is_following": true}, r)
}

func (h *Handler) Unfollow(w http.ResponseWriter, r *http.Request) {
	targetUserID, ok := h.parseUserID(w, r, "unfollow_user")
	if !ok {
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "unfollow_user", mapSocialError)
		return
	}

	if err := h.service.Unfollow(r.Context(), FollowInput{
		FollowerID:  principal.UserID,
		FollowingID: targetUserID,
	}); err != nil {
		share.WriteError(w, r, err, "unfollow_user", mapSocialError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "User unfollowed successfully", map[string]bool{"is_following": false}, r)
}

func (h *Handler) Block(w http.ResponseWriter, r *http.Request) {
	targetUserID, ok := h.parseUserID(w, r, "block_user")
	if !ok {
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "block_user", mapSocialError)
		return
	}

	if err := h.service.Block(r.Context(), FollowInput{
		FollowerID:  principal.UserID,
		FollowingID: targetUserID,
	}); err != nil {
		share.WriteError(w, r, err, "block_user", mapSocialError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "User blocked successfully", map[string]bool{"is_blocked": true}, r)
}

func (h *Handler) Unblock(w http.ResponseWriter, r *http.Request) {
	targetUserID, ok := h.parseUserID(w, r, "unblock_user")
	if !ok {
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "unblock_user", mapSocialError)
		return
	}

	if err := h.service.Unblock(r.Context(), FollowInput{
		FollowerID:  principal.UserID,
		FollowingID: targetUserID,
	}); err != nil {
		share.WriteError(w, r, err, "unblock_user", mapSocialError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "User unblocked successfully", map[string]bool{"is_blocked": false}, r)
}

func (h *Handler) parseUserID(w http.ResponseWriter, r *http.Request, operation string) (uint64, bool) {
	userID, err := strconv.ParseUint(strings.TrimSpace(chi.URLParam(r, "id")), 10, 64)
	if err != nil || userID == 0 {
		share.WriteError(w, r, errInvalidUserIDParam, operation, mapSocialError)
		return 0, false
	}

	return userID, true
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

func (h *Handler) publishFollowNotification(ctx context.Context, principal *auth.AuthenticatedUser, targetUserID uint64) {
	if h.notifications == nil || principal == nil || principal.UserID == 0 || targetUserID == 0 {
		return
	}

	actorName := strings.TrimSpace(principal.Username)
	if actorName == "" {
		actorName = "Someone"
	}

	if err := h.notifications.Publish(ctx, notification.Notification{
		UserID:      targetUserID,
		ActorUserID: principal.UserID,
		ActorName:   actorName,
		Type:        notification.TypeUserFollowed,
		Title:       "New follower",
		Body:        actorName + " followed you.",
		Resource:    notification.ResourceUser,
		ResourceID:  notification.ResourceIDUint64(principal.UserID),
	}); err != nil {
		log.Warn().Err(err).Uint64("target_user_id", targetUserID).Uint64("actor_user_id", principal.UserID).Msg("follow notification publish failed")
	}
}
