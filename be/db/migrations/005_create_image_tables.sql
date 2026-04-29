CREATE TABLE images (
    id SERIAL PRIMARY KEY,
    owner_id UUID,
    object_key TEXT,
    url TEXT,
    mime_type TEXT,
    size BIGINT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
);