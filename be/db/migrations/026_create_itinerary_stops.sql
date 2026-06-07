CREATE TABLE IF NOT EXISTS itinerary_stops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    itinerary_id UUID NOT NULL REFERENCES itineraries(id) ON DELETE CASCADE,
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    day_number INT NOT NULL CHECK (day_number >= 1),
    stop_order INT NOT NULL CHECK (stop_order >= 1),
    start_time TIME NULL,
    end_time TIME NULL,
    note TEXT NOT NULL DEFAULT '',
    estimated_cost INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_itinerary_stop_order UNIQUE (itinerary_id, day_number, stop_order)
);

CREATE INDEX IF NOT EXISTS idx_itinerary_stops_itinerary_id
    ON itinerary_stops (itinerary_id);

CREATE INDEX IF NOT EXISTS idx_itinerary_stops_location_id
    ON itinerary_stops (location_id);

CREATE INDEX IF NOT EXISTS idx_itinerary_stops_day_order
    ON itinerary_stops (itinerary_id, day_number, stop_order);
