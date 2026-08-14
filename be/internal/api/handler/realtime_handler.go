package handler

import (
	"be/internal/api/http/response"
	apimiddleware "be/internal/api/middleware"
	roomapp "be/internal/application/room"
	domainroom "be/internal/domain/room"
	"be/internal/observability"
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
	webSocketReadLimit                           = 4 * 1024
	webSocketWriteWait                           = 5 * time.Second
	webSocketPingEvery                           = 25 * time.Second
	webSocketCleanupWait                         = time.Second
	webSocketReplacedStatus websocket.StatusCode = 4009
)

type RealtimeHandler struct {
	getRoom        *roomapp.GetRoomUseCase
	getRoundState  *roomapp.GetRoundStateUseCase
	eventRoom      domainroom.EventRoom
	hub            *realtime.Hub
	originPatterns []string
	logger         *slog.Logger
	metrics        *observability.Metrics
}

func NewRealtimeHandler(
	getRoom *roomapp.GetRoomUseCase,
	getRoundState *roomapp.GetRoundStateUseCase,
	hub *realtime.Hub,
	originPatterns []string,
	logger *slog.Logger,
	metrics *observability.Metrics,
) *RealtimeHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &RealtimeHandler{
		getRoom:        getRoom,
		getRoundState:  getRoundState,
		hub:            hub,
		originPatterns: originPatterns,
		logger:         logger,
		metrics:        metrics,
	}
}

type clientRealtimeEvent struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id"`
	Payload   json.RawMessage `json:"payload"`
}

type sendChatPayload struct {
	Text string `json:"text"`
}

type connectionSyncPayload struct {
	ConnectionMode string `json:"connection_mode"`
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
		if h.metrics != nil {
			h.metrics.RealtimeConnectionsTotal.WithLabelValues("room", "forbidden").Inc()
		}
		response.Error(w, apperror.Forbidden("Only room members can connect"))
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols:   []string{realtime.Subprotocol},
		OriginPatterns: h.originPatterns,
	})
	if err != nil {
		if h.metrics != nil {
			h.metrics.RealtimeConnectionsTotal.WithLabelValues("room", "upgrade_error").Inc()
		}
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
		if h.metrics != nil {
			h.metrics.RealtimeConnectionsTotal.WithLabelValues("room", "register_error").Inc()
		}
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
	if errors.Is(connectionErr, realtime.ErrConnectionReplaced) {
		_ = conn.Close(webSocketReplacedStatus, "Connection replaced by a newer session")
	} else {
		_ = conn.CloseNow()
	}
	select {
	case <-errCh:
	case <-time.After(webSocketCleanupWait):
		h.logger.WarnContext(r.Context(), "websocket loop cleanup timed out",
			slog.String("room_id", room.ID),
			slog.Int64("user_id", principal.UserID),
		)
	}

	status := websocket.CloseStatus(connectionErr)
	level := slog.LevelWarn
	if status == websocket.StatusNormalClosure ||
		status == websocket.StatusGoingAway ||
		status == webSocketReplacedStatus ||
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
	if h.metrics != nil {
		h.metrics.RealtimeDisconnects.WithLabelValues("room", realtimeDisconnectReason(connectionErr)).Inc()
	}
}

