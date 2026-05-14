CREATE TABLE IF NOT EXISTS categories (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL UNIQUE,
    slug     VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_categories_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_categories_name_length CHECK (length(name) <= 255)
);

INSERT INTO categories (name, slug)
VALUES
    ('Hotel', 'hotel'),
    ('Resort', 'resort'),
    ('Cafe', 'cafe'),
    ('Restaurant', 'restaurant'),
    ('Interior', 'interior'),
    ('Architecture', 'architecture'),
    ('Nature', 'nature'),
    ('Beach', 'beach'),
    ('Mountain', 'mountain'),
    ('City', 'city'),
    ('Heritage', 'heritage'),
    ('Street', 'street'),
    ('Food', 'food'),
    ('Travel Ideas', 'travel-ideas')
ON CONFLICT DO NOTHING;

ALTER TABLE posts
    ADD COLUMN IF NOT EXISTS category_id BIGINT,
    ADD CONSTRAINT fk_posts_category_id FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
