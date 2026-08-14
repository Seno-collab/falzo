CREATE TABLE room_chat_messages (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text VARCHAR(500) NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT room_chat_messages_text_not_blank CHECK (length(trim(text)) > 0)
);

CREATE INDEX room_chat_messages_room_sent_idx
    ON room_chat_messages (room_id, sent_at DESC, id DESC);
