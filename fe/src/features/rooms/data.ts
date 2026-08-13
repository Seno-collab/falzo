import type { RoomLanguage, RoomResponse, RoundCardResponse } from "@/types/room";

export type RoomStatus = "waiting" | "playing";
export type PlayerColor = "lime" | "coral" | "blue" | "sand" | "violet" | "mint";

export type RoomPlayer = {
  id: string;
  name: string;
  color: PlayerColor;
  score: number;
  host?: boolean;
  current?: boolean;
  online?: boolean;
  eliminated?: boolean;
};

export type GameRoom = {
  id: string;
  name: string;
  code: string;
  language: RoomLanguage;
  status: RoomStatus;
  round: number;
  maxPlayers: number;
  discussionSeconds: number;
  mrWhiteEnabled: boolean;
  version: number;
  players: RoomPlayer[];
};

export type PlayerCard = {
  playerId: string;
  role: "Civilian" | "Undercover" | "Mr. White";
  word: string;
};

const playerColors: readonly PlayerColor[] = [
  "lime",
  "coral",
  "blue",
  "sand",
  "violet",
  "mint",
];

export function mapRoomResponse(room: RoomResponse): GameRoom {
  return {
    id: room.id,
    name: room.name,
    code: room.invite_code,
    language: room.language_code,
    status: room.status,
    round: Math.max(room.current_round, 1),
    maxPlayers: room.max_players,
    discussionSeconds: room.discussion_seconds,
    mrWhiteEnabled: room.mr_white_enabled,
    version: room.version,
    players: [...room.players]
      .sort((left, right) => left.seat_number - right.seat_number)
      .map((player) => ({
        id: String(player.id),
        name: player.name,
        color: colorForPlayer(player.id),
        score: 0,
        host: player.host,
        current: player.current,
        online: false,
        eliminated: player.eliminated,
      })),
  };
}

export function mapRoundCardResponse(card: RoundCardResponse): PlayerCard {
  return {
    playerId: String(card.player_id),
    role: card.role === "undercover"
      ? "Undercover"
      : card.role === "mr_white" ? "Mr. White" : "Civilian",
    word: card.word,
  };
}

export function colorForPlayer(playerId: number): PlayerColor {
  const index = Math.abs(playerId) % playerColors.length;
  return playerColors[index];
}
