package postgres

import (
	domainsocial "be/internal/domain/social"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SocialRepository struct {
	db *pgxpool.Pool
}

func NewSocialRepository(db *pgxpool.Pool) *SocialRepository {
	return &SocialRepository{db: db}
}

func (r *SocialRepository) SearchUsers(
	ctx context.Context,
	currentUserID int64,
	query string,
	limit int,
) ([]domainsocial.UserSummary, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.username,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM friendships f
					WHERE f.user_id_low = LEAST($1, u.id) AND f.user_id_high = GREATEST($1, u.id)
				) THEN 'FRIENDS'
				WHEN EXISTS (
					SELECT 1 FROM friend_requests fr
					WHERE fr.sender_id = u.id AND fr.receiver_id = $1 AND fr.status = 'PENDING'
				) THEN 'INCOMING_REQUEST'
				WHEN EXISTS (
					SELECT 1 FROM friend_requests fr
					WHERE fr.sender_id = $1 AND fr.receiver_id = u.id AND fr.status = 'PENDING'
				) THEN 'OUTGOING_REQUEST'
				ELSE 'NONE'
			END
		FROM users u
		WHERE u.id <> $1
		  AND u.status = 'ACTIVE'
		  AND lower(u.username) LIKE lower($2) || '%' ESCAPE '\'
		ORDER BY CASE WHEN lower(u.username) = lower($3) THEN 0 ELSE 1 END, lower(u.username), u.id
		LIMIT $4`, currentUserID, escapeLikePrefix(query), query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domainsocial.UserSummary, 0)
	for rows.Next() {
		var user domainsocial.UserSummary
		if err := rows.Scan(&user.ID, &user.UserName, &user.Relationship); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *SocialRepository) CreateFriendRequest(
	ctx context.Context,
	senderID,
	receiverID int64,
	now time.Time,
) (*domainsocial.FriendRequest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id
		FROM users
		WHERE id IN ($1, $2) AND status = 'ACTIVE'
		ORDER BY id
		FOR UPDATE`, senderID, receiverID)
	if err != nil {
		return nil, err
	}
	userCount := 0
	for rows.Next() {
		userCount++
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if userCount != 2 {
		return nil, domainsocial.ErrUserNotFound
	}

	var alreadyFriends bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM friendships
			WHERE user_id_low = LEAST($1::bigint, $2::bigint)
			  AND user_id_high = GREATEST($1::bigint, $2::bigint)
		)`, senderID, receiverID).Scan(&alreadyFriends); err != nil {
		return nil, err
	}
	if alreadyFriends {
		return nil, domainsocial.ErrAlreadyFriends
	}

	var requestID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO friend_requests (sender_id, receiver_id, status, created_at, updated_at)
		VALUES ($1, $2, 'PENDING', $3, $3)
		RETURNING id`, senderID, receiverID, now).Scan(&requestID)
	if err != nil {
		if IsUniqueViolation(err) {
			return nil, domainsocial.ErrFriendRequestExists
		}
		if IsErrorCode(err, CodeForeignKeyViolation) {
			return nil, domainsocial.ErrUserNotFound
		}
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO notifications (user_id, actor_id, type, reference_id, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		receiverID,
		senderID,
		domainsocial.NotificationFriendRequestReceived,
		requestID,
		now,
	); err != nil {
		return nil, err
	}

	request, err := findFriendRequest(ctx, tx, requestID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return request, nil
}

func (r *SocialRepository) ListPendingFriendRequests(ctx context.Context, userID int64) ([]domainsocial.FriendRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT fr.id, fr.sender_id, sender.username, fr.receiver_id, receiver.username,
		       fr.status, fr.created_at, fr.responded_at
		FROM friend_requests fr
		JOIN users sender ON sender.id = fr.sender_id
		JOIN users receiver ON receiver.id = fr.receiver_id
		WHERE (fr.sender_id = $1 OR fr.receiver_id = $1) AND fr.status = 'PENDING'
		ORDER BY fr.created_at DESC, fr.id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := make([]domainsocial.FriendRequest, 0)
	for rows.Next() {
		request, err := scanFriendRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, *request)
	}
	return requests, rows.Err()
}

