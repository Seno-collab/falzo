"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError, getCurrentRoundState } from "@/lib/api";
import { restoreSession } from "@/lib/auth";
import type { RoundStateResponse } from "@/types/room";
import type { GameStateUpdatedEvent } from "./use-room-realtime";

type Options = {
  roomId?: string;
  playing: boolean;
  roomRound?: number;
  realtimeRevision: number;
  realtimeState?: GameStateUpdatedEvent | null;
  roundStarted?: number;
  votesCast?: number;
};

export function useRoomGameState({
  roomId,
  playing,
  roomRound,
  realtimeRevision,
  realtimeState,
  roundStarted,
  votesCast,
}: Options) {
  const router = useRouter();
  const [state, setState] = useState<RoundStateResponse | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!playing || !realtimeState) return;
    setState((current) => {
      if (!current || current.round !== realtimeState.round) return current;
      const enteringVoting = realtimeState.phase === "VOTING"
        && (current.phase !== "VOTING" || current.cycle !== realtimeState.cycle);
      return {
        ...current,
        cycle: realtimeState.cycle,
        phase: realtimeState.phase,
        phase_deadline_at: realtimeState.phase_deadline_at,
        current_turn_player_id: realtimeState.current_turn_player_id,
        ...(enteringVoting ? {
          votes_cast: 0,
          current_user_vote_id: null,
          eliminated_player_id: null,
          eliminated_role: null,
        } : {}),
      };
    });
  }, [playing, realtimeState]);

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
