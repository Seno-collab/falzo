package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"falzo-be/internal/auth"
	"falzo-be/internal/share"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	subscriber  Subscriber
	lister      Lister
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	}
}

func NewHandler(
	subscriber Subscriber,
	authService interface {
		Authenticate(ctx context.Context, rawToken string) (*auth.AuthenticatedUser, error)
	},
) *Handler {
	h := &Handler{subscriber: subscriber, authService: authService}
	if lister, ok := subscriber.(Lister); ok {
		h.lister = lister
	}
	return h
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Group(func(protected chi.Router) {
		protected.Use(auth.RequireAuth(h.authService))
		protected.Get("/", h.ListNotifications)
		protected.Get("/events", h.StreamNotifications)
	})
	return r
}

func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	if h.lister == nil {
		httpResponse.Error(w, http.StatusServiceUnavailable, "notification service unavailable", r)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "list_notifications", mapNotificationError)
		return
	}

	limit := 30
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			share.WriteError(w, r, ErrInvalidLimit, "list_notifications", mapNotificationError)
			return
		}
		limit = parsed
	}
	if limit > 100 {
		limit = 100
	}

	items, err := h.lister.ListByUser(r.Context(), principal.UserID, limit)
	if err != nil {
		share.WriteError(w, r, err, "list_notifications", mapNotificationError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Notifications fetched successfully", items, r)
}

func (h *Handler) StreamNotifications(w http.ResponseWriter, r *http.Request) {
	if h.subscriber == nil {
		httpResponse.Error(w, http.StatusServiceUnavailable, "notification service unavailable", r)
		return
	}

	principal, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, ErrUserIDRequired, "stream_notifications", mapNotificationError)
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

	notifications, unsubscribe, err := h.subscriber.Subscribe(r.Context(), principal.UserID)
	if err != nil {
		share.WriteError(w, r, err, "stream_notifications", mapNotificationError)
		return
	}
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
		case item, ok := <-notifications:
			if !ok {
				return
			}
			if err := writeNotificationSSE(w, flusher, item); err != nil {
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

func writeNotificationSSE(w http.ResponseWriter, flusher http.Flusher, item Notification) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprint(w, "event: notification.created\n"); err != nil {
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

func mapNotificationError(err error) share.ApiError {
	if errors.Is(err, ErrUserIDRequired) {
		return share.Required("user_id", "user id is required")
	}
	if errors.Is(err, ErrInvalidLimit) {
		return share.BadRequest("limit", "limit must be greater than 0")
	}
	if errors.Is(err, ErrDependencyUnavailable) {
		return share.ServiceUnavailable("Notification service unavailable", "Please try again later")
	}

	return share.Internal()
}
