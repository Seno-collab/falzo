package chat

import "time"

type Message struct {
	ID       string    `json:"id"`
	RoomID   string    `json:"-"`
	UserID   int64     `json:"user_id"`
	UserName string    `json:"username"`
	Text     string    `json:"text"`
	SentAt   time.Time `json:"sent_at"`
}
