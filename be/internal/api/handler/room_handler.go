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
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type RoomHandler struct {
	createRoom       *roomapp.CreateRoomUseCase
	listRooms        *roomapp.ListRoomsUseCase
	getRoom          *roomapp.GetRoomUseCase
	joinRoom         *roomapp.JoinRoomUseCase
	kickMember       *roomapp.KickMemberUseCase
	dealRound        *roomapp.DealRoundUseCase
	getCard          *roomapp.GetCurrentCardUseCase
	updateDiscussion *roomapp.UpdateDiscussionUseCase
	getRoundState    *roomapp.GetRoundStateUseCase
	castVote         *roomapp.CastVoteUseCase
	playerReady      *roomapp.PlayerReadyUseCase
	finishTurn       *roomapp.FinishTurnUseCase
	mrWhiteGuess     *roomapp.MrWhiteGuessUseCase
	realtime         *realtime.Hub
	logger           *slog.Logger
}

func NewRoomHandler(
	createRoom *roomapp.CreateRoomUseCase,
	listRooms *roomapp.ListRoomsUseCase,
	getRoom *roomapp.GetRoomUseCase,
	joinRoom *roomapp.JoinRoomUseCase,
	kickMember *roomapp.KickMemberUseCase,
	dealRound *roomapp.DealRoundUseCase,
	getCard *roomapp.GetCurrentCardUseCase,
	updateDiscussion *roomapp.UpdateDiscussionUseCase,
	getRoundState *roomapp.GetRoundStateUseCase,
	castVote *roomapp.CastVoteUseCase,
	playerReady *roomapp.PlayerReadyUseCase,
	finishTurn *roomapp.FinishTurnUseCase,
	mrWhiteGuess *roomapp.MrWhiteGuessUseCase,
	realtimeHub *realtime.Hub,
	logger *slog.Logger,
) *RoomHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &RoomHandler{
		createRoom:       createRoom,
		listRooms:        listRooms,
		getRoom:          getRoom,
		joinRoom:         joinRoom,
		kickMember:       kickMember,
		dealRound:        dealRound,
		getCard:          getCard,
		updateDiscussion: updateDiscussion,
		getRoundState:    getRoundState,
		castVote:         castVote,
		playerReady:      playerReady,
		finishTurn:       finishTurn,
		mrWhiteGuess:     mrWhiteGuess,
		realtime:         realtimeHub,
		logger:           logger,
	}
}

type createRoomRequest struct {
	Name           string                  `json:"name" validate:"required,max=80"`
	LanguageCode   domainroom.LanguageCode `json:"language_code" validate:"required,oneof=en vi"`
	MaxPlayers     int                     `json:"max_players" validate:"required,gte=4,lte=12"`
	MrWhiteEnabled *bool                   `json:"mr_white_enabled"`
}

type joinRoomRequest struct {
	InviteCode string `json:"invite_code" validate:"required,min=6,max=8"`
}

type updateDiscussionRequest struct {
	DiscussionSeconds int `json:"discussion_seconds" validate:"required,gte=10,lte=30"`
}

type mrWhiteGuessRequest struct {
	Guess string `json:"guess" validate:"required,max=80"`
}

type castVoteRequest struct {
	TargetPlayerID int64 `json:"target_player_id" validate:"required,gt=0"`
}

type roomMemberResponse struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	SeatNumber int       `json:"seat_number"`
	Host       bool      `json:"host"`
	Current    bool      `json:"current"`
	JoinedAt   time.Time `json:"joined_at"`
	Eliminated bool      `json:"eliminated"`
}

type roomResponse struct {
	ID                string                  `json:"id"`
	InviteCode        string                  `json:"invite_code"`
	Name              string                  `json:"name"`
	LanguageCode      domainroom.LanguageCode `json:"language_code"`
	Status            domainroom.Status       `json:"status"`
	MaxPlayers        int                     `json:"max_players"`
	CurrentRound      int                     `json:"current_round"`
	DiscussionSeconds int                     `json:"discussion_seconds"`
	MrWhiteEnabled    bool                    `json:"mr_white_enabled"`
	PlayerCount       int                     `json:"player_count"`
	Version           int64                   `json:"version"`
	ExpiresAt         time.Time               `json:"expires_at"`
	CreatedAt         time.Time               `json:"created_at"`
	Players           []roomMemberResponse    `json:"players"`
}

type roundCardResponse struct {
	RoomID          string                `json:"room_id"`
	RoundNumber     int                   `json:"round"`
	PlayerID        int64                 `json:"player_id"`
	Role            domainroom.CardRole   `json:"role"`
	Word            string                `json:"word"`
	DealtAt         time.Time             `json:"dealt_at"`
	Phase           domainroom.RoundPhase `json:"phase"`
	PhaseDeadlineAt time.Time             `json:"phase_deadline_at"`
}

