CREATE TABLE IF NOT EXISTS saved_collections (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_saved_collections_name_not_blank CHECK (length(btrim(name)) > 0),
    CONSTRAINT chk_saved_collections_name_length CHECK (length(name) <= 120),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_saved_collections_user_name_lower
    ON saved_collections (user_id, lower(name));

CREATE INDEX IF NOT EXISTS idx_saved_collections_user_id
    ON saved_collections (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS saved_collection_posts (
    id            BIGSERIAL PRIMARY KEY,
    collection_id BIGINT      NOT NULL,
    post_id       BIGINT      NOT NULL,
    user_id       BIGINT      NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_saved_collection_posts UNIQUE (collection_id, post_id),
    FOREIGN KEY (collection_id) REFERENCES saved_collections(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id, user_id) REFERENCES post_saves(post_id, user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_saved_collection_posts_collection_id
    ON saved_collection_posts (collection_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_saved_collection_posts_user_post
    ON saved_collection_posts (user_id, post_id);
