package postgres

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	gameengine "be/internal/game"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	db *pgxpool.Pool
}

const roomColumns = `id::text, invite_code, name, language_code, host_user_id, status, max_players, current_round, discussion_seconds, mr_white_enabled, version, expires_at, created_at, updated_at`

func NewRoomRepository(db *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) CreateWithHost(ctx context.Context, room *domainroom.Room) (*domainroom.Room, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	created := &domainroom.Room{}
	err = tx.QueryRow(ctx, `
			INSERT INTO rooms (
				id, invite_code, name, language_code, host_user_id, status,
				max_players, current_round, discussion_seconds, mr_white_enabled, version, expires_at, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING `+roomColumns,
		room.ID,
		room.InviteCode,
		room.Name,
		room.LanguageCode,
		room.HostUserID,
		room.Status,
		room.MaxPlayers,
		room.CurrentRound,
		room.DiscussionSeconds,
		room.MrWhiteEnabled,
		room.Version,
		room.ExpiresAt,
		room.CreatedAt,
		room.UpdatedAt,
	).Scan(roomScanTargets(created)...)
	if err != nil {
		if IsUniqueViolation(err) {
			return nil, domainroom.ErrInviteCodeConflict
		}
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO room_members (room_id, user_id, seat_number, joined_at)
		VALUES ($1, $2, 1, $3)`,
		created.ID,
		created.HostUserID,
		created.CreatedAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, created.ID)
}

func (r *RoomRepository) ListActive(ctx context.Context) ([]*domainroom.Room, error) {
	rows, err := r.db.Query(ctx, `
		SELECT `+roomColumns+`
		FROM rooms
		WHERE expires_at > now() AND status <> $1
		ORDER BY created_at DESC
		LIMIT 50`, domainroom.StatusClosed)
	if err != nil {
		return nil, err
	}

	rooms := make([]*domainroom.Room, 0)
	for rows.Next() {
		room := &domainroom.Room{}
		if err := rows.Scan(roomScanTargets(room)...); err != nil {
			rows.Close()
			return nil, err
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for _, room := range rooms {
		members, err := r.loadMembers(ctx, r.db, room.ID, room.HostUserID)
		if err != nil {
			return nil, err
		}
		room.Members = members
	}
	return rooms, nil
}

func (r *RoomRepository) FindByID(ctx context.Context, roomID string) (*domainroom.Room, error) {
	room := &domainroom.Room{}
	err := r.db.QueryRow(ctx, `
		SELECT `+roomColumns+`
		FROM rooms
		WHERE id = $1 AND expires_at > now() AND status <> $2`,
		roomID,
		domainroom.StatusClosed,
	).Scan(roomScanTargets(room)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}

	members, err := r.loadMembers(ctx, r.db, room.ID, room.HostUserID)
	if err != nil {
		return nil, err
	}
	room.Members = members
	return room, nil
}

func (r *RoomRepository) JoinByInviteCode(ctx context.Context, inviteCode string, userID int64) (*domainroom.Room, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var roomID string
	var status domainroom.Status
	var maxPlayers int
	err = tx.QueryRow(ctx, `
		SELECT id::text, status, max_players
		FROM rooms
		WHERE invite_code = $1 AND expires_at > now() AND status <> $2
		FOR UPDATE`,
		inviteCode,
		domainroom.StatusClosed,
	).Scan(&roomID, &status, &maxPlayers)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}

	var alreadyJoined bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2
		)`, roomID, userID).Scan(&alreadyJoined); err != nil {
		return nil, err
	}
	if alreadyJoined {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return r.FindByID(ctx, roomID)
	}

	if status != domainroom.StatusWaiting {
		return nil, domainroom.ErrRoomNotWaiting
	}

	var memberCount int
	var nextSeat int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(MAX(seat_number), 0) + 1
		FROM room_members
		WHERE room_id = $1`, roomID).Scan(&memberCount, &nextSeat); err != nil {
		return nil, err
	}
	if memberCount >= maxPlayers {
		return nil, domainroom.ErrRoomFull
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO room_members (room_id, user_id, seat_number)
		VALUES ($1, $2, $3)`, roomID, userID, nextSeat); err != nil {
		if IsUniqueViolation(err) {
			return nil, fmt.Errorf("join room conflict: %w", err)
		}
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE rooms
		SET version = version + 1, updated_at = now()
		WHERE id = $1`, roomID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, roomID)
}

func (r *RoomRepository) FindRandomWordPair(
	ctx context.Context,
	languageCode domainroom.LanguageCode,
) (*domainroom.WordPair, error) {
	pair := &domainroom.WordPair{}
	err := r.db.QueryRow(ctx, `
		SELECT id, common_word, different_word, category, language_code
		FROM undercover_word_pairs
		WHERE is_active = TRUE AND language_code = $1
		ORDER BY random()
		LIMIT 1`, languageCode,
	).Scan(
		&pair.ID,
		&pair.CommonWord,
		&pair.DifferentWord,
		&pair.Category,
		&pair.LanguageCode,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrWordPairNotFound
	}
	if err != nil {
		return nil, err
	}
	return pair, nil
}

func (r *RoomRepository) StartRound(ctx context.Context, input roomports.StartRoundInput) (*domainroom.RoundCard, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var hostUserID int64
	var currentRound int
	err = tx.QueryRow(ctx, `
		SELECT host_user_id, current_round
		FROM rooms
		WHERE id = $1 AND expires_at > now() AND status <> $2
		FOR UPDATE`,
		input.RoomID,
		domainroom.StatusClosed,
	).Scan(&hostUserID, &currentRound)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}
	if hostUserID != input.HostUserID {
		return nil, domainroom.ErrNotRoomHost
	}
	if currentRound > 0 {
		var phase domainroom.RoundPhase
		if err := tx.QueryRow(ctx, `
			SELECT phase
			FROM room_rounds
			WHERE room_id = $1 AND round_number = $2`, input.RoomID, currentRound,
		).Scan(&phase); err != nil {
			return nil, err
		}
		if phase != domainroom.RoundPhaseGameFinished {
			return nil, domainroom.ErrRoundInProgress
		}
	}

	var memberCount int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM room_members WHERE room_id = $1`, input.RoomID,
	).Scan(&memberCount); err != nil {
		return nil, err
	}
	if memberCount < domainroom.MinPlayersPerRoom || len(input.TurnOrder) != memberCount {
		return nil, domainroom.ErrNotEnoughPlayers
	}
	uniquePlayers := make(map[int64]struct{}, len(input.TurnOrder))
	for _, playerID := range input.TurnOrder {
		uniquePlayers[playerID] = struct{}{}
	}
	if len(uniquePlayers) != memberCount {
		return nil, domainroom.ErrInvalidGameState
	}
	if _, ok := uniquePlayers[input.UndercoverPlayerID]; !ok {
		return nil, domainroom.ErrInvalidGameState
	}
	if input.MrWhitePlayerID != nil {
		if _, ok := uniquePlayers[*input.MrWhitePlayerID]; !ok || *input.MrWhitePlayerID == input.UndercoverPlayerID {
			return nil, domainroom.ErrInvalidGameState
		}
	}
	var matchedMembers int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM room_members
		WHERE room_id = $1 AND user_id = ANY($2)`, input.RoomID, input.TurnOrder,
	).Scan(&matchedMembers); err != nil {
		return nil, err
	}
	if matchedMembers != memberCount {
		return nil, domainroom.ErrInvalidGameState
	}

	nextRound := currentRound + 1
	if _, err := tx.Exec(ctx, `
		UPDATE room_members SET eliminated_at = NULL WHERE room_id = $1`, input.RoomID,
	); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO room_rounds (
			room_id, round_number, word_pair_id, common_word, different_word,
			undercover_user_id, mr_white_user_id, dealt_at, phase, phase_deadline_at, cycle_number
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 1)`,
		input.RoomID,
		nextRound,
		input.WordPairID,
		input.CommonWord,
		input.DifferentWord,
		input.UndercoverPlayerID,
		input.MrWhitePlayerID,
		input.DealtAt,
		domainroom.RoundPhaseRevealingRole,
		input.RoleRevealEndsAt,
	); err != nil {
		return nil, err
	}
	for index, playerID := range input.TurnOrder {
		if _, err := tx.Exec(ctx, `
			INSERT INTO room_round_turns (room_id, round_number, cycle_number, turn_number, user_id)
			VALUES ($1, $2, 1, $3, $4)`, input.RoomID, nextRound, index+1, playerID,
		); err != nil {
			return nil, err
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE rooms
		SET status = $2,
			current_round = $3,
			version = version + 1,
			updated_at = $4
		WHERE id = $1`,
		input.RoomID,
		domainroom.StatusPlaying,
		nextRound,
		input.DealtAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	card := &domainroom.RoundCard{
		RoomID:          input.RoomID,
		RoundNumber:     nextRound,
		PlayerID:        input.HostUserID,
		Role:            domainroom.CardRoleCivilian,
		Word:            input.CommonWord,
		DealtAt:         input.DealtAt,
		Phase:           domainroom.RoundPhaseRevealingRole,
		PhaseDeadlineAt: input.RoleRevealEndsAt,
	}
	if input.HostUserID == input.UndercoverPlayerID {
		card.Role = domainroom.CardRoleUndercover
		card.Word = input.DifferentWord
	} else if input.MrWhitePlayerID != nil && input.HostUserID == *input.MrWhitePlayerID {
		card.Role = domainroom.CardRoleMrWhite
		card.Word = ""
	}
	return card, nil
}

func (r *RoomRepository) UpdateDiscussionSeconds(
	ctx context.Context,
	input roomports.UpdateDiscussionInput,
) (*domainroom.Room, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var hostUserID int64
	var status domainroom.Status
	err = tx.QueryRow(ctx, `
		SELECT host_user_id, status
		FROM rooms
		WHERE id = $1 AND expires_at > now() AND status <> $2
		FOR UPDATE`, input.RoomID, domainroom.StatusClosed,
	).Scan(&hostUserID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrRoomNotFound
	}
	if err != nil {
		return nil, err
	}
	if hostUserID != input.HostUserID {
		return nil, domainroom.ErrNotRoomHost
	}
	if status != domainroom.StatusWaiting {
		return nil, domainroom.ErrRoomNotWaiting
	}

	if _, err := tx.Exec(ctx, `
		UPDATE rooms
		SET discussion_seconds = $2, version = version + 1, updated_at = now()
		WHERE id = $1`, input.RoomID, input.DiscussionSeconds,
	); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindByID(ctx, input.RoomID)
}

func (r *RoomRepository) FindCurrentRoundState(
	ctx context.Context,
	roomID string,
	userID int64,
	now time.Time,
) (*domainroom.RoundState, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	changed, err := advanceGameState(ctx, tx, roomID, now)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	state, err := loadCurrentRoundState(ctx, r.db, roomID, userID, now)
	if state != nil {
		state.FinalizedNow = changed
	}
	return state, err
}

func (r *RoomRepository) MarkPlayerReady(ctx context.Context, input roomports.PlayerActionInput) (*domainroom.RoundState, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := advanceGameState(ctx, tx, input.RoomID, input.At); err != nil {
		return nil, err
	}
	var roundNumber int
	var phase domainroom.RoundPhase
	err = tx.QueryRow(ctx, `
		SELECT r.current_round, rr.phase
		FROM rooms r
		JOIN room_members m ON m.room_id = r.id AND m.user_id = $2 AND m.eliminated_at IS NULL
		JOIN room_rounds rr ON rr.room_id = r.id AND rr.round_number = r.current_round
		WHERE r.id = $1 AND r.status = $3
		FOR UPDATE OF rr`, input.RoomID, input.UserID, domainroom.StatusPlaying,
	).Scan(&roundNumber, &phase)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrRoundCardNotFound
	}
	if err != nil {
		return nil, err
	}
	if phase == domainroom.RoundPhaseRevealingRole {
		if _, err := tx.Exec(ctx, `
			INSERT INTO room_round_ready_players (room_id, round_number, user_id, ready_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING`, input.RoomID, roundNumber, input.UserID, input.At,
		); err != nil {
			return nil, err
		}
		var readyCount, playerCount int
		if err := tx.QueryRow(ctx, `
			SELECT
				(SELECT COUNT(*) FROM room_round_ready_players WHERE room_id = $1 AND round_number = $2),
				(SELECT COUNT(*) FROM room_members WHERE room_id = $1)`, input.RoomID, roundNumber,
		).Scan(&readyCount, &playerCount); err != nil {
			return nil, err
		}
		if readyCount >= playerCount {
			if err := startDescribing(ctx, tx, input.RoomID, roundNumber, 1, input.At); err != nil {
				return nil, err
			}
		}
	} else if phase != domainroom.RoundPhaseDescribing {
		return nil, domainroom.ErrInvalidGameState
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindCurrentRoundState(ctx, input.RoomID, input.UserID, input.At)
}

func (r *RoomRepository) FinishTurn(ctx context.Context, input roomports.PlayerActionInput) (*domainroom.RoundState, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := advanceGameState(ctx, tx, input.RoomID, input.At); err != nil {
		return nil, err
	}
	var roundNumber, cycleNumber int
	var phase domainroom.RoundPhase
	var currentPlayerID pgtype.Int8
	err = tx.QueryRow(ctx, `
		SELECT r.current_round, rr.cycle_number, rr.phase,
			(SELECT user_id FROM room_round_turns
			 WHERE room_id = r.id AND round_number = r.current_round
			   AND cycle_number = rr.cycle_number AND started_at IS NOT NULL AND finished_at IS NULL
			 ORDER BY turn_number LIMIT 1)
		FROM rooms r
		JOIN room_rounds rr ON rr.room_id = r.id AND rr.round_number = r.current_round
		WHERE r.id = $1
		FOR UPDATE OF rr`, input.RoomID,
	).Scan(&roundNumber, &cycleNumber, &phase, &currentPlayerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrRoundCardNotFound
	}
	if err != nil {
		return nil, err
	}
	if phase != domainroom.RoundPhaseDescribing {
		return nil, domainroom.ErrInvalidGameState
	}
	if !currentPlayerID.Valid || currentPlayerID.Int64 != input.UserID {
		return nil, domainroom.ErrNotCurrentTurn
	}
	if err := finishCurrentTurn(ctx, tx, input.RoomID, roundNumber, cycleNumber, input.At, false); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindCurrentRoundState(ctx, input.RoomID, input.UserID, input.At)
}

func (r *RoomRepository) CastVote(ctx context.Context, input roomports.CastVoteInput) (*domainroom.RoundState, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := advanceGameState(ctx, tx, input.RoomID, input.VotedAt); err != nil {
		return nil, err
	}
	var roundNumber, cycleNumber int
	var phase domainroom.RoundPhase
	var votingEndsAt pgtype.Timestamptz
	err = tx.QueryRow(ctx, `
		SELECT r.current_round, rr.cycle_number, rr.phase, rr.phase_deadline_at
		FROM rooms r
		JOIN room_rounds rr
			ON rr.room_id = r.id AND rr.round_number = r.current_round
		WHERE r.id = $1 AND r.expires_at > now() AND r.status = $2
		FOR UPDATE OF r, rr`, input.RoomID, domainroom.StatusPlaying,
	).Scan(&roundNumber, &cycleNumber, &phase, &votingEndsAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrRoundCardNotFound
	}
	if err != nil {
		return nil, err
	}
	if phase != domainroom.RoundPhaseVoting {
		return nil, domainroom.ErrVotingNotOpen
	}
	if !votingEndsAt.Valid || !input.VotedAt.Before(votingEndsAt.Time) {
		if err := completeVoting(ctx, tx, input.RoomID, roundNumber, cycleNumber, input.VotedAt); err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		state, err := r.FindCurrentRoundState(ctx, input.RoomID, input.VoterUserID, input.VotedAt)
		if state != nil {
			state.FinalizedNow = true
		}
		return state, err
	}

	var voterIsMember bool
	var voterIsActive bool
	var targetIsActive bool
	if err := tx.QueryRow(ctx, `
		SELECT
			EXISTS (SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2),
			EXISTS (SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2 AND eliminated_at IS NULL),
			EXISTS (SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $3 AND eliminated_at IS NULL)`,
		input.RoomID, input.VoterUserID, input.TargetUserID,
	).Scan(&voterIsMember, &voterIsActive, &targetIsActive); err != nil {
		return nil, err
	}
	if voterIsMember && !voterIsActive {
		return nil, domainroom.ErrPlayerEliminated
	}
	if !voterIsMember || !targetIsActive || input.VoterUserID == input.TargetUserID {
		return nil, domainroom.ErrInvalidVoteTarget
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO room_round_votes (
			room_id, round_number, cycle_number, voter_user_id, target_user_id, voted_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		input.RoomID, roundNumber, cycleNumber, input.VoterUserID, input.TargetUserID, input.VotedAt,
	); err != nil {
		if IsUniqueViolation(err) {
			return nil, domainroom.ErrAlreadyVoted
		}
		return nil, err
	}

	var eligibleVoters int
	var votesCast int
	if err := tx.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM room_members WHERE room_id = $1 AND eliminated_at IS NULL),
			(SELECT COUNT(*) FROM room_round_votes WHERE room_id = $1 AND round_number = $2 AND cycle_number = $3)`,
		input.RoomID, roundNumber, cycleNumber,
	).Scan(&eligibleVoters, &votesCast); err != nil {
		return nil, err
	}

	finalizedNow := votesCast >= eligibleVoters
	if finalizedNow {
		if err := completeVoting(ctx, tx, input.RoomID, roundNumber, cycleNumber, input.VotedAt); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	state, err := r.FindCurrentRoundState(ctx, input.RoomID, input.VoterUserID, input.VotedAt)
	if state != nil {
		state.FinalizedNow = finalizedNow
	}
	return state, err
}

func (r *RoomRepository) SubmitMrWhiteGuess(ctx context.Context, input roomports.MrWhiteGuessInput) (*domainroom.RoundState, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := advanceGameState(ctx, tx, input.RoomID, input.At); err != nil {
		return nil, err
	}
	var roundNumber int
	var phase domainroom.RoundPhase
	var mrWhiteUserID pgtype.Int8
	var commonWord string
	var previousGuess pgtype.Text
	err = tx.QueryRow(ctx, `
		SELECT r.current_round, rr.phase, rr.mr_white_user_id, rr.common_word, rr.mr_white_guess
		FROM rooms r
		JOIN room_rounds rr ON rr.room_id = r.id AND rr.round_number = r.current_round
		WHERE r.id = $1
		FOR UPDATE OF rr`, input.RoomID,
	).Scan(&roundNumber, &phase, &mrWhiteUserID, &commonWord, &previousGuess)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrRoundCardNotFound
	}
	if err != nil {
		return nil, err
	}
	if phase != domainroom.RoundPhaseMrWhiteGuessing || !mrWhiteUserID.Valid || mrWhiteUserID.Int64 != input.UserID {
		return nil, domainroom.ErrMrWhiteGuessNotAllowed
	}
	if previousGuess.Valid {
		return nil, domainroom.ErrMrWhiteAlreadyGuessed
	}
	correct := strings.EqualFold(strings.TrimSpace(input.Guess), strings.TrimSpace(commonWord))
	if _, err := tx.Exec(ctx, `
		UPDATE room_rounds
		SET mr_white_guess = $3, mr_white_guess_correct = $4
		WHERE room_id = $1 AND round_number = $2`, input.RoomID, roundNumber, strings.TrimSpace(input.Guess), correct,
	); err != nil {
		return nil, err
	}
	if correct {
		if _, err := tx.Exec(ctx, `
			UPDATE room_rounds SET phase = $3, phase_deadline_at = NULL, winner = $4
			WHERE room_id = $1 AND round_number = $2`, input.RoomID, roundNumber,
			domainroom.RoundPhaseGameFinished, domainroom.WinningSideMrWhite,
		); err != nil {
			return nil, err
		}
	} else if err := continueAfterResult(ctx, tx, input.RoomID, roundNumber, input.At, false); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindCurrentRoundState(ctx, input.RoomID, input.UserID, input.At)
}

type roundStateQueryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func loadCurrentRoundState(
	ctx context.Context,
	queryer roundStateQueryer,
	roomID string,
	userID int64,
	now time.Time,
) (*domainroom.RoundState, error) {
	state := &domainroom.RoundState{RoomID: roomID}
	var phaseDeadline pgtype.Timestamptz
	var currentVote pgtype.Int8
	var undercoverPlayer pgtype.Int8
	var mrWhitePlayer pgtype.Int8
	var eliminatedPlayer pgtype.Int8
	var eliminatedRole pgtype.Text
	var winner pgtype.Text
	var guessCorrect pgtype.Bool
	var currentTurnPlayer pgtype.Int8
	var turnEndsAt pgtype.Timestamptz
	err := queryer.QueryRow(ctx, `
		SELECT r.current_round, rr.cycle_number, rr.phase, rr.phase_deadline_at,
			(SELECT COUNT(*) FROM room_round_ready_players WHERE room_id = r.id AND round_number = r.current_round),
			(SELECT COUNT(*) FROM room_members WHERE room_id = r.id),
			EXISTS (SELECT 1 FROM room_round_ready_players WHERE room_id = r.id AND round_number = r.current_round AND user_id = $2),
			(SELECT COUNT(*) FROM room_members WHERE room_id = r.id AND eliminated_at IS NULL),
			(SELECT COUNT(*) FROM room_round_votes WHERE room_id = r.id AND round_number = r.current_round AND cycle_number = rr.cycle_number),
			(SELECT target_user_id FROM room_round_votes
			 WHERE room_id = r.id AND round_number = r.current_round AND cycle_number = rr.cycle_number AND voter_user_id = $2),
			rr.undercover_user_id, rr.mr_white_user_id, rr.eliminated_user_id,
			rr.eliminated_role, rr.winner, rr.mr_white_guess_correct,
			(SELECT user_id FROM room_round_turns WHERE room_id = r.id AND round_number = r.current_round
			 AND cycle_number = rr.cycle_number AND started_at IS NOT NULL AND finished_at IS NULL ORDER BY turn_number LIMIT 1),
			COALESCE((SELECT turn_number FROM room_round_turns WHERE room_id = r.id AND round_number = r.current_round
			 AND cycle_number = rr.cycle_number AND started_at IS NOT NULL AND finished_at IS NULL ORDER BY turn_number LIMIT 1), 0),
			(SELECT COUNT(*) FROM room_round_turns WHERE room_id = r.id AND round_number = r.current_round AND cycle_number = rr.cycle_number),
			(SELECT deadline_at FROM room_round_turns WHERE room_id = r.id AND round_number = r.current_round
			 AND cycle_number = rr.cycle_number AND started_at IS NOT NULL AND finished_at IS NULL ORDER BY turn_number LIMIT 1)
		FROM rooms r
		JOIN room_members current_member
			ON current_member.room_id = r.id AND current_member.user_id = $2
		JOIN room_rounds rr
			ON rr.room_id = r.id AND rr.round_number = r.current_round
		WHERE r.id = $1 AND r.expires_at > now() AND r.status = $3`,
		roomID, userID, domainroom.StatusPlaying,
	).Scan(
		&state.RoundNumber,
		&state.CycleNumber,
		&state.Phase,
		&phaseDeadline,
		&state.ReadyPlayers,
		&state.EligiblePlayers,
		&state.CurrentUserReady,
		&state.EligibleVoters,
		&state.VotesCast,
		&currentVote,
		&undercoverPlayer,
		&mrWhitePlayer,
		&eliminatedPlayer,
		&eliminatedRole,
		&winner,
		&guessCorrect,
		&currentTurnPlayer,
		&state.TurnNumber,
		&state.TotalTurns,
		&turnEndsAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrRoundCardNotFound
	}
	if err != nil {
		return nil, err
	}

	if phaseDeadline.Valid {
		deadline := phaseDeadline.Time
		state.PhaseDeadlineAt = &deadline
	}
	if currentTurnPlayer.Valid {
		playerID := currentTurnPlayer.Int64
		state.CurrentTurnPlayerID = &playerID
	}
	if turnEndsAt.Valid {
		deadline := turnEndsAt.Time
		state.TurnEndsAt = &deadline
	}
	if state.Phase == domainroom.RoundPhaseGameFinished {
		undercover := undercoverPlayer.Int64
		state.UndercoverPlayerID = &undercover
		if mrWhitePlayer.Valid {
			white := mrWhitePlayer.Int64
			state.MrWhitePlayerID = &white
		}
	}
	if eliminatedPlayer.Valid {
		eliminated := eliminatedPlayer.Int64
		state.EliminatedPlayerID = &eliminated
	}
	if eliminatedRole.Valid {
		role := domainroom.CardRole(eliminatedRole.String)
		state.EliminatedRole = &role
		if role == domainroom.CardRoleUndercover && undercoverPlayer.Valid {
			undercover := undercoverPlayer.Int64
			state.UndercoverPlayerID = &undercover
		}
		if role == domainroom.CardRoleMrWhite && mrWhitePlayer.Valid {
			white := mrWhitePlayer.Int64
			state.MrWhitePlayerID = &white
		}
	}
	if winner.Valid {
		value := domainroom.WinningSide(winner.String)
		state.Winner = &value
	}
	if guessCorrect.Valid {
		value := guessCorrect.Bool
		state.MrWhiteGuessCorrect = &value
	}
	if currentVote.Valid {
		vote := currentVote.Int64
		state.CurrentUserVoteID = &vote
	}
	return state, nil
}

type voteTally struct {
	playerID int64
	votes    int
}

func advanceGameState(ctx context.Context, tx pgx.Tx, roomID string, now time.Time) (bool, error) {
	changed := false
	for range 8 {
		var roundNumber, cycleNumber int
		var phase domainroom.RoundPhase
		var deadline pgtype.Timestamptz
		var eliminatedRole pgtype.Text
		err := tx.QueryRow(ctx, `
			SELECT r.current_round, rr.cycle_number, rr.phase, rr.phase_deadline_at, rr.eliminated_role
			FROM rooms r
			JOIN room_rounds rr ON rr.room_id = r.id AND rr.round_number = r.current_round
			WHERE r.id = $1 AND r.status = $2
			FOR UPDATE OF rr`, roomID, domainroom.StatusPlaying,
		).Scan(&roundNumber, &cycleNumber, &phase, &deadline, &eliminatedRole)
		if errors.Is(err, pgx.ErrNoRows) {
			return changed, nil
		}
		if err != nil {
			return false, err
		}
		if phase == domainroom.RoundPhaseGameFinished || !deadline.Valid || now.Before(deadline.Time) {
			return changed, nil
		}

		switch phase {
		case domainroom.RoundPhaseRevealingRole:
			if err := startDescribing(ctx, tx, roomID, roundNumber, cycleNumber, now); err != nil {
				return false, err
			}
		case domainroom.RoundPhaseDescribing:
			if err := finishCurrentTurn(ctx, tx, roomID, roundNumber, cycleNumber, now, true); err != nil {
				return false, err
			}
		case domainroom.RoundPhaseVoting:
			if err := completeVoting(ctx, tx, roomID, roundNumber, cycleNumber, now); err != nil {
				return false, err
			}
		case domainroom.RoundPhaseRevealingResult:
			if eliminatedRole.Valid && domainroom.CardRole(eliminatedRole.String) == domainroom.CardRoleMrWhite {
				_, err := tx.Exec(ctx, `
					UPDATE room_rounds SET phase = $3, phase_deadline_at = $4
					WHERE room_id = $1 AND round_number = $2`, roomID, roundNumber,
					domainroom.RoundPhaseMrWhiteGuessing,
					now.Add(time.Duration(domainroom.MrWhiteGuessDurationSeconds)*time.Second),
				)
				if err != nil {
					return false, err
				}
			} else if err := continueAfterResult(ctx, tx, roomID, roundNumber, now, false); err != nil {
				return false, err
			}
		case domainroom.RoundPhaseMrWhiteGuessing:
			if _, err := tx.Exec(ctx, `
				UPDATE room_rounds SET mr_white_guess_correct = FALSE
				WHERE room_id = $1 AND round_number = $2 AND mr_white_guess_correct IS NULL`, roomID, roundNumber,
			); err != nil {
				return false, err
			}
			if err := continueAfterResult(ctx, tx, roomID, roundNumber, now, false); err != nil {
				return false, err
			}
		default:
			return false, domainroom.ErrInvalidGameState
		}
		changed = true
	}
	return changed, nil
}

func startDescribing(ctx context.Context, tx pgx.Tx, roomID string, roundNumber, cycleNumber int, now time.Time) error {
	var turnSeconds int
	if err := tx.QueryRow(ctx, `SELECT discussion_seconds FROM rooms WHERE id = $1`, roomID).Scan(&turnSeconds); err != nil {
		return err
	}
	var turnNumber int
	err := tx.QueryRow(ctx, `
		SELECT turn_number FROM room_round_turns
		WHERE room_id = $1 AND round_number = $2 AND cycle_number = $3 AND started_at IS NULL
		ORDER BY turn_number LIMIT 1`, roomID, roundNumber, cycleNumber,
	).Scan(&turnNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return domainroom.ErrInvalidGameState
	}
	if err != nil {
		return err
	}
	deadline := now.Add(time.Duration(turnSeconds) * time.Second)
	if _, err := tx.Exec(ctx, `
		UPDATE room_round_turns SET started_at = $5, deadline_at = $6
		WHERE room_id = $1 AND round_number = $2 AND cycle_number = $3 AND turn_number = $4`,
		roomID, roundNumber, cycleNumber, turnNumber, now, deadline,
	); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE room_rounds SET phase = $3, phase_deadline_at = $4
		WHERE room_id = $1 AND round_number = $2`, roomID, roundNumber,
		domainroom.RoundPhaseDescribing, deadline,
	)
	return err
}

func finishCurrentTurn(
	ctx context.Context,
	tx pgx.Tx,
	roomID string,
	roundNumber, cycleNumber int,
	now time.Time,
	skipped bool,
) error {
	result, err := tx.Exec(ctx, `
		UPDATE room_round_turns SET finished_at = $4, skipped = $5
		WHERE room_id = $1 AND round_number = $2 AND cycle_number = $3
		  AND started_at IS NOT NULL AND finished_at IS NULL`, roomID, roundNumber, cycleNumber, now, skipped,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domainroom.ErrInvalidGameState
	}
	var nextTurn int
	err = tx.QueryRow(ctx, `
		SELECT turn_number FROM room_round_turns
		WHERE room_id = $1 AND round_number = $2 AND cycle_number = $3 AND started_at IS NULL
		ORDER BY turn_number LIMIT 1`, roomID, roundNumber, cycleNumber,
	).Scan(&nextTurn)
	if errors.Is(err, pgx.ErrNoRows) {
		deadline := now.Add(time.Duration(domainroom.VotingDurationSeconds) * time.Second)
		_, err = tx.Exec(ctx, `
			UPDATE room_rounds SET phase = $3, phase_deadline_at = $4, voting_completed_at = NULL,
				eliminated_user_id = NULL, eliminated_role = NULL
			WHERE room_id = $1 AND round_number = $2`, roomID, roundNumber, domainroom.RoundPhaseVoting, deadline,
		)
		return err
	}
	if err != nil {
		return err
	}
	var turnSeconds int
	if err := tx.QueryRow(ctx, `SELECT discussion_seconds FROM rooms WHERE id = $1`, roomID).Scan(&turnSeconds); err != nil {
		return err
	}
	deadline := now.Add(time.Duration(turnSeconds) * time.Second)
	if _, err := tx.Exec(ctx, `
		UPDATE room_round_turns SET started_at = $5, deadline_at = $6
		WHERE room_id = $1 AND round_number = $2 AND cycle_number = $3 AND turn_number = $4`,
		roomID, roundNumber, cycleNumber, nextTurn, now, deadline,
	); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE room_rounds SET phase_deadline_at = $3 WHERE room_id = $1 AND round_number = $2`,
		roomID, roundNumber, deadline)
	return err
}

