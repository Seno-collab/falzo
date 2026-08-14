"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import type { ChatMessage } from "@/features/chat/chat-panel";
import { restoreSession } from "@/lib/auth";
import { listRoomMessages } from "@/lib/api";

const reconnectMaxDelay = 15_000;
const reconnectBaseDelay = 800;
const maxVisibleMessages = 200;
const maxRememberedEventIds = 512;
const connectionReplacedCode = 4009;

export type RealtimeStatus = "connecting" | "connected" | "reconnecting" | "offline";

export type RealtimePlayer = {
  id: number;
  name: string;
  seat_number: number;
  host: boolean;
  online: boolean;
  eliminated: boolean;
};

export type RoundStartedEvent = {
  round: number;
  dealt_at: string;
  phase: string;
  phase_deadline_at: string;
};

export type VoteUpdatedEvent = {
  round: number;
  votes_cast: number;
  eligible_voters: number;
  completed: boolean;
};

type PresencePayload = { players: RealtimePlayer[] };
type ChatPayload = {
  id: string;
  user_id: number;
  username: string;
  text: string;
  sent_at: string;
};
type ServerEvent = {
  event_id?: string;
  type: string;
  request_id?: string;
  occurred_at?: string;
  payload?: unknown;
};

export function useRoomRealtime(roomId: string, currentUsername: string) {
  const [status, setStatus] = useState<RealtimeStatus>("connecting");
  const [players, setPlayers] = useState<RealtimePlayer[]>([]);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [roundStarted, setRoundStarted] = useState<RoundStartedEvent | null>(null);
  const [roomRevision, setRoomRevision] = useState(0);
  const [gameStateRevision, setGameStateRevision] = useState(0);
  const [voteUpdated, setVoteUpdated] = useState<VoteUpdatedEvent | null>(null);
  const [error, setError] = useState("");
  const socketRef = useRef<WebSocket | null>(null);
  const seenEventIdsRef = useRef(new Set<string>());
  const eventOrderRef = useRef<string[]>([]);

  useEffect(() => {
    let active = true;
    let reconnectTimer: number | null = null;
    let attempt = 0;
    seenEventIdsRef.current.clear();
    eventOrderRef.current = [];

    async function connect() {
      setStatus(attempt === 0 ? "connecting" : "reconnecting");
      const session = await restoreSession();
      if (!active) return;
      if (!session) {
        setStatus("offline");
        return;
      }

      const socket = new WebSocket(
        roomWebSocketURL(roomId),
        ["falzo.v1", `bearer.${session.access_token}`],
      );
      socketRef.current = socket;

      socket.onopen = () => {
        if (!active || socketRef.current !== socket) return;
        const connectionMode = attempt === 0 ? "initial" : "reconnect";
        attempt = 0;
        setError("");
        setStatus("connected");
        socket.send(JSON.stringify({
          type: "state.sync",
          request_id: createRequestId(),
          payload: { connection_mode: connectionMode },
        }));
        void listRoomMessages(session.access_token, roomId, { trackActivity: false })
          .then((history) => {
            if (!active || socketRef.current !== socket) return;
            setMessages((current) => mergeMessages(
              current,
              history.map((message) => mapChatMessage(message, currentUsername)),
            ));
          })
          .catch(() => {
            if (active && socketRef.current === socket) {
              setError("Could not restore room chat history");
            }
          });
      };

      socket.onmessage = ({ data }) => {
        if (!active || socketRef.current !== socket || typeof data !== "string") return;
        const event = parseServerEvent(data);
        if (!event) return;
        if (event.event_id && isDuplicateServerEvent(
          event.event_id,
          seenEventIdsRef.current,
          eventOrderRef.current,
        )) return;

        switch (event.type) {
          case "presence.snapshot": {
            const payload = event.payload as PresencePayload;
            if (Array.isArray(payload?.players)) setPlayers(payload.players);
            break;
          }
          case "chat.message": {
            const payload = event.payload as ChatPayload;
            if (!payload?.id || typeof payload.text !== "string") return;
            const message = mapChatMessage(payload, currentUsername);
            setMessages((current) => mergeMessages(current, [message]));
            break;
          }
          case "game.round.started": {
            const payload = event.payload as RoundStartedEvent;
            if (typeof payload?.round === "number") setRoundStarted(payload);
            break;
          }
          case "room.updated": {
            setRoomRevision((current) => current + 1);
            break;
          }
          case "game.state.updated": {
            setGameStateRevision((current) => current + 1);
            break;
          }
          case "game.vote.updated": {
            const payload = event.payload as VoteUpdatedEvent;
            if (typeof payload?.round === "number") setVoteUpdated(payload);
            break;
          }
          case "error": {
            const payload = event.payload as { message?: unknown };
            if (typeof payload?.message === "string") setError(payload.message);
            break;
          }
        }
      };

      socket.onerror = () => {
        if (active && socketRef.current === socket) setError("Realtime connection interrupted");
      };

      socket.onclose = ({ code }) => {
        if (!active || socketRef.current !== socket) return;
        socketRef.current = null;
        if (code === connectionReplacedCode) {
          setError("Phòng này đã được mở bằng một kết nối mới hơn.");
          setStatus("offline");
          return;
        }
        setStatus("reconnecting");
        const delay = Math.min(reconnectBaseDelay * 2 ** attempt, reconnectMaxDelay)
          + Math.floor(Math.random() * 250);
        attempt += 1;
        reconnectTimer = window.setTimeout(() => void connect(), delay);
      };
    }

    void connect();
    return () => {
      active = false;
      if (reconnectTimer !== null) window.clearTimeout(reconnectTimer);
      const socket = socketRef.current;
      socketRef.current = null;
      if (socket && socket.readyState < WebSocket.CLOSING) socket.close(1000, "Room closed");
    };
  }, [currentUsername, roomId]);

  const sendChat = useCallback((text: string) => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) return false;
    socket.send(JSON.stringify({
      type: "chat.send",
      request_id: createRequestId(),
      payload: { text },
    }));
    return true;
  }, []);

  return {
    connected: status === "connected",
    error,
    messages,
    gameStateRevision,
    players,
    roundStarted,
    roomRevision,
    sendChat,
    status,
    voteUpdated,
  };
}

