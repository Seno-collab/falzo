CREATE TABLE friend_requests (
    id BIGSERIAL PRIMARY KEY,
    sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    responded_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT friend_requests_different_users CHECK (sender_id <> receiver_id),
    CONSTRAINT friend_requests_status_valid CHECK (status IN ('PENDING', 'ACCEPTED', 'REJECTED', 'CANCELED'))
);

CREATE UNIQUE INDEX friend_requests_pending_pair_unique_idx
    ON friend_requests (LEAST(sender_id, receiver_id), GREATEST(sender_id, receiver_id))
    WHERE status = 'PENDING';

CREATE INDEX friend_requests_sender_status_idx
    ON friend_requests (sender_id, status, created_at DESC);

CREATE INDEX friend_requests_receiver_status_idx
    ON friend_requests (receiver_id, status, created_at DESC);

CREATE TABLE friendships (
    user_id_low BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_id_high BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id_low, user_id_high),
    CONSTRAINT friendships_ordered_users CHECK (user_id_low < user_id_high)
);

CREATE INDEX friendships_high_user_idx ON friendships (user_id_high, created_at DESC);

CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(40) NOT NULL,
    reference_id BIGINT NOT NULL REFERENCES friend_requests(id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT notifications_different_users CHECK (user_id <> actor_id),
    CONSTRAINT notifications_type_valid CHECK (type IN ('FRIEND_REQUEST_RECEIVED', 'FRIEND_REQUEST_ACCEPTED'))
);

CREATE INDEX notifications_user_created_idx
    ON notifications (user_id, created_at DESC, id DESC);

CREATE INDEX notifications_user_unread_idx
    ON notifications (user_id, created_at DESC, id DESC)
    WHERE read_at IS NULL;

CREATE INDEX users_username_lower_prefix_idx
    ON users (lower(username) text_pattern_ops);
