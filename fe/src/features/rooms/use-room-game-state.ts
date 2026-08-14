"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError, getCurrentRoundState } from "@/lib/api";
import { restoreSession } from "@/lib/auth";
import type { RoundStateResponse } from "@/types/room";

type Options = {
  roomId?: string;
  playing: boolean;
  roomRound?: number;
  realtimeRevision: number;
  roundStarted?: number;
  votesCast?: number;
};

export function useRoomGameState({
  roomId,
  playing,
  roomRound,
  realtimeRevision,
  roundStarted,
  votesCast,
}: Options) {
  const router = useRouter();
  const [state, setState] = useState<RoundStateResponse | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!roomId || !playing) {
      setState(null);
      return;
    }
    const activeRoomId = roomId;
    let active = true;
    async function load() {
      try {
        const session = await restoreSession();
        if (!session) {
          router.replace("/login");
          return;
        }
        const response = await getCurrentRoundState(session.access_token, activeRoomId, {
          trackActivity: false,
        });
        if (!active) return;
        setState(response);
        setError("");
      } catch (loadError) {
        if (!active || (loadError instanceof ApiError && loadError.status === 404)) return;
        setError(loadError instanceof Error ? loadError.message : "Could not load game state");
      }
    }
    void load();
    return () => {
      active = false;
    };
  }, [playing, realtimeRevision, roomId, roomRound, roundStarted, router, votesCast]);

  return { state, setState, error, setError };
}