function roomWebSocketURL(roomId: string) {
  const configuredURL = process.env.NEXT_PUBLIC_WEBSOCKET_URL?.replace(/\/$/, "");
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const defaultHost = process.env.NODE_ENV === "development"
    ? `${window.location.hostname}:8080`
    : window.location.host;
  const baseURL = configuredURL ?? `${protocol}//${defaultHost}`;
  return `${baseURL}/api/v1/rooms/${encodeURIComponent(roomId)}/ws`;
}

function parseServerEvent(raw: string): ServerEvent | null {
  try {
    const event = JSON.parse(raw) as Partial<ServerEvent>;
    return typeof event.type === "string" ? {
      event_id: typeof event.event_id === "string" ? event.event_id : undefined,
      type: event.type,
      request_id: typeof event.request_id === "string" ? event.request_id : undefined,
      occurred_at: typeof event.occurred_at === "string" ? event.occurred_at : undefined,
      payload: event.payload,
    } : null;
  } catch {
    return null;
  }
}

function createRequestId() {
  if (typeof crypto.randomUUID === "function") return crypto.randomUUID();
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function isDuplicateServerEvent(eventId: string, seen: Set<string>, order: string[]) {
  if (seen.has(eventId)) return true;
  seen.add(eventId);
  order.push(eventId);
  if (order.length > maxRememberedEventIds) {
    const oldest = order.shift();
    if (oldest) seen.delete(oldest);
  }
  return false;
}

function formatMessageTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Now";
  return new Intl.DateTimeFormat("en", { hour: "2-digit", minute: "2-digit" }).format(date);
}

function mapChatMessage(payload: ChatPayload, currentUsername: string): ChatMessage {
  return {
    id: payload.id,
    sender: payload.username,
    text: payload.text,
    time: formatMessageTime(payload.sent_at),
    sentAt: payload.sent_at,
    own: payload.username === currentUsername,
  };
}

function mergeMessages(current: ChatMessage[], incoming: ChatMessage[]) {
  const messages = new Map(current.map((message) => [message.id, message]));
  for (const message of incoming) messages.set(message.id, message);
  return [...messages.values()]
    .sort((left, right) => (left.sentAt ?? "").localeCompare(right.sentAt ?? ""))
    .slice(-maxVisibleMessages);
}