func continueAfterResult(
	ctx context.Context,
	tx pgx.Tx,
	roomID string,
	roundNumber int,
	now time.Time,
	mrWhiteGuessCorrect bool,
) error {
	var totalAlive, undercoverAlive, mrWhiteAlive, currentCycle int
	if err := tx.QueryRow(ctx, `
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE m.user_id = rr.undercover_user_id),
			COUNT(*) FILTER (WHERE m.user_id = rr.mr_white_user_id),
			rr.cycle_number
		FROM room_rounds rr
		JOIN room_members m ON m.room_id = rr.room_id AND m.eliminated_at IS NULL
		WHERE rr.room_id = $1 AND rr.round_number = $2
		GROUP BY rr.cycle_number`, roomID, roundNumber,
	).Scan(&totalAlive, &undercoverAlive, &mrWhiteAlive, &currentCycle); err != nil {
		return err
	}
	winner := gameengine.EvaluateWinner(gameengine.AliveRoles{
		Civilians:  totalAlive - undercoverAlive - mrWhiteAlive,
		Undercover: undercoverAlive,
		MrWhite:    mrWhiteAlive,
	}, mrWhiteGuessCorrect)
	if winner != nil {
		_, err := tx.Exec(ctx, `
			UPDATE room_rounds SET phase = $3, phase_deadline_at = NULL, winner = $4
			WHERE room_id = $1 AND round_number = $2`, roomID, roundNumber,
			domainroom.RoundPhaseGameFinished, *winner,
		)
		return err
	}
	nextCycle := currentCycle + 1
	if _, err := tx.Exec(ctx, `
		INSERT INTO room_round_turns (room_id, round_number, cycle_number, turn_number, user_id)
		SELECT $1, $2, $3, ROW_NUMBER() OVER (ORDER BY random()), user_id
		FROM room_members WHERE room_id = $1 AND eliminated_at IS NULL`, roomID, roundNumber, nextCycle,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE room_rounds SET cycle_number = $3, eliminated_user_id = NULL,
			eliminated_role = NULL, voting_completed_at = NULL
		WHERE room_id = $1 AND round_number = $2`, roomID, roundNumber, nextCycle,
	); err != nil {
		return err
	}
	return startDescribing(ctx, tx, roomID, roundNumber, nextCycle, now)
}

func completeVoting(ctx context.Context, tx pgx.Tx, roomID string, roundNumber, cycleNumber int, completedAt time.Time) error {
	rows, err := tx.Query(ctx, `
		SELECT target_user_id, COUNT(*) AS vote_count
		FROM room_round_votes
		WHERE room_id = $1 AND round_number = $2 AND cycle_number = $3
		GROUP BY target_user_id
		ORDER BY vote_count DESC, target_user_id
		LIMIT 2`, roomID, roundNumber, cycleNumber)
	if err != nil {
		return err
	}
	tallies := make([]voteTally, 0, 2)
	for rows.Next() {
		var item voteTally
		if err := rows.Scan(&item.playerID, &item.votes); err != nil {
			rows.Close()
			return err
		}
		tallies = append(tallies, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	var eliminatedPlayerID *int64
	if len(tallies) == 1 || (len(tallies) > 1 && tallies[0].votes > tallies[1].votes) {
		eliminatedPlayerID = &tallies[0].playerID
	}
	var eliminatedRole *domainroom.CardRole
	if eliminatedPlayerID != nil {
		var undercoverID int64
		var mrWhiteID pgtype.Int8
		if err := tx.QueryRow(ctx, `SELECT undercover_user_id, mr_white_user_id FROM room_rounds
			WHERE room_id = $1 AND round_number = $2`, roomID, roundNumber,
		).Scan(&undercoverID, &mrWhiteID); err != nil {
			return err
		}
		role := domainroom.CardRoleCivilian
		if *eliminatedPlayerID == undercoverID {
			role = domainroom.CardRoleUndercover
		} else if mrWhiteID.Valid && *eliminatedPlayerID == mrWhiteID.Int64 {
			role = domainroom.CardRoleMrWhite
		}
		eliminatedRole = &role
		if _, err := tx.Exec(ctx, `
			UPDATE room_members
			SET eliminated_at = COALESCE(eliminated_at, $3)
			WHERE room_id = $1 AND user_id = $2`, roomID, *eliminatedPlayerID, completedAt,
		); err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `
		UPDATE room_rounds
		SET phase = $3, phase_deadline_at = $4, voting_completed_at = $4,
			eliminated_user_id = $5, eliminated_role = $6
		WHERE room_id = $1 AND round_number = $2`,
		roomID, roundNumber, domainroom.RoundPhaseRevealingResult,
		completedAt.Add(time.Duration(domainroom.ResultRevealDurationSeconds)*time.Second),
		eliminatedPlayerID, eliminatedRole,
	)
	return err
}

func (r *RoomRepository) FindCurrentCard(ctx context.Context, roomID string, userID int64) (*domainroom.RoundCard, error) {
	card := &domainroom.RoundCard{RoomID: roomID, PlayerID: userID}
	var commonWord string
	var differentWord string
	var undercoverUserID int64
	var mrWhiteUserID pgtype.Int8
	err := r.db.QueryRow(ctx, `
			SELECT r.current_round, rr.common_word, rr.different_word,
				rr.undercover_user_id, rr.mr_white_user_id, rr.dealt_at, rr.phase, rr.phase_deadline_at
		FROM rooms r
			JOIN room_members rm
				ON rm.room_id = r.id AND rm.user_id = $2
		JOIN room_rounds rr
			ON rr.room_id = r.id AND rr.round_number = r.current_round
		WHERE r.id = $1 AND r.expires_at > now() AND r.status = $3`,
		roomID,
		userID,
		domainroom.StatusPlaying,
	).Scan(
		&card.RoundNumber,
		&commonWord,
		&differentWord,
		&undercoverUserID,
		&mrWhiteUserID,
		&card.DealtAt,
		&card.Phase,
		&card.PhaseDeadlineAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainroom.ErrRoundCardNotFound
	}
	if err != nil {
		return nil, err
	}

	card.Role = domainroom.CardRoleCivilian
	card.Word = commonWord
	if userID == undercoverUserID {
		card.Role = domainroom.CardRoleUndercover
		card.Word = differentWord
	} else if mrWhiteUserID.Valid && userID == mrWhiteUserID.Int64 {
		card.Role = domainroom.CardRoleMrWhite
		card.Word = ""
	}
	return card, nil
}

type roomQueryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func (r *RoomRepository) loadMembers(
	ctx context.Context,
	queryer roomQueryer,
	roomID string,
	hostUserID int64,
) ([]domainroom.Member, error) {
	rows, err := queryer.Query(ctx, `
		SELECT m.user_id, u.username, m.seat_number, m.joined_at, m.eliminated_at IS NOT NULL
		FROM room_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.room_id = $1
		ORDER BY m.seat_number`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]domainroom.Member, 0)
	for rows.Next() {
		member := domainroom.Member{}
		if err := rows.Scan(&member.UserID, &member.UserName, &member.SeatNumber, &member.JoinedAt, &member.Eliminated); err != nil {
			return nil, err
		}
		member.IsHost = member.UserID == hostUserID
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

func roomScanTargets(room *domainroom.Room) []any {
	return []any{
		&room.ID,
		&room.InviteCode,
		&room.Name,
		&room.LanguageCode,
		&room.HostUserID,
		&room.Status,
		&room.MaxPlayers,
		&room.CurrentRound,
		&room.DiscussionSeconds,
		&room.MrWhiteEnabled,
		&room.Version,
		&room.ExpiresAt,
		&room.CreatedAt,
		&room.UpdatedAt,
	}
}
