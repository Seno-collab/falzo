ALTER TABLE rooms
    ADD COLUMN discussion_seconds INT NOT NULL DEFAULT 30,
    ADD COLUMN mr_white_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD CONSTRAINT rooms_discussion_seconds_valid
        CHECK (discussion_seconds BETWEEN 10 AND 30);

ALTER TABLE room_members
    ADD COLUMN eliminated_at TIMESTAMPTZ;

ALTER TABLE room_rounds
    ADD COLUMN mr_white_user_id BIGINT,
    ADD COLUMN phase VARCHAR(32) NOT NULL DEFAULT 'REVEALING_ROLE',
    ADD COLUMN phase_deadline_at TIMESTAMPTZ,
    ADD COLUMN cycle_number INT NOT NULL DEFAULT 1,
    ADD COLUMN voting_completed_at TIMESTAMPTZ,
    ADD COLUMN eliminated_user_id BIGINT,
    ADD COLUMN eliminated_role VARCHAR(20),
    ADD COLUMN winner VARCHAR(20),
    ADD COLUMN mr_white_guess VARCHAR(80),
    ADD COLUMN mr_white_guess_correct BOOLEAN,
    ADD CONSTRAINT room_rounds_mr_white_member_fk
        FOREIGN KEY (room_id, mr_white_user_id)
        REFERENCES room_members(room_id, user_id),
    ADD CONSTRAINT room_rounds_eliminated_member_fk
        FOREIGN KEY (room_id, eliminated_user_id)
        REFERENCES room_members(room_id, user_id),
    ADD CONSTRAINT room_rounds_phase_valid CHECK (phase IN (
        'REVEALING_ROLE', 'DESCRIBING', 'VOTING', 'REVEALING_RESULT',
        'MR_WHITE_GUESSING', 'GAME_FINISHED'
    )),
    ADD CONSTRAINT room_rounds_cycle_positive CHECK (cycle_number > 0),
    ADD CONSTRAINT room_rounds_eliminated_role_valid CHECK (
        eliminated_role IS NULL OR eliminated_role IN ('civilian', 'undercover', 'mr_white')
    ),
    ADD CONSTRAINT room_rounds_winner_valid CHECK (
        winner IS NULL OR winner IN ('civilians', 'undercover', 'mr_white')
    );

CREATE TABLE room_round_ready_players (
    room_id UUID NOT NULL,
    round_number INT NOT NULL,
    user_id BIGINT NOT NULL,
    ready_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, round_number, user_id),
    CONSTRAINT room_round_ready_round_fk
        FOREIGN KEY (room_id, round_number)
        REFERENCES room_rounds(room_id, round_number) ON DELETE CASCADE,
    CONSTRAINT room_round_ready_member_fk
        FOREIGN KEY (room_id, user_id)
        REFERENCES room_members(room_id, user_id) ON DELETE CASCADE
);

CREATE TABLE room_round_turns (
    room_id UUID NOT NULL,
    round_number INT NOT NULL,
    cycle_number INT NOT NULL,
    turn_number INT NOT NULL,
    user_id BIGINT NOT NULL,
    started_at TIMESTAMPTZ,
    deadline_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    skipped BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (room_id, round_number, cycle_number, turn_number),
    CONSTRAINT room_round_turns_player_unique
        UNIQUE (room_id, round_number, cycle_number, user_id),
    CONSTRAINT room_round_turns_round_fk
        FOREIGN KEY (room_id, round_number)
        REFERENCES room_rounds(room_id, round_number) ON DELETE CASCADE,
    CONSTRAINT room_round_turns_member_fk
        FOREIGN KEY (room_id, user_id)
        REFERENCES room_members(room_id, user_id) ON DELETE CASCADE,
    CONSTRAINT room_round_turn_number_positive CHECK (turn_number > 0),
    CONSTRAINT room_round_turn_cycle_positive CHECK (cycle_number > 0)
);

CREATE INDEX room_round_turns_active_idx
    ON room_round_turns (room_id, round_number, cycle_number, turn_number)
    WHERE finished_at IS NULL;

CREATE TABLE room_round_votes (
    room_id UUID NOT NULL,
    round_number INT NOT NULL,
    cycle_number INT NOT NULL,
    voter_user_id BIGINT NOT NULL,
    target_user_id BIGINT NOT NULL,
    voted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, round_number, cycle_number, voter_user_id),
    CONSTRAINT room_round_votes_round_fk
        FOREIGN KEY (room_id, round_number)
        REFERENCES room_rounds(room_id, round_number) ON DELETE CASCADE,
    CONSTRAINT room_round_votes_voter_member_fk
        FOREIGN KEY (room_id, voter_user_id)
        REFERENCES room_members(room_id, user_id) ON DELETE CASCADE,
    CONSTRAINT room_round_votes_target_member_fk
        FOREIGN KEY (room_id, target_user_id)
        REFERENCES room_members(room_id, user_id) ON DELETE CASCADE,
    CONSTRAINT room_round_votes_not_self CHECK (voter_user_id <> target_user_id),
    CONSTRAINT room_round_votes_cycle_positive CHECK (cycle_number > 0)
);

CREATE INDEX room_round_votes_target_idx
    ON room_round_votes (room_id, round_number, cycle_number, target_user_id);
