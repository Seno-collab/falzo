package postgres

import (
	domainchat "be/internal/domain/chat"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Save(ctx context.Context, message domainchat.Message) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO room_chat_messages (id, room_id, user_id, text, sent_at)
		VALUES ($1, $2, $3, $4, $5)`,
		message.ID, message.RoomID, message.UserID, message.Text, message.SentAt,
	)
	return err
}

func (r *ChatRepository) ListRoom(
	ctx context.Context,
	roomID string,
	before *time.Time,
	limit int,
) ([]domainchat.Message, error) {
	rows, err := r.db.Query(ctx, `
		SELECT m.id::text, m.room_id::text, m.user_id, u.username, m.text, m.sent_at
		FROM room_chat_messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.room_id = $1 AND ($2::timestamptz IS NULL OR m.sent_at < $2)
		ORDER BY m.sent_at DESC, m.id DESC
		LIMIT $3`, roomID, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]domainchat.Message, 0, limit)
	for rows.Next() {
		var message domainchat.Message
		if err := rows.Scan(
			&message.ID, &message.RoomID, &message.UserID, &message.UserName, &message.Text, &message.SentAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}
	return messages, nil
}
