ALTER TABLE rooms
    ADD COLUMN language_code VARCHAR(10) NOT NULL DEFAULT 'vi',
    ADD CONSTRAINT rooms_language_code_valid
        CHECK (language_code IN ('en', 'vi'));
