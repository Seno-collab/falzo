package api

import (
	"be/internal/api/handler"
	apimiddleware "be/internal/api/middleware"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	roomHandler *handler.RoomHandler,
	realtimeHandler *handler.RealtimeHandler,
	socialHandler *handler.SocialHandler,
	authenticator *apimiddleware.Authenticator,
	logger *slog.Logger,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(apimiddleware.RequestLogger(logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(timeoutUnlessWebSocket(30 * time.Second))
	r.Get("/health/live", healthHandler.Live)
	r.Get("/health/ready", healthHandler.Ready)
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
			r.Post("/logout", authHandler.Logout)
			r.Post("/google/credential", authHandler.GoogleLogin)
		})
		r.Group(func(r chi.Router) {
			r.Use(authenticator.RequireAccessToken)
			r.Post("/rooms", roomHandler.Create)
			r.Get("/rooms", roomHandler.List)
			r.Post("/rooms/join", roomHandler.Join)
			r.Get("/rooms/{roomID}", roomHandler.Get)
			r.Post("/rooms/{roomID}/rounds", roomHandler.DealRound)
			r.Get("/rooms/{roomID}/rounds/current/card", roomHandler.GetCurrentCard)
			r.Patch("/rooms/{roomID}/settings/discussion", roomHandler.UpdateDiscussion)
			r.Get("/rooms/{roomID}/rounds/current", roomHandler.GetRoundState)
			r.Post("/rooms/{roomID}/rounds/current/ready", roomHandler.PlayerReady)
			r.Post("/rooms/{roomID}/rounds/current/turn/finish", roomHandler.FinishTurn)
			r.Post("/rooms/{roomID}/rounds/current/votes", roomHandler.CastVote)
			r.Post("/rooms/{roomID}/rounds/current/mr-white/guess", roomHandler.MrWhiteGuess)

			r.Get("/users/search", socialHandler.SearchUsers)
			r.Post("/friend-requests", socialHandler.SendFriendRequest)
			r.Get("/friend-requests", socialHandler.ListFriendRequests)
			r.Post("/friend-requests/{requestID}/accept", socialHandler.AcceptFriendRequest)
			r.Post("/friend-requests/{requestID}/reject", socialHandler.RejectFriendRequest)
			r.Delete("/friend-requests/{requestID}", socialHandler.CancelFriendRequest)
			r.Get("/friends", socialHandler.ListFriends)
			r.Delete("/friends/{friendUserID}", socialHandler.Unfriend)
			r.Get("/notifications", socialHandler.ListNotifications)
			r.Get("/notifications/unread-count", socialHandler.CountUnreadNotifications)
			r.Patch("/notifications/read-all", socialHandler.MarkAllNotificationsRead)
			r.Patch("/notifications/{notificationID}/read", socialHandler.MarkNotificationRead)
		})
		r.Group(func(r chi.Router) {
			r.Use(authenticator.RequireWebSocketAccessToken)
			r.Get("/rooms/{roomID}/ws", realtimeHandler.ConnectRoom)
		})
	})
	return r
}

func timeoutUnlessWebSocket(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		withTimeout := chimiddleware.Timeout(timeout)(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") && headerContainsToken(r.Header.Get("Connection"), "upgrade") {
				next.ServeHTTP(w, r)
				return
			}
			withTimeout.ServeHTTP(w, r)
		})
	}
}

func headerContainsToken(header, target string) bool {
	for token := range strings.SplitSeq(header, ",") {
		if strings.EqualFold(strings.TrimSpace(token), target) {
			return true
		}
	}
	return false
}
