package handler

import (
	"be/internal/api/http/request"
	"be/internal/api/http/response"
	apimiddleware "be/internal/api/middleware"
	socialapp "be/internal/application/social"
	domainsocial "be/internal/domain/social"
	"be/internal/realtime"
	"be/internal/shared/apperror"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type SocialHandler struct {
	service  *socialapp.Service
	logger   *slog.Logger
	realtime *realtime.Hub
}

func NewSocialHandler(service *socialapp.Service, logger *slog.Logger, realtimeHub *realtime.Hub) *SocialHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &SocialHandler{service: service, logger: logger, realtime: realtimeHub}
}

type sendFriendRequestBody struct {
	ReceiverID int64 `json:"receiver_id" validate:"required,gt=0"`
}

type socialUserResponse struct {
	ID           int64                           `json:"id"`
	UserName     string                          `json:"username"`
	Relationship domainsocial.RelationshipStatus `json:"relationship"`
	Online       bool                            `json:"online"`
}

type friendRequestResponse struct {
	ID           int64                            `json:"id"`
	SenderID     int64                            `json:"sender_id"`
	SenderName   string                           `json:"sender_name"`
	ReceiverID   int64                            `json:"receiver_id"`
	ReceiverName string                           `json:"receiver_name"`
	Status       domainsocial.FriendRequestStatus `json:"status"`
	CreatedAt    time.Time                        `json:"created_at"`
	RespondedAt  *time.Time                       `json:"responded_at,omitempty"`
}

type friendResponse struct {
	ID        int64     `json:"id"`
	UserName  string    `json:"username"`
	FriendsAt time.Time `json:"friends_at"`
	Online    bool      `json:"online"`
}

type notificationResponse struct {
	ID          int64                         `json:"id"`
	Type        domainsocial.NotificationType `json:"type"`
	ActorID     int64                         `json:"actor_id"`
	ActorName   string                        `json:"actor_name"`
	ReferenceID int64                         `json:"reference_id"`
	Read        bool                          `json:"read"`
	ReadAt      *time.Time                    `json:"read_at,omitempty"`
	CreatedAt   time.Time                     `json:"created_at"`
}

func (h *SocialHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	limit, err := positiveQueryInt(r, "limit", socialapp.DefaultUserSearchLimit)
	if err != nil {
		response.Error(w, err)
		return
	}
	users, err := h.service.SearchUsers(r.Context(), principal.UserID, r.URL.Query().Get("q"), limit)
	if err != nil {
		h.writeError(w, r, "search_users", err)
		return
	}
	onlineUsers := h.onlineUsers(r)
	data := make([]socialUserResponse, 0, len(users))
	for _, user := range users {
		data = append(data, socialUserResponse{
			ID:           user.ID,
			UserName:     user.UserName,
			Relationship: user.Relationship,
			Online:       onlineUsers[user.ID],
		})
	}
	response.OK(w, data)
}

func (h *SocialHandler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	body, err := request.DecodeJSON[sendFriendRequestBody](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	friendRequest, err := h.service.SendFriendRequest(r.Context(), principal.UserID, body.ReceiverID)
	if err != nil {
		h.writeError(w, r, "send_friend_request", err)
		return
	}
	h.publishSocialUpdated(friendRequest.ReceiverID, "friend_request_received")
	response.Created(w, mapFriendRequest(friendRequest))
}

func (h *SocialHandler) ListFriendRequests(w http.ResponseWriter, r *http.Request) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	requests, err := h.service.ListPendingFriendRequests(r.Context(), principal.UserID)
	if err != nil {
		h.writeError(w, r, "list_friend_requests", err)
		return
	}
	data := make([]friendRequestResponse, 0, len(requests))
	for i := range requests {
		data = append(data, mapFriendRequest(&requests[i]))
	}
	response.OK(w, data)
}

func (h *SocialHandler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	h.respondFriendRequest(w, r, true)
}

func (h *SocialHandler) RejectFriendRequest(w http.ResponseWriter, r *http.Request) {
	h.respondFriendRequest(w, r, false)
}

