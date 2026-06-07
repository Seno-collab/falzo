CREATE TABLE IF NOT EXISTS place_details (
    location_id UUID PRIMARY KEY REFERENCES locations(id) ON DELETE CASCADE,
    slug TEXT UNIQUE NOT NULL,
    province TEXT NOT NULL DEFAULT '',
    district TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    best_time_to_visit TEXT NOT NULL DEFAULT '',
    estimated_cost_min INT NOT NULL DEFAULT 0,
    estimated_cost_max INT NOT NULL DEFAULT 0,
    travel_styles TEXT[] NOT NULL DEFAULT '{}',
    suitable_for TEXT[] NOT NULL DEFAULT '{}',
    warning_note TEXT NOT NULL DEFAULT '',
    is_hidden_gem BOOLEAN NOT NULL DEFAULT FALSE,
    rating_reality SMALLINT NULL CHECK (rating_reality BETWEEN 1 AND 10),
    rating_photo SMALLINT NULL CHECK (rating_photo BETWEEN 1 AND 10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_place_details_slug
    ON place_details (slug);

CREATE INDEX IF NOT EXISTS idx_place_details_province
    ON place_details (province);

CREATE INDEX IF NOT EXISTS idx_place_details_is_hidden_gem
    ON place_details (is_hidden_gem);
