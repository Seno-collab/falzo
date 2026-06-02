CREATE TABLE IF NOT EXISTS post_trust_votes (
    id         BIGSERIAL PRIMARY KEY,
    post_id    BIGINT      NOT NULL,
    user_id    BIGINT      NOT NULL,
    vote_type  VARCHAR(32) NOT NULL,
    reason     TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_post_trust_votes UNIQUE (post_id, user_id),
    CONSTRAINT chk_post_trust_votes_type CHECK (
        vote_type IN ('credible', 'suspicious', 'ai_generated', 'wrong_context', 'unsure')
    ),
    CONSTRAINT chk_post_trust_votes_reason_length CHECK (length(reason) <= 500),
    CONSTRAINT fk_post_trust_votes_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    CONSTRAINT fk_post_trust_votes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_post_trust_votes_post_id
    ON post_trust_votes (post_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_post_trust_votes_user_id
    ON post_trust_votes (user_id, updated_at DESC);
