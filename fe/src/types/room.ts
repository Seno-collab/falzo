export type RoomStatus = "waiting" | "playing";
export type RoomLanguage = "en" | "vi";

export type RoomMemberResponse = {
  id: number;
  name: string;
  seat_number: number;
  host: boolean;
  current: boolean;
  joined_at: string;
  eliminated: boolean;
};

export type RoomResponse = {
  id: string;
  invite_code: string;
  name: string;
  language_code: RoomLanguage;
  status: RoomStatus;
  max_players: number;
  current_round: number;
  discussion_seconds: number;
  mr_white_enabled: boolean;
  player_count: number;
  version: number;
  expires_at: string;
  created_at: string;
  players: RoomMemberResponse[];
};

export type RoundCardResponse = {
  room_id: string;
  round: number;
  player_id: number;
  role: "civilian" | "undercover" | "mr_white";
  word: string;
  dealt_at: string;
  phase: RoundPhase;
  phase_deadline_at: string;
};

export type RoundPhase =
  | "REVEALING_ROLE"
  | "DESCRIBING"
  | "VOTING"
  | "REVEALING_RESULT"
  | "MR_WHITE_GUESSING"
  | "GAME_FINISHED";

export type RoundStateResponse = {
  room_id: string;
  round: number;
  cycle: number;
  phase: RoundPhase;
  phase_deadline_at: string | null;
  ready_players: number;
  eligible_players: number;
  current_user_ready: boolean;
  current_turn_player_id: number | null;
  turn_number: number;
  total_turns: number;
  turn_ends_at: string | null;
  eligible_voters: number;
  votes_cast: number;
  current_user_vote_id: number | null;
  undercover_player_id: number | null;
  mr_white_player_id: number | null;
  eliminated_player_id: number | null;
  eliminated_role: "civilian" | "undercover" | "mr_white" | null;
  winner: "civilians" | "undercover" | "mr_white" | null;
  mr_white_guess_correct: boolean | null;
};
