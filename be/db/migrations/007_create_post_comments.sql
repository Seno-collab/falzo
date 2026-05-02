CREATE TABLE IF NOT EXISTS post_comments (
    id         BIGSERIAL PRIMARY KEY,
    post_id    BIGINT      NOT NULL,
    user_id    BIGINT      NOT NULL,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_post_comments_content_not_empty CHECK (length(trim(content)) > 0),
    CONSTRAINT chk_post_comments_content_length CHECK (length(content) <= 1000),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_post_comments_post_id_created_at
    ON post_comments (post_id, created_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_post_comments_user_id
    ON post_comments (user_id);