func (h *RealtimeHandler) ConnectUser(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{realtime.Subprotocol}, OriginPatterns: h.originPatterns,
	})
	if err != nil {
		if h.metrics != nil {
			h.metrics.RealtimeConnectionsTotal.WithLabelValues("user", "upgrade_error").Inc()
		}
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(webSocketReadLimit)
	client, err := h.hub.RegisterUser(principal.UserID, principal.UserName)
	if err != nil {
		if h.metrics != nil {
			h.metrics.RealtimeConnectionsTotal.WithLabelValues("user", "register_error").Inc()
		}
		_ = conn.Close(websocket.StatusTryAgainLater, "Realtime service unavailable")
		return
	}
	defer h.hub.Unregister(client)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 2)
	go func() { errCh <- h.userReadLoop(ctx, conn, client) }()
	go func() { errCh <- h.writeLoop(ctx, conn, client) }()
	connectionErr := <-errCh
	cancel()
	if errors.Is(connectionErr, realtime.ErrConnectionReplaced) {
		_ = conn.Close(webSocketReplacedStatus, "Connection replaced by a newer session")
	} else {
		_ = conn.CloseNow()
	}
	select {
	case <-errCh:
	case <-time.After(webSocketCleanupWait):
	}
	if h.metrics != nil {
		h.metrics.RealtimeDisconnects.WithLabelValues("user", realtimeDisconnectReason(connectionErr)).Inc()
	}
}

func (h *RealtimeHandler) userReadLoop(ctx context.Context, conn *websocket.Conn, client *realtime.Client) error {
	for {
		var event clientRealtimeEvent
		if err := wsjson.Read(ctx, conn, &event); err != nil {
			return err
		}
		if event.Type != "notification.sync" {
			h.hub.SendRequestError(client, event.RequestID, "UNSUPPORTED_EVENT", "Unsupported user realtime event")
			continue
		}
		var syncPayload connectionSyncPayload
		_ = json.Unmarshal(event.Payload, &syncPayload)
		if syncPayload.ConnectionMode == "reconnect" && h.metrics != nil {
			h.metrics.RealtimeReconnects.WithLabelValues("user").Inc()
		}
		if !h.claimRequest(client, event.RequestID) {
			continue
		}
		h.hub.Send(client, realtime.EventSocialUpdated, event.RequestID, map[string]string{"reason": "sync"})
	}
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
				h.hub.SendRequestError(client, event.RequestID, "INVALID_PAYLOAD", "Chat payload is invalid")
				continue
			}
			if !h.claimRequest(client, event.RequestID) {
				continue
			}
			state, stateErr := h.getRoundState.Execute(ctx, roomapp.GetRoundStateInput{
				RoomID: client.RoomID(), UserID: client.UserID(),
			})
			if stateErr == nil && state.Phase != domainroom.RoundPhaseGameFinished &&
				(state.Phase != domainroom.RoundPhaseDescribing || state.CurrentTurnPlayerID == nil || *state.CurrentTurnPlayerID != client.UserID()) {
				h.hub.SendRequestError(client, event.RequestID, "NOT_CURRENT_TURN", "Chỉ người đang đến lượt mới có thể mô tả")
				continue
			}
			if stateErr != nil && !errors.Is(stateErr, domainroom.ErrRoundCardNotFound) {
				return stateErr
			}
			if err := h.hub.PublishChatForRequest(client, event.RequestID, payload.Text); err != nil {
				switch {
				case errors.Is(err, realtime.ErrEmptyMessage):
					h.hub.SendRequestError(client, event.RequestID, "EMPTY_MESSAGE", "Chat message is required")
				case errors.Is(err, realtime.ErrMessageTooLong):
					h.hub.SendRequestError(client, event.RequestID, "MESSAGE_TOO_LONG", "Chat message must be at most 500 characters")
				case errors.Is(err, realtime.ErrRateLimited):
					h.hub.SendRequestError(client, event.RequestID, "RATE_LIMITED", "Send at most 10 messages every 5 seconds")
				case errors.Is(err, realtime.ErrSpectator):
					h.hub.SendRequestError(client, event.RequestID, "SPECTATOR_READ_ONLY", "Eliminated players can watch but cannot chat")
				case errors.Is(err, realtime.ErrChatUnavailable):
					h.hub.SendRequestError(client, event.RequestID, "CHAT_UNAVAILABLE", "Chat is temporarily unavailable")
				default:
					return err
				}
				if h.metrics != nil {
					h.metrics.RealtimeCommands.WithLabelValues(event.Type, "error").Inc()
				}
			} else if h.metrics != nil {
				h.metrics.RealtimeCommands.WithLabelValues(event.Type, "success").Inc()
			}
		case "state.sync":
			var syncPayload connectionSyncPayload
			_ = json.Unmarshal(event.Payload, &syncPayload)
			if syncPayload.ConnectionMode == "reconnect" && h.metrics != nil {
				h.metrics.RealtimeReconnects.WithLabelValues("room").Inc()
			}
			if !h.claimRequest(client, event.RequestID) {
				continue
			}
			state, err := h.getRoundState.Execute(ctx, roomapp.GetRoundStateInput{
				RoomID: client.RoomID(), UserID: client.UserID(),
			})
			if errors.Is(err, domainroom.ErrRoundCardNotFound) {
				h.hub.Send(client, realtime.EventStateUpdated, event.RequestID, map[string]string{"status": "waiting"})
				continue
			}
			if err != nil {
				h.sendGameError(client, event.RequestID, err)
				continue
			}
			h.sendState(client, state, event.RequestID)
			if h.metrics != nil {
				h.metrics.RealtimeCommands.WithLabelValues(event.Type, "success").Inc()
			}
		default:
			h.hub.SendRequestError(client, event.RequestID, "UNSUPPORTED_EVENT", "Unsupported realtime event")
			if h.metrics != nil {
				h.metrics.RealtimeCommands.WithLabelValues("unsupported", "error").Inc()
			}
		}
	}
}

