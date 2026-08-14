"use client";

import type { FormEvent } from "react";
import type { GameRoom, RoomPlayer } from "@/features/rooms/data";
import type { RoundStateResponse } from "@/types/room";
import styles from "./room.module.css";

type RoomPhasePanelProps = {
  room: GameRoom;
  state: RoundStateResponse | null;
  stateError: string;
  remainingSeconds: number;
  currentCardAvailable: boolean;
  selectedVote: string;
  votePending: boolean;
  turnPending: boolean;
  mrWhiteGuess: string;
  mrWhiteGuessPending: boolean;
  onRevealCard: () => void;
  onSelectedVoteChange: (playerID: string) => void;
  onSubmitVote: () => void | Promise<void>;
  onFinishTurn: () => void | Promise<void>;
  onMrWhiteGuessChange: (guess: string) => void;
  onSubmitWhiteGuess: () => void | Promise<void>;
};

export function RoomPhasePanel({
  room,
  state,
  stateError,
  remainingSeconds,
  currentCardAvailable,
  selectedVote,
  votePending,
  turnPending,
  mrWhiteGuess,
  mrWhiteGuessPending,
  onRevealCard,
  onSelectedVoteChange,
  onSubmitVote,
  onFinishTurn,
  onMrWhiteGuessChange,
  onSubmitWhiteGuess,
}: RoomPhasePanelProps) {
  const currentPlayer = room.players.find((player) => player.current);
  const currentTurnPlayer = state?.current_turn_player_id
    ? room.players.find((player) => player.id === String(state.current_turn_player_id))
    : undefined;
  const isCurrentTurn = Boolean(currentPlayer && currentTurnPlayer?.id === currentPlayer.id);
  const activeVoteTargets = room.players.filter(
    (player) => !player.current && !player.eliminated,
  );

  return (
    <section className={styles.votePanel} aria-labelledby="round-flow-title">
      <div className={styles.votePanelHeading}>
        <div>
          <small>VÁN {room.round} · VÒNG {state?.cycle ?? 1}</small>
          <h2 id="round-flow-title">{phaseTitle(state?.phase)}</h2>
        </div>
        {state?.phase_deadline_at && state.phase !== "GAME_FINISHED" && (
          <strong className={styles.roundTimer}>{formatCountdown(remainingSeconds)}</strong>
        )}
      </div>

      {!state && <p className={styles.voteHelp}>Đang đồng bộ trạng thái ván…</p>}

      {state?.phase === "REVEALING_ROLE" && (
        <RoleRevealPhase
          currentCardAvailable={currentCardAvailable}
          onRevealCard={onRevealCard}
          state={state}
        />
      )}
      {state?.phase === "DESCRIBING" && (
        <DescribingPhase
          currentTurnPlayer={currentTurnPlayer}
          isCurrentTurn={isCurrentTurn}
          onFinishTurn={onFinishTurn}
          state={state}
          turnPending={turnPending}
        />
      )}
      {state?.phase === "VOTING" && (
        <VotingPhase
          currentPlayer={currentPlayer}
          onSelectedVoteChange={onSelectedVoteChange}
          onSubmitVote={onSubmitVote}
          selectedVote={selectedVote}
          state={state}
          targets={activeVoteTargets}
          votePending={votePending}
        />
      )}
      {state?.phase === "REVEALING_RESULT" && (
        <RevealingResultPhase room={room} state={state} />
      )}
      {state?.phase === "MR_WHITE_GUESSING" && (
        <MrWhiteGuessingPhase
          currentPlayer={currentPlayer}
          guess={mrWhiteGuess}
          onGuessChange={onMrWhiteGuessChange}
          onSubmit={onSubmitWhiteGuess}
          pending={mrWhiteGuessPending}
          state={state}
        />
      )}
      {state?.phase === "GAME_FINISHED" && <GameFinishedPhase room={room} state={state} />}

      {stateError && <small className={styles.dealError}>{stateError}</small>}
    </section>
  );
}

