"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  castRoomVote,
  confirmRoomRole,
  finishRoomTurn,
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
    try {
      const session = await restoreSession();
      if (!session) {
        router.replace("/login");
        return false;
      }
      onState(await action(session.access_token, roomId));
      return true;
    } catch (error) {
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
