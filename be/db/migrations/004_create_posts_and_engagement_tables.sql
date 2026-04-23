CREATE TABLE IF NOT EXISTS posts (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT           NOT NULL,
    image_url     TEXT             NOT NULL,
    caption       TEXT             NOT NULL DEFAULT '',
    location_id   UUID             NULL,
    location_name TEXT             NULL,
    latitude      DOUBLE PRECISION NULL,
    longitude     DOUBLE PRECISION NULL,
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_posts_latitude_range CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    CONSTRAINT chk_posts_longitude_range CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180),
    CONSTRAINT fk_posts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_posts_location_id FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_posts_created_at
    ON posts (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_posts_location_id
    ON posts (location_id);

CREATE INDEX IF NOT EXISTS idx_posts_location_name
    ON posts (location_name);

CREATE TABLE IF NOT EXISTS post_likes (
    id         BIGSERIAL PRIMARY KEY,
    post_id    BIGINT      NOT NULL,
    user_id    BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_post_likes UNIQUE (post_id, user_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_post_likes_post_id
    ON post_likes (post_id);

CREATE INDEX IF NOT EXISTS idx_post_likes_user_id
    ON post_likes (user_id);

CREATE TABLE IF NOT EXISTS post_saves (
    id         BIGSERIAL PRIMARY KEY,
    post_id    BIGINT      NOT NULL,
    user_id    BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_post_saves UNIQUE (post_id, user_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_post_saves_post_id
    ON post_saves (post_id);

CREATE INDEX IF NOT EXISTS idx_post_saves_user_id
    ON post_saves (user_id);
