"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  ApiError,
  castRoomVote,
  confirmRoomRole,
  finishRoomTurn,
  getCurrentRoundState,
  submitMrWhiteGuess,
} from "@/lib/api";
import { restoreSession } from "@/lib/auth";
import type { RoundStateResponse } from "@/types/room";

type Command = "vote" | "ready" | "finish" | "guess";

export function useRoomGameCommands(
  roomId: string | undefined,
  onState: (state: RoundStateResponse) => void,
  onError: (message: string) => void,
) {
  const router = useRouter();
  const [pending, setPending] = useState<Command | null>(null);

  useEffect(() => setPending(null), [roomId]);

  const execute = useCallback(async (
    command: Command,
    action: (accessToken: string, activeRoomId: string) => Promise<RoundStateResponse>,
    fallback: string,
  ) => {
    if (!roomId || pending !== null) return false;
    setPending(command);
    onError("");
    let accessToken = "";
    try {
      const session = await restoreSession();
      if (!session) {
        router.replace("/login");
        return false;
      }
      accessToken = session.access_token;
      onState(await action(accessToken, roomId));
      return true;
    } catch (error) {
      // A phase deadline can expire between rendering an action and the API
      // receiving it. A conflict in this case means the server already moved
      // the round forward, so refresh instead of showing a stale phase error.
      if (accessToken && error instanceof ApiError && error.status === 409) {
        try {
          onState(await getCurrentRoundState(accessToken, roomId, { trackActivity: false }));
          onError("");
          return false;
        } catch {
          // Keep the original conflict when the recovery request also fails.
        }
      }
      onError(error instanceof Error ? error.message : fallback);
      return false;
    } finally {
      setPending(null);
    }
  }, [onError, onState, pending, roomId, router]);

  return {
    pending,
    confirmRole: () => execute("ready", confirmRoomRole, "Không thể xác nhận thẻ vai trò"),
    finishTurn: () => execute("finish", finishRoomTurn, "Không thể kết thúc lượt"),
    castVote: (targetPlayerId: number) => execute(
      "vote",
      (token, activeRoomId) => castRoomVote(token, activeRoomId, targetPlayerId),
      "Could not submit your vote",
    ),
    submitGuess: (guess: string) => execute(
      "guess",
      (token, activeRoomId) => submitMrWhiteGuess(token, activeRoomId, guess),
      "Không thể gửi đáp án",
    ),
  };
}