type roundStateResponse struct {
	RoomID              string                  `json:"room_id"`
	RoundNumber         int                     `json:"round"`
	CycleNumber         int                     `json:"cycle"`
	Phase               domainroom.RoundPhase   `json:"phase"`
	PhaseDeadlineAt     *time.Time              `json:"phase_deadline_at"`
	ReadyPlayers        int                     `json:"ready_players"`
	EligiblePlayers     int                     `json:"eligible_players"`
	CurrentUserReady    bool                    `json:"current_user_ready"`
	CurrentTurnPlayerID *int64                  `json:"current_turn_player_id"`
	TurnNumber          int                     `json:"turn_number"`
	TotalTurns          int                     `json:"total_turns"`
	TurnEndsAt          *time.Time              `json:"turn_ends_at"`
	EligibleVoters      int                     `json:"eligible_voters"`
	VotesCast           int                     `json:"votes_cast"`
	CurrentUserVoteID   *int64                  `json:"current_user_vote_id"`
	UndercoverPlayerID  *int64                  `json:"undercover_player_id"`
	MrWhitePlayerID     *int64                  `json:"mr_white_player_id"`
	EliminatedPlayerID  *int64                  `json:"eliminated_player_id"`
	EliminatedRole      *domainroom.CardRole    `json:"eliminated_role"`
	Winner              *domainroom.WinningSide `json:"winner"`
	MrWhiteGuessCorrect *bool                   `json:"mr_white_guess_correct"`
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
	mrWhiteEnabled := true
	if body.MrWhiteEnabled != nil {
		mrWhiteEnabled = *body.MrWhiteEnabled
	}
	room, err := h.createRoom.Execute(r.Context(), roomapp.CreateRoomInput{
		HostUserID:     principal.UserID,
		Name:           body.Name,
		LanguageCode:   body.LanguageCode,
		MaxPlayers:     body.MaxPlayers,
		MrWhiteEnabled: mrWhiteEnabled,
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
	if !isRoomMember(room, principal.UserID) {
		response.Error(w, apperror.Forbidden("Only room members can view this room"))
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
	h.publishRoomUpdated(room, "member_joined")
	response.OK(w, mapRoomResponse(room, principal.UserID))
}

func (h *RoomHandler) KickMember(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}

	targetUserID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil || targetUserID <= 0 {
		response.Error(w, apperror.InvalidRequest("Invalid room member id"))
		return
	}
	room, err := h.kickMember.Execute(r.Context(), roomapp.KickMemberInput{
		RoomID:       chi.URLParam(r, "roomID"),
		HostUserID:   principal.UserID,
		TargetUserID: targetUserID,
	})
	if err != nil {
		h.writeRoomError(w, r, "kick_room_member", err)
		return
	}
	if h.realtime != nil {
		h.realtime.EvictRoomMember(room.ID, targetUserID)
	}
	h.syncRoomMembers(room)
	h.publishRoomUpdated(room, "member_kicked")
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
			Round:           card.RoundNumber,
			DealtAt:         card.DealtAt,
			Phase:           card.Phase,
			PhaseDeadlineAt: card.PhaseDeadlineAt,
		})
	}
	if refreshedRoom, refreshErr := h.getRoom.Execute(r.Context(), card.RoomID); refreshErr == nil {
		h.syncRoomMembers(refreshedRoom)
		h.publishRoomUpdated(refreshedRoom, "game_started")
	}
	response.Created(w, mapRoundCardResponse(card))
}

func (h *RoomHandler) UpdateDiscussion(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}
	body, err := request.DecodeJSON[updateDiscussionRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	room, err := h.updateDiscussion.Execute(r.Context(), roomapp.UpdateDiscussionInput{
		RoomID:            chi.URLParam(r, "roomID"),
		HostUserID:        principal.UserID,
		DiscussionSeconds: body.DiscussionSeconds,
	})
	if err != nil {
		h.writeRoomError(w, r, "update_discussion", err)
		return
	}
	h.publishRoomUpdated(room, "discussion_updated")
	response.OK(w, mapRoomResponse(room, principal.UserID))
}

func (h *RoomHandler) GetRoundState(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}
	state, err := h.getRoundState.Execute(r.Context(), roomapp.GetRoundStateInput{
		RoomID: chi.URLParam(r, "roomID"),
		UserID: principal.UserID,
	})
	if err != nil {
		h.writeRoomError(w, r, "get_round_state", err)
		return
	}
	response.OK(w, mapRoundStateResponse(state))
}

