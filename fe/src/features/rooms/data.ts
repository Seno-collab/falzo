export type RoomStatus = "waiting" | "playing";

export type RoomPlayer = {
  id: string;
  name: string;
  color: "lime" | "coral" | "blue" | "sand" | "violet" | "mint";
  score: number;
  host?: boolean;
  current?: boolean;
};

export type GameRoom = {
  id: string;
  name: string;
  code: string;
  status: RoomStatus;
  round: number;
  maxPlayers: number;
  players: RoomPlayer[];
};

export type PlayerCard = {
  playerId: string;
  role: "Civilian" | "Undercover";
  word: string;
};

const currentPlayer: RoomPlayer = {
  id: "current-player",
  name: "You",
  color: "lime",
  score: 0,
  current: true,
};

export const rooms: readonly GameRoom[] = [
  {
    id: "night-owls",
    name: "Night Owls",
    code: "OWL824",
    status: "waiting",
    round: 1,
    maxPlayers: 8,
    players: [
      { id: "minh", name: "Minh", color: "coral", score: 18, host: true },
      { ...currentPlayer, score: 12 },
      { id: "lan", name: "Lan", color: "blue", score: 10 },
      { id: "bao", name: "Bảo", color: "sand", score: 7 },
      { id: "vy", name: "Vy", color: "violet", score: 4 },
    ],
  },
  {
    id: "sunday-table",
    name: "Sunday Table",
    code: "SUN417",
    status: "playing",
    round: 2,
    maxPlayers: 8,
    players: [
      { id: "an", name: "An", color: "blue", score: 32, host: true },
      { id: "phuong", name: "Phương", color: "coral", score: 26 },
      { ...currentPlayer, score: 22 },
      { id: "khoa", name: "Khoa", color: "mint", score: 18 },
      { id: "tram", name: "Trâm", color: "violet", score: 14 },
      { id: "duy", name: "Duy", color: "sand", score: 8 },
    ],
  },
  {
    id: "after-work",
    name: "After Work",
    code: "WORK19",
    status: "waiting",
    round: 1,
    maxPlayers: 6,
    players: [
      { id: "ha", name: "Hà", color: "mint", score: 8 },
      { id: "nam", name: "Nam", color: "sand", score: 6 },
      { ...currentPlayer, score: 4, host: true },
    ],
  },
  {
    id: "quiet-chaos",
    name: "Quiet Chaos",
    code: "QC2048",
    status: "waiting",
    round: 1,
    maxPlayers: 10,
    players: [
      { id: "linh", name: "Linh", color: "violet", score: 21, host: true },
      { id: "son", name: "Sơn", color: "coral", score: 17 },
      { ...currentPlayer, score: 15 },
      { id: "mai", name: "Mai", color: "blue", score: 11 },
      { id: "tuan", name: "Tuấn", color: "sand", score: 6 },
    ],
  },
  {
    id: "late-checkout",
    name: "Late Checkout",
    code: "LATE88",
    status: "playing",
    round: 3,
    maxPlayers: 8,
    players: [
      { id: "nhi", name: "Nhi", color: "coral", score: 46, host: true },
      { ...currentPlayer, score: 39 },
      { id: "quan", name: "Quân", color: "blue", score: 35 },
      { id: "thu", name: "Thu", color: "mint", score: 29 },
      { id: "dat", name: "Đạt", color: "sand", score: 21 },
      { id: "yen", name: "Yến", color: "violet", score: 16 },
      { id: "long", name: "Long", color: "coral", score: 9 },
    ],
  },
] as const;

const wordPairs = [
  { civilian: "Forest", undercover: "Jungle" },
  { civilian: "Coffee", undercover: "Tea" },
  { civilian: "Moon", undercover: "Sun" },
  { civilian: "Beach", undercover: "Island" },
  { civilian: "Train", undercover: "Bus" },
] as const;

export function getRoomsForPlayer(username: string): GameRoom[] {
  return rooms.map((room) => ({
    ...room,
    players: room.players.map((player) => (
      player.current ? { ...player, name: username } : { ...player }
    )),
  }));
}

export function getRoomForPlayer(roomId: string, username: string): GameRoom | undefined {
  return getRoomsForPlayer(username).find((room) => room.id === roomId);
}

export function dealCards(players: RoomPlayer[]): PlayerCard[] {
  if (players.length === 0) {
    return [];
  }

  const pair = wordPairs[Math.floor(Math.random() * wordPairs.length)];
  const undercoverIndex = Math.floor(Math.random() * players.length);

  return players.map((player, index) => ({
    playerId: player.id,
    role: index === undercoverIndex ? "Undercover" : "Civilian",
    word: index === undercoverIndex ? pair.undercover : pair.civilian,
  }));
}
