package handler

import (
	"be/internal/api/http/request"
	"be/internal/api/http/response"
	apimiddleware "be/internal/api/middleware"
	roomapp "be/internal/application/room"
	domainroom "be/internal/domain/room"
	"be/internal/realtime"
	"be/internal/shared/apperror"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type RoomHandler struct {
	createRoom *roomapp.CreateRoomUseCase
	listRooms  *roomapp.ListRoomsUseCase
	getRoom    *roomapp.GetRoomUseCase
	joinRoom   *roomapp.JoinRoomUseCase
	dealRound  *roomapp.DealRoundUseCase
	getCard    *roomapp.GetCurrentCardUseCase
	realtime   *realtime.Hub
	logger     *slog.Logger
}

func NewRoomHandler(
	createRoom *roomapp.CreateRoomUseCase,
	listRooms *roomapp.ListRoomsUseCase,
	getRoom *roomapp.GetRoomUseCase,
	joinRoom *roomapp.JoinRoomUseCase,
	dealRound *roomapp.DealRoundUseCase,
	getCard *roomapp.GetCurrentCardUseCase,
	realtimeHub *realtime.Hub,
	logger *slog.Logger,
) *RoomHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &RoomHandler{
		createRoom: createRoom,
		listRooms:  listRooms,
		getRoom:    getRoom,
		joinRoom:   joinRoom,
		dealRound:  dealRound,
		getCard:    getCard,
		realtime:   realtimeHub,
		logger:     logger,
	}
}

type createRoomRequest struct {
	Name         string                  `json:"name" validate:"required,max=80"`
	LanguageCode domainroom.LanguageCode `json:"language_code" validate:"required,oneof=en vi"`
	MaxPlayers   int                     `json:"max_players" validate:"required,gte=4,lte=12"`
}

type joinRoomRequest struct {
	InviteCode string `json:"invite_code" validate:"required,min=6,max=8"`
}

type roomMemberResponse struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	SeatNumber int       `json:"seat_number"`
	Host       bool      `json:"host"`
	Current    bool      `json:"current"`
	JoinedAt   time.Time `json:"joined_at"`
}

type roomResponse struct {
	ID           string                  `json:"id"`
	InviteCode   string                  `json:"invite_code"`
	Name         string                  `json:"name"`
	LanguageCode domainroom.LanguageCode `json:"language_code"`
	Status       domainroom.Status       `json:"status"`
	MaxPlayers   int                     `json:"max_players"`
	CurrentRound int                     `json:"current_round"`
	PlayerCount  int                     `json:"player_count"`
	Version      int64                   `json:"version"`
	ExpiresAt    time.Time               `json:"expires_at"`
	CreatedAt    time.Time               `json:"created_at"`
	Players      []roomMemberResponse    `json:"players"`
}

type roundCardResponse struct {
	RoomID      string              `json:"room_id"`
	RoundNumber int                 `json:"round"`
	PlayerID    int64               `json:"player_id"`
	Role        domainroom.CardRole `json:"role"`
	Word        string              `json:"word"`
	DealtAt     time.Time           `json:"dealt_at"`
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}

	body, err := request.DecodeJSON[createRoomRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	room, err := h.createRoom.Execute(r.Context(), roomapp.CreateRoomInput{
		HostUserID:   principal.UserID,
		Name:         body.Name,
		LanguageCode: body.LanguageCode,
		MaxPlayers:   body.MaxPlayers,
	})
	if err != nil {
		h.writeRoomError(w, r, "create_room", err)
		return
	}
	h.syncRoomMembers(room)
	response.Created(w, mapRoomResponse(room, principal.UserID))
}

func (h *RoomHandler) List(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}

	rooms, err := h.listRooms.Execute(r.Context())
	if err != nil {
		h.writeRoomError(w, r, "list_rooms", err)
		return
	}
	data := make([]roomResponse, 0, len(rooms))
	for _, room := range rooms {
		data = append(data, mapRoomResponse(room, principal.UserID))
	}
	response.OK(w, data)
}

func (h *RoomHandler) Get(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}

	room, err := h.getRoom.Execute(r.Context(), chi.URLParam(r, "roomID"))
	if err != nil {
		h.writeRoomError(w, r, "get_room", err)
		return
	}
	response.OK(w, mapRoomResponse(room, principal.UserID))
}

func (h *RoomHandler) Join(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}

	body, err := request.DecodeJSON[joinRoomRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	room, err := h.joinRoom.Execute(r.Context(), roomapp.JoinRoomInput{
		UserID:     principal.UserID,
		InviteCode: body.InviteCode,
	})
	if err != nil {
		h.writeRoomError(w, r, "join_room", err)
		return
	}
	h.syncRoomMembers(room)
	response.OK(w, mapRoomResponse(room, principal.UserID))
}

