INSERT INTO categories (name, slug)
VALUES
    ('Hotel', 'hotel'),
    ('Resort', 'resort'),
    ('Cafe', 'cafe'),
    ('Restaurant', 'restaurant'),
    ('Interior', 'interior'),
    ('Architecture', 'architecture'),
    ('Nature', 'nature'),
    ('Beach', 'beach'),
    ('Mountain', 'mountain'),
    ('City', 'city'),
    ('Heritage', 'heritage'),
    ('Street', 'street'),
    ('Food', 'food'),
    ('Travel Ideas', 'travel-ideas')
ON CONFLICT DO NOTHING;
