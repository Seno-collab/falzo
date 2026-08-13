DROP TABLE IF EXISTS room_round_votes;
DROP TABLE IF EXISTS room_round_turns;
DROP TABLE IF EXISTS room_round_ready_players;

ALTER TABLE room_rounds
    DROP CONSTRAINT IF EXISTS room_rounds_winner_valid,
    DROP CONSTRAINT IF EXISTS room_rounds_eliminated_role_valid,
    DROP CONSTRAINT IF EXISTS room_rounds_cycle_positive,
    DROP CONSTRAINT IF EXISTS room_rounds_phase_valid,
    DROP CONSTRAINT IF EXISTS room_rounds_eliminated_member_fk,
    DROP CONSTRAINT IF EXISTS room_rounds_mr_white_member_fk,
    DROP COLUMN IF EXISTS mr_white_guess_correct,
    DROP COLUMN IF EXISTS mr_white_guess,
    DROP COLUMN IF EXISTS winner,
    DROP COLUMN IF EXISTS eliminated_role,
    DROP COLUMN IF EXISTS eliminated_user_id,
    DROP COLUMN IF EXISTS voting_completed_at,
    DROP COLUMN IF EXISTS cycle_number,
    DROP COLUMN IF EXISTS phase_deadline_at,
    DROP COLUMN IF EXISTS phase,
    DROP COLUMN IF EXISTS mr_white_user_id;

ALTER TABLE rooms
    DROP CONSTRAINT IF EXISTS rooms_discussion_seconds_valid,
    DROP COLUMN IF EXISTS mr_white_enabled,
    DROP COLUMN IF EXISTS discussion_seconds;

ALTER TABLE room_members
    DROP COLUMN IF EXISTS eliminated_at;
