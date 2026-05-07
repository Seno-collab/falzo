ALTER TABLE post_comments
    ADD COLUMN IF NOT EXISTS parent_comment_id BIGINT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_post_comments_parent_comment'
    ) THEN
        ALTER TABLE post_comments
            ADD CONSTRAINT fk_post_comments_parent_comment
            FOREIGN KEY (parent_comment_id) REFERENCES post_comments(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_post_comments_parent_comment_id
    ON post_comments (parent_comment_id);
