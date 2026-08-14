"use client";

import { useEffect, useState } from "react";
import { restoreSession } from "@/lib/auth";

const reconnectBaseDelay = 800;
const reconnectMaxDelay = 15_000;
const connectionReplacedCode = 4009;

export function useUserRealtime() {
  const [revision, setRevision] = useState(0);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    let active = true;
    let socket: WebSocket | null = null;
    let reconnectTimer: number | null = null;
    let attempt = 0;

    async function connect() {
      const session = await restoreSession();
      if (!active || !session) return;
      const nextSocket = new WebSocket(userWebSocketURL(), [
        "falzo.v1",
        `bearer.${session.access_token}`,
      ]);
      socket = nextSocket;
      nextSocket.onopen = () => {
        if (!active || socket !== nextSocket) return;
        const connectionMode = attempt === 0 ? "initial" : "reconnect";
        attempt = 0;
        setConnected(true);
        nextSocket.send(JSON.stringify({
          type: "notification.sync",
          request_id: createRequestId(),
          payload: { connection_mode: connectionMode },
        }));
      };
      nextSocket.onmessage = ({ data }) => {
        if (!active || socket !== nextSocket || typeof data !== "string") return;
        try {
          const event = JSON.parse(data) as { type?: unknown };
          if (event.type === "social.notifications.updated") {
            setRevision((current) => current + 1);
          }
        } catch {
          // Invalid server frames are ignored; reconnect sync remains authoritative.
        }
      };
      nextSocket.onclose = ({ code }) => {
        if (!active || socket !== nextSocket) return;
        socket = null;
        setConnected(false);
        if (code === connectionReplacedCode) return;
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
      if (socket && socket.readyState < WebSocket.CLOSING) socket.close(1000, "Page closed");
    };
  }, []);

  return { connected, revision };
}

function userWebSocketURL() {
  const configuredURL = process.env.NEXT_PUBLIC_WEBSOCKET_URL?.replace(/\/$/, "");
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const defaultHost = process.env.NODE_ENV === "development"
    ? `${window.location.hostname}:8080`
    : window.location.host;
  return `${configuredURL ?? `${protocol}//${defaultHost}`}/api/v1/ws`;
}

function createRequestId() {
  if (typeof crypto.randomUUID === "function") return crypto.randomUUID();
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}
