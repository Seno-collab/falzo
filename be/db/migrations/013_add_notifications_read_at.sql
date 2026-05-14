ALTER TABLE notifications
    ADD COLUMN IF NOT EXISTS read_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_notifications_user_read_at
    ON notifications (user_id, read_at);
