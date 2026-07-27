export type RoomStatus = "waiting" | "playing";
export type RoomLanguage = "en" | "vi";

export type RoomMemberResponse = {
  id: number;
  name: string;
  seat_number: number;
  host: boolean;
  current: boolean;
  joined_at: string;
};

export type RoomResponse = {
  id: string;
  invite_code: string;
  name: string;
  language_code: RoomLanguage;
  status: RoomStatus;
  max_players: number;
  current_round: number;
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
  role: "civilian" | "undercover";
  word: string;
  dealt_at: string;
};
