ALTER TABLE posts
    ADD COLUMN IF NOT EXISTS image_urls JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE posts
SET image_urls = jsonb_build_array(image_url)
WHERE image_urls = '[]'::jsonb
  AND image_url IS NOT NULL
  AND image_url <> '';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_posts_image_urls_array'
    ) THEN
        ALTER TABLE posts
            ADD CONSTRAINT chk_posts_image_urls_array CHECK (jsonb_typeof(image_urls) = 'array');
    END IF;
END $$;
