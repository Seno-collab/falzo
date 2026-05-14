CREATE TABLE IF NOT EXISTS user_follows (
    id           BIGSERIAL PRIMARY KEY,
    follower_id  BIGINT      NOT NULL,
    following_id BIGINT      NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_follows UNIQUE (follower_id, following_id),
    CONSTRAINT chk_user_follows_not_self CHECK (follower_id <> following_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_follows_follower_id
    ON user_follows (follower_id);

CREATE INDEX IF NOT EXISTS idx_user_follows_following_id
    ON user_follows (following_id);
