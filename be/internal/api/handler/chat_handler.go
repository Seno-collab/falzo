package handler

import (
	"be/internal/api/http/response"
	apimiddleware "be/internal/api/middleware"
	chatapp "be/internal/application/chat"
	roomapp "be/internal/application/room"
	"be/internal/shared/apperror"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type ChatHandler struct {
	chat    *chatapp.Service
	getRoom *roomapp.GetRoomUseCase
	logger  *slog.Logger
}

func NewChatHandler(chat *chatapp.Service, getRoom *roomapp.GetRoomUseCase, logger *slog.Logger) *ChatHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &ChatHandler{chat: chat, getRoom: getRoom, logger: logger}
}

func (h *ChatHandler) ListRoomMessages(w http.ResponseWriter, r *http.Request) {
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
		response.Error(w, apperror.Forbidden("Only room members can read chat history"))
		return
	}
	limit := chatapp.DefaultHistoryLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed <= 0 {
			response.Error(w, apperror.InvalidRequest("limit must be a positive integer"))
			return
		}
		limit = parsed
	}
	var before *time.Time
	if raw := r.URL.Query().Get("before"); raw != "" {
		parsed, parseErr := time.Parse(time.RFC3339Nano, raw)
		if parseErr != nil {
			response.Error(w, apperror.InvalidRequest("before must be an RFC3339 timestamp"))
			return
		}
		before = &parsed
	}
	messages, err := h.chat.ListRoomMessages(r.Context(), room.ID, before, limit)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list room chat messages", slog.Any("error", err))
		response.Error(w, apperror.Internal(err))
		return
	}
	response.OK(w, messages)
}
