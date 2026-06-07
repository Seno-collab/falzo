CREATE TABLE IF NOT EXISTS itineraries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    province TEXT NOT NULL,
    duration_days INT NOT NULL CHECK (duration_days BETWEEN 1 AND 14),
    budget_min INT NOT NULL DEFAULT 0,
    budget_max INT NOT NULL DEFAULT 0,
    travel_style TEXT NOT NULL DEFAULT '',
    transportation TEXT NOT NULL DEFAULT '',
    start_location TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    cover_image_url TEXT NOT NULL DEFAULT '',
    created_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_itineraries_slug
    ON itineraries (slug);

CREATE INDEX IF NOT EXISTS idx_itineraries_province
    ON itineraries (province);

CREATE INDEX IF NOT EXISTS idx_itineraries_duration_days
    ON itineraries (duration_days);

CREATE INDEX IF NOT EXISTS idx_itineraries_is_public
    ON itineraries (is_public);

CREATE INDEX IF NOT EXISTS idx_itineraries_created_at
    ON itineraries (created_at DESC);
