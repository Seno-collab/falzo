ALTER TABLE saved_collections
    ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS share_slug TEXT;

UPDATE saved_collections
SET share_slug = lower(regexp_replace(btrim(name), '[^a-zA-Z0-9]+', '-', 'g')) || '-' || id
WHERE share_slug IS NULL OR btrim(share_slug) = '';

ALTER TABLE saved_collections
    ALTER COLUMN share_slug SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_saved_collections_share_slug
    ON saved_collections (share_slug);

CREATE INDEX IF NOT EXISTS idx_saved_collections_public_share_slug
    ON saved_collections (share_slug)
    WHERE is_public = TRUE;
