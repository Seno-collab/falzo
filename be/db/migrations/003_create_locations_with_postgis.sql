CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS locations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT             NOT NULL,
    address    TEXT             NOT NULL DEFAULT '',
    latitude   DOUBLE PRECISION NOT NULL,
    longitude  DOUBLE PRECISION NOT NULL,
    geom       GEOGRAPHY(Point, 4326) NOT NULL,
    created_at TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_locations_latitude_range CHECK (latitude BETWEEN -90 AND 90),
    CONSTRAINT chk_locations_longitude_range CHECK (longitude BETWEEN -180 AND 180)
);

CREATE OR REPLACE FUNCTION set_locations_geom()
RETURNS trigger AS $$
BEGIN
    NEW.geom := ST_SetSRID(ST_MakePoint(NEW.longitude, NEW.latitude), 4326)::geography;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_locations_set_geom ON locations;

CREATE TRIGGER trg_locations_set_geom
BEFORE INSERT OR UPDATE OF latitude, longitude ON locations
FOR EACH ROW
EXECUTE FUNCTION set_locations_geom();

UPDATE locations
SET geom = ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography
WHERE geom IS NULL;

CREATE INDEX IF NOT EXISTS idx_locations_geom_gist
    ON locations
    USING GIST (geom);

CREATE INDEX IF NOT EXISTS idx_locations_name_trgm
    ON locations
    USING GIN (name gin_trgm_ops);

DO $$
DECLARE
    location_id_data_type TEXT;
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public'
          AND table_name = 'posts'
    ) THEN
        ALTER TABLE posts
            ADD COLUMN IF NOT EXISTS location_id UUID NULL,
            ADD COLUMN IF NOT EXISTS location_name TEXT NULL,
            ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION NULL,
            ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION NULL;

        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.table_constraints
            WHERE table_schema = 'public'
              AND table_name = 'posts'
              AND constraint_name = 'chk_posts_latitude_range'
        ) THEN
            ALTER TABLE posts
                ADD CONSTRAINT chk_posts_latitude_range
                CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90);
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM information_schema.table_constraints
            WHERE table_schema = 'public'
              AND table_name = 'posts'
              AND constraint_name = 'chk_posts_longitude_range'
        ) THEN
            ALTER TABLE posts
                ADD CONSTRAINT chk_posts_longitude_range
                CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180);
        END IF;

        SELECT c.data_type
        INTO location_id_data_type
        FROM information_schema.columns c
        WHERE c.table_schema = 'public'
          AND c.table_name = 'posts'
          AND c.column_name = 'location_id';

        IF location_id_data_type = 'uuid'
           AND NOT EXISTS (
               SELECT 1
               FROM information_schema.table_constraints
               WHERE table_schema = 'public'
                 AND table_name = 'posts'
                 AND constraint_name = 'fk_posts_location_id'
           )
        THEN
            ALTER TABLE posts
                ADD CONSTRAINT fk_posts_location_id
                FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE SET NULL;
        END IF;

        IF NOT EXISTS (
            SELECT 1
            FROM pg_indexes
            WHERE schemaname = 'public'
              AND tablename = 'posts'
              AND indexname = 'idx_posts_location_id'
        ) THEN
            CREATE INDEX idx_posts_location_id
                ON posts (location_id);
        END IF;
    END IF;
END
$$;