func (h *SocialHandler) respondFriendRequest(w http.ResponseWriter, r *http.Request, accept bool) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	requestID, err := positivePathID(r, "requestID")
	if err != nil {
		response.Error(w, err)
		return
	}

	var friendRequest *domainsocial.FriendRequest
	operation := "reject_friend_request"
	if accept {
		operation = "accept_friend_request"
		friendRequest, err = h.service.AcceptFriendRequest(r.Context(), principal.UserID, requestID)
	} else {
		friendRequest, err = h.service.RejectFriendRequest(r.Context(), principal.UserID, requestID)
	}
	if err != nil {
		h.writeError(w, r, operation, err)
		return
	}
	reason := "friend_request_rejected"
	if accept {
		reason = "friend_request_accepted"
	}
	h.publishSocialUpdated(friendRequest.SenderID, reason)
	h.publishSocialUpdated(friendRequest.ReceiverID, reason)
	response.OK(w, mapFriendRequest(friendRequest))
}

func (h *SocialHandler) CancelFriendRequest(w http.ResponseWriter, r *http.Request) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	requestID, err := positivePathID(r, "requestID")
	if err != nil {
		response.Error(w, err)
		return
	}
	if err := h.service.CancelFriendRequest(r.Context(), principal.UserID, requestID); err != nil {
		h.writeError(w, r, "cancel_friend_request", err)
		return
	}
	response.NoContent(w)
}

func (h *SocialHandler) ListFriends(w http.ResponseWriter, r *http.Request) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	friends, err := h.service.ListFriends(r.Context(), principal.UserID)
	if err != nil {
		h.writeError(w, r, "list_friends", err)
		return
	}
	onlineUsers := h.onlineUsers(r)
	data := make([]friendResponse, 0, len(friends))
	for _, friend := range friends {
		data = append(data, friendResponse{
			ID:        friend.ID,
			UserName:  friend.UserName,
			FriendsAt: friend.FriendsAt,
			Online:    onlineUsers[friend.ID],
		})
	}
	response.OK(w, data)
}

func (h *SocialHandler) onlineUsers(r *http.Request) map[int64]bool {
	if h.realtime == nil {
		return map[int64]bool{}
	}
	onlineUsers, err := h.realtime.OnlineUserIDs(r.Context())
	if err != nil {
		h.logger.WarnContext(r.Context(), "could not load online users", slog.Any("error", err))
		return map[int64]bool{}
	}
	return onlineUsers
}

func (h *SocialHandler) Unfriend(w http.ResponseWriter, r *http.Request) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	friendUserID, err := positivePathID(r, "friendUserID")
	if err != nil {
		response.Error(w, err)
		return
	}
	if err := h.service.Unfriend(r.Context(), principal.UserID, friendUserID); err != nil {
		h.writeError(w, r, "unfriend", err)
		return
	}
	h.publishSocialUpdated(friendUserID, "friend_removed")
	response.NoContent(w)
}

func (h *SocialHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	limit, err := positiveQueryInt(r, "limit", socialapp.DefaultNotificationLimit)
	if err != nil {
		response.Error(w, err)
		return
	}
	offset, err := nonNegativeQueryInt(r, "offset", 0)
	if err != nil {
		response.Error(w, err)
		return
	}
	unreadOnly, err := boolQuery(r, "unread_only", false)
	if err != nil {
		response.Error(w, err)
		return
	}
	notifications, err := h.service.ListNotifications(r.Context(), principal.UserID, unreadOnly, limit, offset)
	if err != nil {
		h.writeError(w, r, "list_notifications", err)
		return
	}
	data := make([]notificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		data = append(data, mapNotification(notification))
	}
	response.OK(w, data)
}

func (h *SocialHandler) CountUnreadNotifications(w http.ResponseWriter, r *http.Request) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	count, err := h.service.CountUnreadNotifications(r.Context(), principal.UserID)
	if err != nil {
		h.writeError(w, r, "count_unread_notifications", err)
		return
	}
	response.OK(w, struct {
		Count int `json:"count"`
	}{Count: count})
}

func (h *SocialHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	notificationID, err := positivePathID(r, "notificationID")
	if err != nil {
		response.Error(w, err)
		return
	}
	if err := h.service.MarkNotificationRead(r.Context(), principal.UserID, notificationID); err != nil {
		h.writeError(w, r, "mark_notification_read", err)
		return
	}
	h.publishSocialUpdated(principal.UserID, "notification_read")
	response.NoContent(w)
}

