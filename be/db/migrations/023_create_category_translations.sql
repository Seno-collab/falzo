CREATE TABLE IF NOT EXISTS category_translations (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    locale TEXT NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_category_translations_locale CHECK (locale IN ('en', 'vi')),
    CONSTRAINT chk_category_translations_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_category_translations_name_length CHECK (length(name) <= 255),
    CONSTRAINT uq_category_translations_category_locale UNIQUE (category_id, locale)
);

CREATE INDEX IF NOT EXISTS idx_category_translations_locale
    ON category_translations (locale);

CREATE OR REPLACE FUNCTION set_category_translations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_category_translations_updated_at ON category_translations;
CREATE TRIGGER trg_category_translations_updated_at
BEFORE UPDATE ON category_translations
FOR EACH ROW
EXECUTE FUNCTION set_category_translations_updated_at();

INSERT INTO category_translations (category_id, locale, name)
SELECT id, 'en', name
FROM categories
ON CONFLICT (category_id, locale) DO UPDATE
SET name = EXCLUDED.name;

INSERT INTO category_translations (category_id, locale, name)
SELECT categories.id, translations.locale, translations.name
FROM categories
INNER JOIN (
    VALUES
        ('hotel', 'vi', 'Khách sạn'),
        ('resort', 'vi', 'Khu nghỉ dưỡng'),
        ('cafe', 'vi', 'Quán cà phê'),
        ('restaurant', 'vi', 'Nhà hàng'),
        ('interior', 'vi', 'Nội thất'),
        ('architecture', 'vi', 'Kiến trúc'),
        ('nature', 'vi', 'Thiên nhiên'),
        ('beach', 'vi', 'Bãi biển'),
        ('mountain', 'vi', 'Núi'),
        ('city', 'vi', 'Thành phố'),
        ('heritage', 'vi', 'Di sản'),
        ('street', 'vi', 'Đường phố'),
        ('food', 'vi', 'Ẩm thực'),
        ('travel-ideas', 'vi', 'Ý tưởng du lịch')
) AS translations(slug, locale, name) ON translations.slug = categories.slug
ON CONFLICT (category_id, locale) DO UPDATE
SET name = EXCLUDED.name;
