CREATE TABLE rooms (
    id UUID PRIMARY KEY,
    invite_code VARCHAR(8) NOT NULL,
    name VARCHAR(80) NOT NULL,
    host_user_id BIGINT NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'waiting',
    max_players INT NOT NULL,
    current_round INT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT rooms_invite_code_unique UNIQUE (invite_code),
    CONSTRAINT rooms_status_valid CHECK (status IN ('waiting', 'playing', 'closed')),
    CONSTRAINT rooms_max_players_valid CHECK (max_players BETWEEN 4 AND 12),
    CONSTRAINT rooms_round_valid CHECK (current_round >= 0),
    CONSTRAINT rooms_version_valid CHECK (version > 0)
);

CREATE INDEX rooms_active_created_at_idx
    ON rooms (created_at DESC)
    WHERE status <> 'closed';

CREATE TABLE room_members (
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seat_number INT NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, user_id),
    CONSTRAINT room_members_seat_unique UNIQUE (room_id, seat_number),
    CONSTRAINT room_members_seat_positive CHECK (seat_number > 0)
);

CREATE INDEX room_members_user_id_idx ON room_members (user_id);
