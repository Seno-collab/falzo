CREATE TABLE IF NOT EXISTS post_categories (
    post_id     BIGINT NOT NULL,
    category_id BIGINT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, category_id),
    CONSTRAINT fk_post_categories_post_id FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    CONSTRAINT fk_post_categories_category_id FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

INSERT INTO post_categories (post_id, category_id)
SELECT id, category_id
FROM posts
WHERE category_id IS NOT NULL
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_post_categories_category_created
    ON post_categories (category_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_post_categories_post_created
    ON post_categories (post_id, created_at ASC);