func (h *SocialHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	principal, ok := socialPrincipal(w, r)
	if !ok {
		return
	}
	count, err := h.service.MarkAllNotificationsRead(r.Context(), principal.UserID)
	if err != nil {
		h.writeError(w, r, "mark_all_notifications_read", err)
		return
	}
	h.publishSocialUpdated(principal.UserID, "notifications_read_all")
	response.OK(w, struct {
		Updated int64 `json:"updated"`
	}{Updated: count})
}

func (h *SocialHandler) publishSocialUpdated(userID int64, reason string) {
	if h.realtime != nil {
		h.realtime.PublishUser(userID, realtime.EventSocialUpdated, map[string]string{"reason": reason})
	}
}

func (h *SocialHandler) writeError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	mappedErr := mapSocialError(err)
	appErr := apperror.FromError(mappedErr)
	level := slog.LevelWarn
	if appErr.Code == apperror.CodeInternalServerError {
		level = slog.LevelError
	}
	h.logger.LogAttrs(r.Context(), level, "social operation failed",
		slog.String("operation", operation),
		slog.String("code", string(appErr.Code)),
		slog.Any("error", err),
	)
	response.Error(w, mappedErr)
}

func mapSocialError(err error) error {
	switch {
	case errors.Is(err, domainsocial.ErrCannotFriendSelf),
		errors.Is(err, domainsocial.ErrSearchQueryTooShort):
		return apperror.InvalidRequest(err.Error())
	case errors.Is(err, domainsocial.ErrUserNotFound):
		return apperror.NotFound("User not found")
	case errors.Is(err, domainsocial.ErrAlreadyFriends):
		return apperror.Conflict("Users are already friends")
	case errors.Is(err, domainsocial.ErrFriendRequestExists):
		return apperror.Conflict("A pending friend request already exists")
	case errors.Is(err, domainsocial.ErrFriendRequestNotPending):
		return apperror.Conflict("Friend request is no longer pending")
	case errors.Is(err, domainsocial.ErrFriendRequestForbidden):
		return apperror.Forbidden("You cannot perform this action on the friend request")
	case errors.Is(err, domainsocial.ErrFriendRequestNotFound):
		return apperror.NotFound("Friend request not found")
	case errors.Is(err, domainsocial.ErrFriendshipNotFound):
		return apperror.NotFound("Friendship not found")
	case errors.Is(err, domainsocial.ErrNotificationNotFound):
		return apperror.NotFound("Notification not found")
	default:
		return apperror.Internal(err)
	}
}

func mapFriendRequest(request *domainsocial.FriendRequest) friendRequestResponse {
	return friendRequestResponse{
		ID:           request.ID,
		SenderID:     request.SenderID,
		SenderName:   request.SenderName,
		ReceiverID:   request.ReceiverID,
		ReceiverName: request.ReceiverName,
		Status:       request.Status,
		CreatedAt:    request.CreatedAt,
		RespondedAt:  request.RespondedAt,
	}
}

func mapNotification(notification domainsocial.Notification) notificationResponse {
	return notificationResponse{
		ID:          notification.ID,
		Type:        notification.Type,
		ActorID:     notification.ActorID,
		ActorName:   notification.ActorName,
		ReferenceID: notification.ReferenceID,
		Read:        notification.ReadAt != nil,
		ReadAt:      notification.ReadAt,
		CreatedAt:   notification.CreatedAt,
	}
}

func socialPrincipal(w http.ResponseWriter, r *http.Request) (apimiddleware.Principal, bool) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
	}
	return principal, ok
}

func positivePathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		return 0, apperror.InvalidRequest(name + " must be a positive integer")
	}
	return id, nil
}

func positiveQueryInt(r *http.Request, name string, fallback int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, apperror.InvalidRequest(name + " must be a positive integer")
	}
	return parsed, nil
}

func nonNegativeQueryInt(r *http.Request, name string, fallback int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, apperror.InvalidRequest(name + " must be a non-negative integer")
	}
	return parsed, nil
}

func boolQuery(r *http.Request, name string, fallback bool) (bool, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, apperror.InvalidRequest(name + " must be true or false")
	}
	return parsed, nil
}