function RoleRevealPhase({
  state,
  currentCardAvailable,
  onRevealCard,
}: {
  state: RoundStateResponse;
  currentCardAvailable: boolean;
  onRevealCard: () => void;
}) {
  return (
    <div className={styles.voteWaiting}>
      <strong>{state.ready_players}/{state.eligible_players} người đã hiểu vai trò</strong>
      <p>{state.current_user_ready
        ? "Đang chờ những người còn lại."
        : "Mở thẻ bí mật và bấm “Đã hiểu” để sẵn sàng."}</p>
      {!state.current_user_ready && currentCardAvailable && (
        <button onClick={onRevealCard} type="button">Xem thẻ của tôi</button>
      )}
    </div>
  );
}

function DescribingPhase({
  state,
  currentTurnPlayer,
  isCurrentTurn,
  turnPending,
  onFinishTurn,
}: {
  state: RoundStateResponse;
  currentTurnPlayer?: RoomPlayer;
  isCurrentTurn: boolean;
  turnPending: boolean;
  onFinishTurn: () => void | Promise<void>;
}) {
  return (
    <div className={styles.voteWaiting}>
      <strong>
        Lượt {state.turn_number}/{state.total_turns}: {currentTurnPlayer?.name ?? "Đang chuyển lượt"}
      </strong>
      <p>{isCurrentTurn
        ? "Hãy mô tả từ của bạn nhưng không nói thẳng từ khóa."
        : `Đang nghe ${currentTurnPlayer?.name ?? "người chơi"} mô tả.`}</p>
      {isCurrentTurn && (
        <button disabled={turnPending} onClick={onFinishTurn} type="button">
          {turnPending ? "Đang chuyển lượt…" : "Kết thúc lượt sớm"}
        </button>
      )}
    </div>
  );
}

function VotingPhase({
  state,
  currentPlayer,
  targets,
  selectedVote,
  votePending,
  onSelectedVoteChange,
  onSubmitVote,
}: {
  state: RoundStateResponse;
  currentPlayer?: RoomPlayer;
  targets: RoomPlayer[];
  selectedVote: string;
  votePending: boolean;
  onSelectedVoteChange: (playerID: string) => void;
  onSubmitVote: () => void | Promise<void>;
}) {
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void onSubmitVote();
  }

  return (
    <>
      {state.current_user_vote_id === null && !currentPlayer?.eliminated && (
        <form className={styles.voteForm} onSubmit={submit}>
          <p>Chọn người đáng nghi. Nếu không chọn, lượt vote sẽ được bỏ qua khi hết giờ.</p>
          <div className={styles.voteChoices}>
            {targets.map((player) => (
              <label key={player.id}>
                <input
                  checked={selectedVote === player.id}
                  name="vote-target"
                  onChange={() => onSelectedVoteChange(player.id)}
                  type="radio"
                  value={player.id}
                />
                <span className={`${styles.voteAvatar} ${styles[player.color]}`} aria-hidden="true">
                  {player.name.charAt(0).toUpperCase()}
                </span>
                <strong>{player.name}</strong>
              </label>
            ))}
          </div>
          <button disabled={!selectedVote || votePending} type="submit">
            {votePending ? "Đang gửi…" : "Chốt phiếu"}
          </button>
        </form>
      )}

      {currentPlayer?.eliminated && (
        <div className={styles.voteWaiting}>
          <strong>Spectator mode</strong>
          <p>You can follow the vote, but eliminated players cannot vote or chat.</p>
        </div>
      )}

      {state.current_user_vote_id !== null && (
        <div className={styles.voteWaiting}>
          <strong>Vote locked</strong>
          <p>Waiting for the remaining players.</p>
        </div>
      )}

      <div className={styles.voteProgress}>
        <span>{state.votes_cast}/{state.eligible_voters} votes</span>
        <div aria-hidden="true">
          <span style={{ width: `${Math.min(100, state.votes_cast / Math.max(1, state.eligible_voters) * 100)}%` }} />
        </div>
      </div>
    </>
  );
}

