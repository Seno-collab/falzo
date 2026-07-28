"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import type { ChatMessage } from "@/features/chat/chat-panel";
import { restoreSession } from "@/lib/auth";

const reconnectMaxDelay = 15_000;
const reconnectBaseDelay = 800;
const maxVisibleMessages = 200;

export type RealtimeStatus = "connecting" | "connected" | "reconnecting" | "offline";

export type RealtimePlayer = {
  id: number;
  name: string;
  seat_number: number;
  host: boolean;
  online: boolean;
};

export type RoundStartedEvent = {
  round: number;
  dealt_at: string;
};

type PresencePayload = { players: RealtimePlayer[] };
type ChatPayload = {
  id: string;
  user_id: number;
  username: string;
  text: string;
  sent_at: string;
};
type ServerEvent = { type: string; payload?: unknown };

export function useRoomRealtime(roomId: string, currentUsername: string) {
  const [status, setStatus] = useState<RealtimeStatus>("connecting");
  const [players, setPlayers] = useState<RealtimePlayer[]>([]);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [roundStarted, setRoundStarted] = useState<RoundStartedEvent | null>(null);
  const [error, setError] = useState("");
  const socketRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let active = true;
    let reconnectTimer: number | null = null;
    let attempt = 0;

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
        attempt = 0;
        setError("");
        setStatus("connected");
      };

      socket.onmessage = ({ data }) => {
        if (!active || typeof data !== "string") return;
        const event = parseServerEvent(data);
        if (!event) return;

        switch (event.type) {
          case "presence.snapshot": {
            const payload = event.payload as PresencePayload;
            if (Array.isArray(payload?.players)) setPlayers(payload.players);
            break;
          }
          case "chat.message": {
            const payload = event.payload as ChatPayload;
            if (!payload?.id || typeof payload.text !== "string") return;
            const message: ChatMessage = {
              id: payload.id,
              sender: payload.username,
              text: payload.text,
              time: formatMessageTime(payload.sent_at),
              own: payload.username === currentUsername,
            };
            setMessages((current) => current.some(({ id }) => id === message.id)
              ? current
              : [...current, message].slice(-maxVisibleMessages));
            break;
          }
          case "game.round.started": {
            const payload = event.payload as RoundStartedEvent;
            if (typeof payload?.round === "number") setRoundStarted(payload);
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

      socket.onclose = () => {
        if (!active || socketRef.current !== socket) return;
        socketRef.current = null;
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
    socket.send(JSON.stringify({ type: "chat.send", payload: { text } }));
    return true;
  }, []);

  return {
    connected: status === "connected",
    error,
    messages,
    players,
    roundStarted,
    sendChat,
    status,
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
    return typeof event.type === "string" ? { type: event.type, payload: event.payload } : null;
  } catch {
    return null;
  }
}

function formatMessageTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Now";
  return new Intl.DateTimeFormat("en", { hour: "2-digit", minute: "2-digit" }).format(date);
}