func (h *RoomHandler) DealRound(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}

	card, err := h.dealRound.Execute(r.Context(), roomapp.DealRoundInput{
		RoomID:     chi.URLParam(r, "roomID"),
		HostUserID: principal.UserID,
	})
	if err != nil {
		h.writeRoomError(w, r, "deal_round", err)
		return
	}
	if h.realtime != nil {
		h.realtime.Publish(card.RoomID, realtime.EventRoundStarted, realtime.RoundStarted{
			Round:   card.RoundNumber,
			DealtAt: card.DealtAt,
		})
	}
	response.Created(w, mapRoundCardResponse(card))
}

func (h *RoomHandler) syncRoomMembers(room *domainroom.Room) {
	if h.realtime != nil {
		h.realtime.UpdateMembers(room.ID, mapRealtimeMembers(room))
	}
}

func (h *RoomHandler) GetCurrentCard(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}

	card, err := h.getCard.Execute(r.Context(), roomapp.GetCurrentCardInput{
		RoomID: chi.URLParam(r, "roomID"),
		UserID: principal.UserID,
	})
	if err != nil {
		h.writeRoomError(w, r, "get_current_card", err)
		return
	}
	response.OK(w, mapRoundCardResponse(card))
}

func (h *RoomHandler) writeRoomError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	mappedErr := mapRoomError(err)
	appErr := apperror.FromError(mappedErr)
	level := slog.LevelWarn
	if appErr.Code == apperror.CodeInternalServerError {
		level = slog.LevelError
	}
	h.logger.LogAttrs(r.Context(), level, "room operation failed",
		slog.String("operation", operation),
		slog.String("code", string(appErr.Code)),
		slog.Any("error", err),
	)
	response.Error(w, mappedErr)
}

func mapRoomError(err error) error {
	switch {
	case errors.Is(err, domainroom.ErrRoomNotFound):
		return apperror.NotFound("Room not found")
	case errors.Is(err, domainroom.ErrRoomFull):
		return apperror.GameRoomFull()
	case errors.Is(err, domainroom.ErrRoomNotWaiting):
		return apperror.Conflict("Room is not waiting for players")
	case errors.Is(err, domainroom.ErrNotRoomHost):
		return apperror.Forbidden("Only the room admin can deal a round")
	case errors.Is(err, domainroom.ErrNotEnoughPlayers):
		return apperror.Conflict("At least two players are required to deal a round")
	case errors.Is(err, domainroom.ErrRoundCardNotFound):
		return apperror.NotFound("No card is available for this player")
	case errors.Is(err, domainroom.ErrWordPairNotFound):
		return apperror.Conflict("No active word pairs are configured")
	case errors.Is(err, domainroom.ErrInvalidRoomID),
		errors.Is(err, domainroom.ErrInvalidInviteCode),
		errors.Is(err, domainroom.ErrInvalidLanguageCode),
		errors.Is(err, domainroom.ErrInvalidMaxPlayers),
		errors.Is(err, domainroom.ErrRoomNameRequired),
		errors.Is(err, domainroom.ErrRoomNameTooLong),
		errors.Is(err, domainroom.ErrInvalidHostUserID),
		errors.Is(err, domainroom.ErrInvalidRoomExpiration):
		return apperror.InvalidRequest(err.Error())
	default:
		return apperror.Internal(err)
	}
}

func mapRoundCardResponse(card *domainroom.RoundCard) roundCardResponse {
	return roundCardResponse{
		RoomID:      card.RoomID,
		RoundNumber: card.RoundNumber,
		PlayerID:    card.PlayerID,
		Role:        card.Role,
		Word:        card.Word,
		DealtAt:     card.DealtAt,
	}
}

func mapRoomResponse(room *domainroom.Room, currentUserID int64) roomResponse {
	players := make([]roomMemberResponse, 0, len(room.Members))
	for _, member := range room.Members {
		players = append(players, roomMemberResponse{
			ID:         member.UserID,
			Name:       member.UserName,
			SeatNumber: member.SeatNumber,
			Host:       member.IsHost,
			Current:    member.UserID == currentUserID,
			JoinedAt:   member.JoinedAt,
		})
	}
	return roomResponse{
		ID:           room.ID,
		InviteCode:   room.InviteCode,
		Name:         room.Name,
		LanguageCode: room.LanguageCode,
		Status:       room.Status,
		MaxPlayers:   room.MaxPlayers,
		CurrentRound: room.CurrentRound,
		PlayerCount:  len(players),
		Version:      room.Version,
		ExpiresAt:    room.ExpiresAt,
		CreatedAt:    room.CreatedAt,
		Players:      players,
	}
}