func (r *SocialRepository) RespondFriendRequest(
	ctx context.Context,
	userID,
	requestID int64,
	accept bool,
	now time.Time,
) (*domainsocial.FriendRequest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var senderID, receiverID int64
	var status domainsocial.FriendRequestStatus
	err = tx.QueryRow(ctx, `
		SELECT sender_id, receiver_id, status
		FROM friend_requests
		WHERE id = $1
		FOR UPDATE`, requestID).Scan(&senderID, &receiverID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainsocial.ErrFriendRequestNotFound
	}
	if err != nil {
		return nil, err
	}
	if receiverID != userID {
		return nil, domainsocial.ErrFriendRequestForbidden
	}
	if status != domainsocial.FriendRequestPending {
		return nil, domainsocial.ErrFriendRequestNotPending
	}

	nextStatus := domainsocial.FriendRequestRejected
	if accept {
		nextStatus = domainsocial.FriendRequestAccepted
		if _, err := tx.Exec(ctx, `
			INSERT INTO friendships (user_id_low, user_id_high, created_at)
			VALUES (LEAST($1::bigint, $2::bigint), GREATEST($1::bigint, $2::bigint), $3)
			ON CONFLICT (user_id_low, user_id_high) DO NOTHING`, senderID, receiverID, now); err != nil {
			return nil, err
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE friend_requests
		SET status = $2, responded_at = $3, updated_at = $3
		WHERE id = $1`, requestID, nextStatus, now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, $2)
		WHERE user_id = $1 AND reference_id = $3 AND type = $4`,
		receiverID,
		now,
		requestID,
		domainsocial.NotificationFriendRequestReceived,
	); err != nil {
		return nil, err
	}
	if accept {
		if _, err := tx.Exec(ctx, `
			INSERT INTO notifications (user_id, actor_id, type, reference_id, created_at)
			VALUES ($1, $2, $3, $4, $5)`,
			senderID,
			receiverID,
			domainsocial.NotificationFriendRequestAccepted,
			requestID,
			now,
		); err != nil {
			return nil, err
		}
	}

	request, err := findFriendRequest(ctx, tx, requestID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return request, nil
}

func (r *SocialRepository) CancelFriendRequest(ctx context.Context, userID, requestID int64, now time.Time) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var senderID int64
	var status domainsocial.FriendRequestStatus
	err = tx.QueryRow(ctx, `
		SELECT sender_id, status
		FROM friend_requests
		WHERE id = $1
		FOR UPDATE`, requestID).Scan(&senderID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domainsocial.ErrFriendRequestNotFound
	}
	if err != nil {
		return err
	}
	if senderID != userID {
		return domainsocial.ErrFriendRequestForbidden
	}
	if status != domainsocial.FriendRequestPending {
		return domainsocial.ErrFriendRequestNotPending
	}

	if _, err := tx.Exec(ctx, `
		UPDATE friend_requests
		SET status = 'CANCELED', responded_at = $2, updated_at = $2
		WHERE id = $1`, requestID, now); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM notifications
		WHERE reference_id = $1 AND type = $2`,
		requestID,
		domainsocial.NotificationFriendRequestReceived,
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *SocialRepository) ListFriends(ctx context.Context, userID int64) ([]domainsocial.Friend, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.username, f.created_at
		FROM friendships f
		JOIN users u ON u.id = CASE
			WHEN f.user_id_low = $1 THEN f.user_id_high
			ELSE f.user_id_low
		END
		WHERE f.user_id_low = $1 OR f.user_id_high = $1
		ORDER BY lower(u.username), u.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	friends := make([]domainsocial.Friend, 0)
	for rows.Next() {
		var friend domainsocial.Friend
		if err := rows.Scan(&friend.ID, &friend.UserName, &friend.FriendsAt); err != nil {
			return nil, err
		}
		friends = append(friends, friend)
	}
	return friends, rows.Err()
}

func (r *SocialRepository) Unfriend(ctx context.Context, userID, friendUserID int64) error {
	command, err := r.db.Exec(ctx, `
		DELETE FROM friendships
		WHERE user_id_low = LEAST($1::bigint, $2::bigint)
		  AND user_id_high = GREATEST($1::bigint, $2::bigint)`, userID, friendUserID)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return domainsocial.ErrFriendshipNotFound
	}
	return nil
}

func (r *SocialRepository) ListNotifications(
	ctx context.Context,
	userID int64,
	unreadOnly bool,
	limit,
	offset int,
) ([]domainsocial.Notification, error) {
	rows, err := r.db.Query(ctx, `
		SELECT n.id, n.type, n.actor_id, actor.username, n.reference_id, n.read_at, n.created_at
		FROM notifications n
		JOIN users actor ON actor.id = n.actor_id
		WHERE n.user_id = $1 AND (NOT $2 OR n.read_at IS NULL)
		ORDER BY n.created_at DESC, n.id DESC
		LIMIT $3 OFFSET $4`, userID, unreadOnly, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]domainsocial.Notification, 0)
	for rows.Next() {
		var notification domainsocial.Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.Type,
			&notification.ActorID,
			&notification.ActorName,
			&notification.ReferenceID,
			&notification.ReadAt,
			&notification.CreatedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, rows.Err()
}

func (r *SocialRepository) CountUnreadNotifications(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND read_at IS NULL`, userID).Scan(&count)
	return count, err
}

func (r *SocialRepository) MarkNotificationRead(
	ctx context.Context,
	userID,
	notificationID int64,
	now time.Time,
) error {
	command, err := r.db.Exec(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, $3)
		WHERE id = $1 AND user_id = $2`, notificationID, userID, now)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return domainsocial.ErrNotificationNotFound
	}
	return nil
}

func (r *SocialRepository) MarkAllNotificationsRead(ctx context.Context, userID int64, now time.Time) (int64, error) {
	command, err := r.db.Exec(ctx, `
		UPDATE notifications
		SET read_at = $2
		WHERE user_id = $1 AND read_at IS NULL`, userID, now)
	if err != nil {
		return 0, err
	}
	return command.RowsAffected(), nil
}

func findFriendRequest(ctx context.Context, queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, requestID int64) (*domainsocial.FriendRequest, error) {
	return scanFriendRequest(queryer.QueryRow(ctx, `
		SELECT fr.id, fr.sender_id, sender.username, fr.receiver_id, receiver.username,
		       fr.status, fr.created_at, fr.responded_at
		FROM friend_requests fr
		JOIN users sender ON sender.id = fr.sender_id
		JOIN users receiver ON receiver.id = fr.receiver_id
		WHERE fr.id = $1`, requestID))
}

func scanFriendRequest(row rowScanner) (*domainsocial.FriendRequest, error) {
	request := &domainsocial.FriendRequest{}
	err := row.Scan(
		&request.ID,
		&request.SenderID,
		&request.SenderName,
		&request.ReceiverID,
		&request.ReceiverName,
		&request.Status,
		&request.CreatedAt,
		&request.RespondedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainsocial.ErrFriendRequestNotFound
	}
	if err != nil {
		return nil, err
	}
	return request, nil
}

func escapeLikePrefix(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	return strings.ReplaceAll(value, `_`, `\_`)
}
