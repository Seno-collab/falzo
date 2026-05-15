ALTER TABLE posts
    ADD COLUMN IF NOT EXISTS status VARCHAR(24) NOT NULL DEFAULT 'visible',
    ADD COLUMN IF NOT EXISTS hidden_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS hidden_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS deleted_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS moderation_reason TEXT NULL;

ALTER TABLE post_comments
    ADD COLUMN IF NOT EXISTS status VARCHAR(24) NOT NULL DEFAULT 'visible',
    ADD COLUMN IF NOT EXISTS hidden_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS hidden_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS deleted_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS moderation_reason TEXT NULL;

CREATE TABLE IF NOT EXISTS content_reports (
    id               BIGSERIAL PRIMARY KEY,
    reporter_user_id BIGINT      NOT NULL,
    post_id          BIGINT      NULL,
    comment_id       BIGINT      NULL,
    reason           TEXT        NOT NULL,
    status           VARCHAR(24) NOT NULL DEFAULT 'open',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reviewed_at      TIMESTAMPTZ NULL,
    reviewed_by      BIGINT      NULL,
    CONSTRAINT chk_content_reports_target CHECK (
        (post_id IS NOT NULL AND comment_id IS NULL)
        OR (post_id IS NOT NULL AND comment_id IS NOT NULL)
    ),
    CONSTRAINT chk_content_reports_reason_not_blank CHECK (length(btrim(reason)) > 0),
    CONSTRAINT fk_content_reports_reporter FOREIGN KEY (reporter_user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_content_reports_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    CONSTRAINT fk_content_reports_comment FOREIGN KEY (comment_id) REFERENCES post_comments(id) ON DELETE CASCADE,
    CONSTRAINT fk_content_reports_reviewer FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_content_reports_post_id
    ON content_reports (post_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_content_reports_comment_id
    ON content_reports (comment_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_content_reports_status
    ON content_reports (status, created_at DESC);

CREATE TABLE IF NOT EXISTS user_blocks (
    id              BIGSERIAL PRIMARY KEY,
    blocker_user_id BIGINT      NOT NULL,
    blocked_user_id BIGINT      NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_user_blocks_not_self CHECK (blocker_user_id <> blocked_user_id),
    CONSTRAINT uq_user_blocks UNIQUE (blocker_user_id, blocked_user_id),
    CONSTRAINT fk_user_blocks_blocker FOREIGN KEY (blocker_user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_blocks_blocked FOREIGN KEY (blocked_user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_blocks_blocked_user_id
    ON user_blocks (blocked_user_id);

CREATE INDEX IF NOT EXISTS idx_posts_visible_feed
    ON posts (status, deleted_at, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_post_comments_visible
    ON post_comments (post_id, status, deleted_at, created_at ASC, id ASC);
