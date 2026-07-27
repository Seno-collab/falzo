ALTER TABLE rooms
    DROP CONSTRAINT IF EXISTS rooms_language_code_valid,
    DROP COLUMN IF EXISTS language_code;