func (h *RoomHandler) PlayerReady(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}
	state, err := h.playerReady.Execute(r.Context(), roomapp.GetRoundStateInput{
		RoomID: chi.URLParam(r, "roomID"), UserID: principal.UserID,
	})
	if err != nil {
		h.writeRoomError(w, r, "player_ready", err)
		return
	}
	h.publishStateUpdated(state)
	response.OK(w, mapRoundStateResponse(state))
}

func (h *RoomHandler) FinishTurn(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}
	state, err := h.finishTurn.Execute(r.Context(), roomapp.GetRoundStateInput{
		RoomID: chi.URLParam(r, "roomID"), UserID: principal.UserID,
	})
	if err != nil {
		h.writeRoomError(w, r, "finish_turn", err)
		return
	}
	h.publishStateUpdated(state)
	response.OK(w, mapRoundStateResponse(state))
}

func (h *RoomHandler) MrWhiteGuess(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}
	body, err := request.DecodeJSON[mrWhiteGuessRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	state, err := h.mrWhiteGuess.Execute(r.Context(), roomapp.MrWhiteGuessInput{
		RoomID: chi.URLParam(r, "roomID"), UserID: principal.UserID, Guess: body.Guess,
	})
	if err != nil {
		h.writeRoomError(w, r, "mr_white_guess", err)
		return
	}
	h.syncEliminatedMembers(r, state.RoomID)
	h.publishStateUpdated(state)
	response.OK(w, mapRoundStateResponse(state))
}

func (h *RoomHandler) CastVote(w http.ResponseWriter, r *http.Request) {
	principal, ok := apimiddleware.PrincipalFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.Unauthorized("Authentication required"))
		return
	}
	body, err := request.DecodeJSON[castVoteRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	state, err := h.castVote.Execute(r.Context(), roomapp.CastVoteInput{
		RoomID:       chi.URLParam(r, "roomID"),
		VoterUserID:  principal.UserID,
		TargetUserID: body.TargetPlayerID,
	})
	if err != nil {
		h.writeRoomError(w, r, "cast_vote", err)
		return
	}
	if state.FinalizedNow {
		h.syncEliminatedMembers(r, state.RoomID)
	}
	if h.realtime != nil {
		h.realtime.Publish(state.RoomID, realtime.EventVoteUpdated, realtime.VoteUpdated{
			Round:          state.RoundNumber,
			VotesCast:      state.VotesCast,
			EligibleVoters: state.EligibleVoters,
			Completed:      state.Phase != domainroom.RoundPhaseVoting,
		})
	}
	h.publishStateUpdated(state)
	response.Created(w, mapRoundStateResponse(state))
}

func (h *RoomHandler) publishStateUpdated(state *domainroom.RoundState) {
	if h.realtime == nil {
		return
	}
	h.realtime.Publish(state.RoomID, realtime.EventStateUpdated, realtime.StateUpdated{
		Round:             state.RoundNumber,
		Cycle:             state.CycleNumber,
		Phase:             state.Phase,
		CurrentTurnUserID: state.CurrentTurnPlayerID,
		PhaseDeadlineAt:   state.PhaseDeadlineAt,
	})
}

func (h *RoomHandler) syncEliminatedMembers(r *http.Request, roomID string) {
	room, err := h.getRoom.Execute(r.Context(), roomID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "could not refresh eliminated room members",
			slog.String("room_id", roomID),
			slog.Any("error", err),
		)
		return
	}
	h.syncRoomMembers(room)
	h.publishRoomUpdated(room, "player_eliminated")
}