func (h *RealtimeHandler) sendState(client *realtime.Client, state *domainroom.RoundState, requestID string) {
	h.hub.Send(client, realtime.EventStateUpdated, requestID, realtime.StateUpdated{
		Round:             state.RoundNumber,
		Cycle:             state.CycleNumber,
		Phase:             state.Phase,
		CurrentTurnUserID: state.CurrentTurnPlayerID,
		PhaseDeadlineAt:   state.PhaseDeadlineAt,
	})
}

func (h *RealtimeHandler) claimRequest(client *realtime.Client, requestID string) bool {
	err := h.hub.ClaimRequest(client, requestID)
	switch {
	case err == nil:
		return true
	case errors.Is(err, realtime.ErrRequestIDRequired):
		h.hub.SendRequestError(client, requestID, "REQUEST_ID_REQUIRED", "A request_id is required for socket commands")
	case errors.Is(err, realtime.ErrDuplicateEvent):
		h.hub.SendRequestError(client, requestID, "DUPLICATE_EVENT", "This socket command was already processed")
	case errors.Is(err, realtime.ErrConnectionReplaced), errors.Is(err, realtime.ErrClientNotFound):
		return false
	default:
		h.hub.SendRequestError(client, requestID, "SOCKET_DEDUPE_UNAVAILABLE", "Could not verify socket command uniqueness")
	}
	return false
}

func (h *RealtimeHandler) sendGameError(client *realtime.Client, requestID string, err error) {
	mapped := apperror.FromError(mapRoomError(err))
	h.hub.SendRequestError(client, requestID, string(mapped.Code), mapped.Message)
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
			if err := h.hub.Heartbeat(client); err != nil {
				return err
			}
			pingCtx, cancel := context.WithTimeout(ctx, webSocketWriteWait)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return err
			}
		case <-client.Done():
			return client.CloseReason()
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
			Eliminated: member.Eliminated,
		})
	}
	return members
}

func realtimeDisconnectReason(err error) string {
	switch {
	case err == nil,
		websocket.CloseStatus(err) == websocket.StatusNormalClosure,
		websocket.CloseStatus(err) == websocket.StatusGoingAway:
		return "normal"
	case errors.Is(err, realtime.ErrConnectionReplaced):
		return "replaced"
	case errors.Is(err, realtime.ErrSlowConsumer):
		return "slow_consumer"
	case errors.Is(err, context.Canceled), errors.Is(err, realtime.ErrHubClosed):
		return "shutdown"
	default:
		return "transport_error"
	}
}
