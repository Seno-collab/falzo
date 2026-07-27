package postgres

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepository struct {
	db *pgxpool.Pool
}

const roomColumns = `id::text, invite_code, name, language_code, host_user_id, status, max_players, current_round, version, expires_at, created_at, updated_at`

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
			max_players, current_round, version, expires_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING `+roomColumns,
		room.ID,
		room.InviteCode,
		room.Name,
		room.LanguageCode,
		room.HostUserID,
		room.Status,
		room.MaxPlayers,
		room.CurrentRound,
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

	var undercoverIsMember bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM room_members
			WHERE room_id = $1 AND user_id = $2
		)`, input.RoomID, input.UndercoverPlayerID).Scan(&undercoverIsMember); err != nil {
		return nil, err
	}
	if !undercoverIsMember {
		return nil, domainroom.ErrRoundCardNotFound
	}

	nextRound := currentRound + 1
	if _, err := tx.Exec(ctx, `
		INSERT INTO room_rounds (
			room_id, round_number, word_pair_id, common_word,
			different_word, undercover_user_id, dealt_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		input.RoomID,
		nextRound,
		input.WordPairID,
		input.CommonWord,
		input.DifferentWord,
		input.UndercoverPlayerID,
		input.DealtAt,
	); err != nil {
		return nil, err
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
		RoomID:      input.RoomID,
		RoundNumber: nextRound,
		PlayerID:    input.HostUserID,
		Role:        domainroom.CardRoleCivilian,
		Word:        input.CommonWord,
		DealtAt:     input.DealtAt,
	}
	if input.HostUserID == input.UndercoverPlayerID {
		card.Role = domainroom.CardRoleUndercover
		card.Word = input.DifferentWord
	}
	return card, nil
}

func (r *RoomRepository) FindCurrentCard(ctx context.Context, roomID string, userID int64) (*domainroom.RoundCard, error) {
	card := &domainroom.RoundCard{RoomID: roomID, PlayerID: userID}
	var commonWord string
	var differentWord string
	var undercoverUserID int64
	err := r.db.QueryRow(ctx, `
		SELECT r.current_round, rr.common_word, rr.different_word,
			rr.undercover_user_id, rr.dealt_at
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
		&card.DealtAt,
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
		SELECT m.user_id, u.username, m.seat_number, m.joined_at
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
		if err := rows.Scan(&member.UserID, &member.UserName, &member.SeatNumber, &member.JoinedAt); err != nil {
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
		&room.Version,
		&room.ExpiresAt,
		&room.CreatedAt,
		&room.UpdatedAt,
	}
}