func (h *RoomHandler) publishRoomUpdated(room *domainroom.Room, reason string) {
	if h.realtime != nil {
		h.realtime.Publish(room.ID, realtime.EventRoomUpdated, realtime.RoomUpdated{
			Version: room.Version,
			Reason:  reason,
		})
	}
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
		return apperror.Forbidden("Only the room admin can perform this action")
	case errors.Is(err, domainroom.ErrRoomMemberNotFound):
		return apperror.NotFound("Room member not found")
	case errors.Is(err, domainroom.ErrCannotKickSelf):
		return apperror.InvalidRequest("The room admin cannot remove themselves")
	case errors.Is(err, domainroom.ErrNotEnoughPlayers):
		return apperror.Conflict("At least four players are required to start a game")
	case errors.Is(err, domainroom.ErrRoundCardNotFound):
		return apperror.NotFound("No card is available for this player")
	case errors.Is(err, domainroom.ErrWordPairNotFound):
		return apperror.Conflict("No active word pairs are configured")
	case errors.Is(err, domainroom.ErrRoundInProgress):
		return apperror.Conflict("Finish the current vote before starting the next round")
	case errors.Is(err, domainroom.ErrVotingNotOpen):
		return apperror.Conflict("Voting is not open in the current game phase")
	case errors.Is(err, domainroom.ErrVotingCompleted):
		return apperror.Conflict("Voting has already completed")
	case errors.Is(err, domainroom.ErrAlreadyVoted):
		return apperror.Conflict("You have already voted in this round")
	case errors.Is(err, domainroom.ErrInvalidVoteTarget):
		return apperror.InvalidRequest("Vote for another player in this room")
	case errors.Is(err, domainroom.ErrPlayerEliminated):
		return apperror.Forbidden("Eliminated players can only spectate")
	case errors.Is(err, domainroom.ErrNotCurrentTurn):
		return apperror.Forbidden("Only the current player can finish this turn")
	case errors.Is(err, domainroom.ErrMrWhiteGuessNotAllowed):
		return apperror.Forbidden("Only the eliminated Mr. White can guess during this phase")
	case errors.Is(err, domainroom.ErrMrWhiteAlreadyGuessed):
		return apperror.Conflict("Mr. White has already used the guess")
	case errors.Is(err, domainroom.ErrInvalidGameState):
		return apperror.Conflict("This action is not available in the current game phase")
	case errors.Is(err, domainroom.ErrInvalidRoomID),
		errors.Is(err, domainroom.ErrInvalidInviteCode),
		errors.Is(err, domainroom.ErrInvalidLanguageCode),
		errors.Is(err, domainroom.ErrInvalidMaxPlayers),
		errors.Is(err, domainroom.ErrRoomNameRequired),
		errors.Is(err, domainroom.ErrRoomNameTooLong),
		errors.Is(err, domainroom.ErrInvalidHostUserID),
		errors.Is(err, domainroom.ErrInvalidRoomMemberID),
		errors.Is(err, domainroom.ErrInvalidRoomExpiration),
		errors.Is(err, domainroom.ErrInvalidDiscussionTime):
		return apperror.InvalidRequest(err.Error())
	default:
		return apperror.Internal(err)
	}
}

func mapRoundCardResponse(card *domainroom.RoundCard) roundCardResponse {
	return roundCardResponse{
		RoomID:          card.RoomID,
		RoundNumber:     card.RoundNumber,
		PlayerID:        card.PlayerID,
		Role:            card.Role,
		Word:            card.Word,
		DealtAt:         card.DealtAt,
		Phase:           card.Phase,
		PhaseDeadlineAt: card.PhaseDeadlineAt,
	}
}

func mapRoundStateResponse(state *domainroom.RoundState) roundStateResponse {
	return roundStateResponse{
		RoomID:              state.RoomID,
		RoundNumber:         state.RoundNumber,
		CycleNumber:         state.CycleNumber,
		Phase:               state.Phase,
		PhaseDeadlineAt:     state.PhaseDeadlineAt,
		ReadyPlayers:        state.ReadyPlayers,
		EligiblePlayers:     state.EligiblePlayers,
		CurrentUserReady:    state.CurrentUserReady,
		CurrentTurnPlayerID: state.CurrentTurnPlayerID,
		TurnNumber:          state.TurnNumber,
		TotalTurns:          state.TotalTurns,
		TurnEndsAt:          state.TurnEndsAt,
		EligibleVoters:      state.EligibleVoters,
		VotesCast:           state.VotesCast,
		CurrentUserVoteID:   state.CurrentUserVoteID,
		UndercoverPlayerID:  state.UndercoverPlayerID,
		MrWhitePlayerID:     state.MrWhitePlayerID,
		EliminatedPlayerID:  state.EliminatedPlayerID,
		EliminatedRole:      state.EliminatedRole,
		Winner:              state.Winner,
		MrWhiteGuessCorrect: state.MrWhiteGuessCorrect,
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
			Eliminated: member.Eliminated,
		})
	}
	return roomResponse{
		ID:                room.ID,
		InviteCode:        room.InviteCode,
		Name:              room.Name,
		LanguageCode:      room.LanguageCode,
		Status:            room.Status,
		MaxPlayers:        room.MaxPlayers,
		CurrentRound:      room.CurrentRound,
		DiscussionSeconds: room.DiscussionSeconds,
		MrWhiteEnabled:    room.MrWhiteEnabled,
		PlayerCount:       len(players),
		Version:           room.Version,
		ExpiresAt:         room.ExpiresAt,
		CreatedAt:         room.CreatedAt,
		Players:           players,
	}
}
