CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_posts_visible_created_id
    ON posts (created_at DESC, id DESC)
    WHERE deleted_at IS NULL AND status = 'visible';

CREATE INDEX IF NOT EXISTS idx_posts_visible_category_created
    ON posts (category_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL AND status = 'visible';

CREATE INDEX IF NOT EXISTS idx_posts_visible_user_created
    ON posts (user_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL AND status = 'visible';

CREATE INDEX IF NOT EXISTS idx_posts_visible_lat_lng
    ON posts (latitude, longitude)
    WHERE deleted_at IS NULL AND status = 'visible' AND latitude IS NOT NULL AND longitude IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_posts_caption_trgm
    ON posts
    USING GIN (caption gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_posts_location_name_trgm
    ON posts
    USING GIN (location_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_categories_name_trgm
    ON categories
    USING GIN (name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_categories_slug
    ON categories (slug);

CREATE INDEX IF NOT EXISTS idx_post_likes_post_user
    ON post_likes (post_id, user_id);

CREATE INDEX IF NOT EXISTS idx_post_saves_post_user
    ON post_saves (post_id, user_id);

CREATE INDEX IF NOT EXISTS idx_user_follows_follower_following
    ON user_follows (follower_id, following_id);

CREATE INDEX IF NOT EXISTS idx_user_follows_following_created
    ON user_follows (following_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_user_blocks_blocker_blocked
    ON user_blocks (blocker_user_id, blocked_user_id);

CREATE INDEX IF NOT EXISTS idx_saved_collections_user_created
    ON saved_collections (user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_saved_collection_posts_collection_created
    ON saved_collection_posts (collection_id, user_id, created_at DESC, id DESC);
