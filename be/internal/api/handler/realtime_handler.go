package handler

import (
	"be/internal/api/http/response"
	apimiddleware "be/internal/api/middleware"
	roomapp "be/internal/application/room"
	domainroom "be/internal/domain/room"
	"be/internal/realtime"
	"be/internal/shared/apperror"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5"
)

const (
	webSocketReadLimit = 4 * 1024
	webSocketWriteWait = 5 * time.Second
	webSocketPingEvery = 25 * time.Second
)

type RealtimeHandler struct {
	getRoom        *roomapp.GetRoomUseCase
	hub            *realtime.Hub
	originPatterns []string
	logger         *slog.Logger
}

func NewRealtimeHandler(
	getRoom *roomapp.GetRoomUseCase,
	hub *realtime.Hub,
	originPatterns []string,
	logger *slog.Logger,
) *RealtimeHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &RealtimeHandler{
		getRoom:        getRoom,
		hub:            hub,
		originPatterns: originPatterns,
		logger:         logger,
	}
}

type clientRealtimeEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type sendChatPayload struct {
	Text string `json:"text"`
}

func (h *RealtimeHandler) ConnectRoom(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}

	room, err := h.getRoom.Execute(r.Context(), chi.URLParam(r, "roomID"))
	if err != nil {
		response.Error(w, mapRoomError(err))
		return
	}
	if !isRoomMember(room, principal.UserID) {
		response.Error(w, apperror.Forbidden("Only room members can connect"))
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:   []string{realtime.Subprotocol},
		OriginPatterns: h.originPatterns,
	})
	if err != nil {
		h.logger.WarnContext(r.Context(), "websocket upgrade failed",
			slog.String("room_id", room.ID),
			slog.Int64("user_id", principal.UserID),
			slog.Any("error", err),
		)
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(webSocketReadLimit)

	client, err := h.hub.Register(room.ID, principal.UserID, principal.UserName, mapRealtimeMembers(room))
	if err != nil {
		_ = conn.Close(websocket.StatusTryAgainLater, "Realtime service unavailable")
		return
	}
	defer h.hub.Unregister(client)

	h.logger.InfoContext(r.Context(), "websocket client connected",
		slog.String("room_id", room.ID),
		slog.Int64("user_id", principal.UserID),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 2)
	go func() { errCh <- h.readLoop(ctx, conn, client) }()
	go func() { errCh <- h.writeLoop(ctx, conn, client) }()

	connectionErr := <-errCh
	cancel()
	_ = conn.CloseNow()

	status := websocket.CloseStatus(connectionErr)
	level := slog.LevelWarn
	if status == websocket.StatusNormalClosure ||
		status == websocket.StatusGoingAway ||
		errors.Is(connectionErr, context.Canceled) ||
		errors.Is(connectionErr, realtime.ErrHubClosed) {
		level = slog.LevelInfo
	}
	h.logger.LogAttrs(r.Context(), level, "websocket client disconnected",
		slog.String("room_id", room.ID),
		slog.Int64("user_id", principal.UserID),
		slog.Int("close_status", int(status)),
		slog.Any("error", connectionErr),
	)
}

func (h *RealtimeHandler) readLoop(ctx context.Context, conn *websocket.Conn, client *realtime.Client) error {
	for {
		var event clientRealtimeEvent
		if err := wsjson.Read(ctx, conn, &event); err != nil {
			return err
		}

		switch event.Type {
		case "chat.send":
			var payload sendChatPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				h.hub.SendError(client, "INVALID_PAYLOAD", "Chat payload is invalid")
				continue
			}
			if err := h.hub.PublishChat(client, payload.Text); err != nil {
				switch {
				case errors.Is(err, realtime.ErrEmptyMessage):
					h.hub.SendError(client, "EMPTY_MESSAGE", "Chat message is required")
				case errors.Is(err, realtime.ErrMessageTooLong):
					h.hub.SendError(client, "MESSAGE_TOO_LONG", "Chat message must be at most 500 characters")
				case errors.Is(err, realtime.ErrRateLimited):
					h.hub.SendError(client, "RATE_LIMITED", "Send at most 10 messages every 5 seconds")
				default:
					return err
				}
			}
		default:
			h.hub.SendError(client, "UNSUPPORTED_EVENT", "Unsupported realtime event")
		}
	}
}

func (h *RealtimeHandler) writeLoop(ctx context.Context, conn *websocket.Conn, client *realtime.Client) error {
	ticker := time.NewTicker(webSocketPingEvery)
	defer ticker.Stop()

	for {
		select {
		case event := <-client.Events():
			writeCtx, cancel := context.WithTimeout(ctx, webSocketWriteWait)
			err := wsjson.Write(writeCtx, conn, event)
			cancel()
			if err != nil {
				return err
			}
		case <-ticker.C:
			h.hub.Heartbeat(client)
			pingCtx, cancel := context.WithTimeout(ctx, webSocketWriteWait)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return err
			}
		case <-client.Done():
			return realtime.ErrHubClosed
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func isRoomMember(room *domainroom.Room, userID int64) bool {
	for _, member := range room.Members {
		if member.UserID == userID {
			return true
		}
	}
	return false
}

func mapRealtimeMembers(room *domainroom.Room) []realtime.Member {
	members := make([]realtime.Member, 0, len(room.Members))
	for _, member := range room.Members {
		members = append(members, realtime.Member{
			ID:         member.UserID,
			Name:       member.UserName,
			SeatNumber: member.SeatNumber,
			Host:       member.IsHost,
		})
	}
	return members
}