function RevealingResultPhase({ room, state }: { room: GameRoom; state: RoundStateResponse }) {
  const eliminatedPlayer = state.eliminated_player_id
    ? room.players.find((player) => player.id === String(state.eliminated_player_id))
    : undefined;

  return (
    <div className={`${styles.voteResult} ${state.eliminated_role === "undercover" ? styles.voteCaught : ""}`}>
      <small>{state.eliminated_role
        ? `ĐÃ LỘ VAI TRÒ: ${roleLabel(state.eliminated_role)}`
        : "HÒA PHIẾU"}</small>
      <strong>{eliminatedPlayer
        ? `${eliminatedPlayer.name} đã bị loại.`
        : "Không ai bị loại ở vòng này."}</strong>
      <p>Kết quả được công bố trước khi hệ thống kiểm tra điều kiện thắng.</p>
    </div>
  );
}

function MrWhiteGuessingPhase({
  state,
  currentPlayer,
  guess,
  pending,
  onGuessChange,
  onSubmit,
}: {
  state: RoundStateResponse;
  currentPlayer?: RoomPlayer;
  guess: string;
  pending: boolean;
  onGuessChange: (guess: string) => void;
  onSubmit: () => void | Promise<void>;
}) {
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void onSubmit();
  }

  if (state.eliminated_player_id !== Number(currentPlayer?.id)) {
    return (
      <div className={styles.voteWaiting}>
        <strong>Mr. White đang đoán từ</strong>
        <p>Hãy chờ đáp án cuối cùng.</p>
      </div>
    );
  }

  return (
    <form className={styles.voteForm} onSubmit={submit}>
      <p>Bạn là Mr. White. Hãy đoán từ khóa của Dân thường để thắng ngay.</p>
      <input
        maxLength={80}
        onChange={(event) => onGuessChange(event.target.value)}
        placeholder="Nhập từ khóa…"
        value={guess}
      />
      <button disabled={!guess.trim() || pending} type="submit">
        {pending ? "Đang kiểm tra…" : "Gửi đáp án"}
      </button>
    </form>
  );
}

function GameFinishedPhase({ room, state }: { room: GameRoom; state: RoundStateResponse }) {
  const undercoverPlayer = state.undercover_player_id
    ? room.players.find((player) => player.id === String(state.undercover_player_id))
    : undefined;
  const mrWhitePlayer = state.mr_white_player_id
    ? room.players.find((player) => player.id === String(state.mr_white_player_id))
    : undefined;

  return (
    <div className={`${styles.voteResult} ${styles.voteCaught}`}>
      <small>KẾT THÚC VÁN</small>
      <strong>{winnerLabel(state.winner)} chiến thắng</strong>
      <p>
        Undercover: <b>{undercoverPlayer?.name ?? "—"}</b>
        {room.mrWhiteEnabled && <> · Mr. White: <b>{mrWhitePlayer?.name ?? "—"}</b></>}
      </p>
    </div>
  );
}

function formatCountdown(totalSeconds: number) {
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

function phaseTitle(phase?: RoundStateResponse["phase"]) {
  switch (phase) {
    case "REVEALING_ROLE": return "Xem vai trò bí mật";
    case "DESCRIBING": return "Lượt mô tả";
    case "VOTING": return "Ai đáng nghi nhất?";
    case "REVEALING_RESULT": return "Công bố người bị loại";
    case "MR_WHITE_GUESSING": return "Mr. White đoán từ";
    case "GAME_FINISHED": return "Kết quả ván";
    default: return "Đang tải ván chơi";
  }
}

function roleLabel(role: NonNullable<RoundStateResponse["eliminated_role"]>) {
  if (role === "undercover") return "UNDERCOVER";
  if (role === "mr_white") return "MR. WHITE";
  return "DÂN THƯỜNG";
}

function winnerLabel(winner: RoundStateResponse["winner"]) {
  if (winner === "undercover") return "Undercover";
  if (winner === "mr_white") return "Mr. White";
  return "Dân thường";
}
