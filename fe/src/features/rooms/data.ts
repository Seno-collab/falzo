export type RoomStatus = "waiting" | "playing";

export type RoomPlayer = {
  id: string;
  name: string;
  color: "lime" | "coral" | "blue" | "sand" | "violet" | "mint";
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
      { id: "minh", name: "Minh", color: "coral", host: true },
      currentPlayer,
      { id: "lan", name: "Lan", color: "blue" },
      { id: "bao", name: "Bảo", color: "sand" },
      { id: "vy", name: "Vy", color: "violet" },
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
      { id: "an", name: "An", color: "blue", host: true },
      { id: "phuong", name: "Phương", color: "coral" },
      currentPlayer,
      { id: "khoa", name: "Khoa", color: "mint" },
      { id: "tram", name: "Trâm", color: "violet" },
      { id: "duy", name: "Duy", color: "sand" },
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
      { id: "ha", name: "Hà", color: "mint", host: true },
      { id: "nam", name: "Nam", color: "sand" },
      currentPlayer,
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
      { id: "linh", name: "Linh", color: "violet", host: true },
      { id: "son", name: "Sơn", color: "coral" },
      currentPlayer,
      { id: "mai", name: "Mai", color: "blue" },
      { id: "tuan", name: "Tuấn", color: "sand" },
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
      { id: "nhi", name: "Nhi", color: "coral", host: true },
      currentPlayer,
      { id: "quan", name: "Quân", color: "blue" },
      { id: "thu", name: "Thu", color: "mint" },
      { id: "dat", name: "Đạt", color: "sand" },
      { id: "yen", name: "Yến", color: "violet" },
      { id: "long", name: "Long", color: "coral" },
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
