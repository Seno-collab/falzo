CREATE TABLE room_rounds (
    room_id UUID NOT NULL,
    round_number INT NOT NULL,
    common_word VARCHAR(80) NOT NULL,
    different_word VARCHAR(80) NOT NULL,
    undercover_user_id BIGINT NOT NULL,
    dealt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, round_number),
    CONSTRAINT room_rounds_room_fk
        FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    CONSTRAINT room_rounds_undercover_member_fk
        FOREIGN KEY (room_id, undercover_user_id)
        REFERENCES room_members(room_id, user_id) ON DELETE CASCADE,
    CONSTRAINT room_rounds_number_positive CHECK (round_number > 0),
    CONSTRAINT room_rounds_common_word_not_blank CHECK (length(trim(common_word)) > 0),
    CONSTRAINT room_rounds_different_word_not_blank CHECK (length(trim(different_word)) > 0),
    CONSTRAINT room_rounds_words_different CHECK (lower(common_word) <> lower(different_word))
);

CREATE INDEX room_rounds_undercover_user_idx
    ON room_rounds (undercover_user_id);
