"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ApiLoading } from "@/components/api-loading";
import { LogoutButton } from "@/components/logout-button";
import { useSession } from "@/components/session-guard";
import { ErrorScreen } from "@/components/error-screen";
import { ChatPanel } from "@/features/chat/chat-panel";
import { GameGuide } from "./game-guide";
import { RoomPhasePanel } from "./room-phase-panel";
import {
  colorForPlayer,
  mapRoomResponse,
  mapRoundCardResponse,
  type GameRoom,
  type PlayerCard,
  type RoomPlayer,
} from "@/features/rooms/data";
import { useRoomRealtime } from "@/features/rooms/use-room-realtime";
import { useRoomGameState } from "@/features/rooms/use-room-game-state";
import { useRoomGameCommands } from "@/features/rooms/use-room-game-commands";
import {
  ApiError,
  dealRoomRound,
  getCurrentRoomCard,
  getRoom,
  kickRoomMember,
  updateRoomDiscussion,
} from "@/lib/api";
import { restoreSession } from "@/lib/auth";
import styles from "./room.module.css";

type DealPhase = "waiting" | "dealing" | "ready";
type CardOverlay = "hidden" | "facedown" | "revealed";
type MobileRoomView = "table" | "chat" | "ranking";

export default function RoomDetailPage() {
  const params = useParams<{ roomId: string }>();
  const router = useRouter();
  const session = useSession();
  const username = session.username;
  const [room, setRoom] = useState<GameRoom | null>(null);
  const [roomState, setRoomState] = useState<
    "loading" | "ready" | "not-found" | "removed" | "error"
  >("loading");
  const [reloadToken, setReloadToken] = useState(0);
  const [cards, setCards] = useState<PlayerCard[]>([]);
  const [cardRound, setCardRound] = useState(0);
  const [cardOverlay, setCardOverlay] = useState<CardOverlay>("hidden");
  const [dealPhase, setDealPhase] = useState<DealPhase>("waiting");
  const [dealPending, setDealPending] = useState(false);
  const [dealError, setDealError] = useState("");
  const [dealtPlayerIds, setDealtPlayerIds] = useState<string[]>([]);
  const [activeDealPlayerId, setActiveDealPlayerId] = useState<string | null>(
    null,
  );
  const [showNextRoundConfirm, setShowNextRoundConfirm] = useState(false);
  const [discussionDraft, setDiscussionDraft] = useState(30);
  const [settingsPending, setSettingsPending] = useState(false);
  const [settingsError, setSettingsError] = useState("");
  const [kickPendingId, setKickPendingId] = useState<string | null>(null);
  const [kickError, setKickError] = useState("");
  const [selectedVote, setSelectedVote] = useState("");
  const [mrWhiteGuess, setMrWhiteGuess] = useState("");
  const [clockNow, setClockNow] = useState(() => Date.now());
  const [mobileView, setMobileView] = useState<MobileRoomView>("table");
  const dealTimersRef = useRef<number[]>([]);
  const realtime = useRoomRealtime(params.roomId, username);
  const gameState = useRoomGameState({
    roomId: room?.id,
    playing: room?.status === "playing",
    roomRound: room?.round,
    realtimeRevision: realtime.gameStateRevision,
    realtimeState: realtime.gameStateUpdated,
    roundStarted: realtime.roundStarted?.round,
    votesCast: realtime.voteUpdated?.votes_cast,
  });
  const {
    state: roundState,
    setState: setRoundState,
    error: roundStateError,
    setError: setRoundStateError,
  } = gameState;
  const gameCommands = useRoomGameCommands(
    room?.id,
    setRoundState,
    setRoundStateError,
  );
  const votePending = gameCommands.pending === "vote";
  const roleReadyPending = gameCommands.pending === "ready";
  const turnPending = gameCommands.pending === "finish";
  const mrWhiteGuessPending = gameCommands.pending === "guess";

  const rankedPlayers = useMemo(
    () =>
      room
        ? [...room.players].sort((left, right) => right.score - left.score)
        : [],
    [room],
  );
  const playerRanks = useMemo(
    () => new Map(rankedPlayers.map((player, index) => [player.id, index + 1])),
    [rankedPlayers],
  );

  const rosterKey = room?.id ?? "";

  useEffect(() => {
    let active = true;

    async function loadRoom(trackActivity: boolean) {
      try {
        const activeSession = await restoreSession();
        if (!activeSession) {
          router.replace("/login");
          return;
        }
        const response = await getRoom(
          activeSession.access_token,
          params.roomId,
          { trackActivity },
        );
        if (!active) return;
        const mappedRoom = mapRoomResponse(response);
        setRoom(mappedRoom);
        setDiscussionDraft(mappedRoom.discussionSeconds);
        setRoomState("ready");
      } catch (error) {
        if (!active) return;
        if (error instanceof ApiError && error.status === 404) {
          setRoom(null);
          setRoomState("not-found");
          return;
        }
        if (error instanceof ApiError && error.status === 403) {
          setRoom(null);
          setRoomState("removed");
          return;
        }
        if (trackActivity) {
          setRoom(null);
          setRoomState("error");
        }
      }
    }

    setRoomState((current) => (current === "ready" ? current : "loading"));
    void loadRoom(true);
    return () => {
      active = false;
    };
  }, [params.roomId, realtime.roomRevision, reloadToken, router]);

  useEffect(() => {
    if (realtime.players.length === 0) return;
    setRoom((current) => {
      if (!current) return current;
      const currentPlayers = new Map(
        current.players.map((player) => [player.id, player]),
      );
      return {
        ...current,
        players: realtime.players.map((player) => {
          const existing = currentPlayers.get(String(player.id));
          return {
            id: String(player.id),
            name: player.name,
            color: existing?.color ?? colorForPlayer(player.id),
            score: existing?.score ?? 0,
            host: player.host,
            current: existing?.current ?? player.name === username,
            online: player.online,
            eliminated: player.eliminated,
          };
        }),
      };
    });
  }, [realtime.players, username]);

  useEffect(() => {
    if (!realtime.roundStarted) return;
    const startedRound = realtime.roundStarted.round;
    setRoom((current) =>
      current
        ? {
            ...current,
            status: "playing",
            round: Math.max(current.round, startedRound),
          }
        : current,
    );
    setRoundState((current) => ({
      room_id: params.roomId,
      round: startedRound,
      cycle: 1,
      phase: "REVEALING_ROLE",
      phase_deadline_at:
        realtime.roundStarted?.phase_deadline_at ??
        current?.phase_deadline_at ??
        null,
      ready_players: 0,
      eligible_players: room?.players.length ?? 0,
      current_user_ready: false,
      current_turn_player_id: null,
      turn_number: 0,
      total_turns: room?.players.length ?? 0,
      turn_ends_at: null,
      eligible_voters: current?.eligible_voters ?? room?.players.length ?? 0,
      votes_cast: 0,
      current_user_vote_id: null,
      undercover_player_id: null,
      mr_white_player_id: null,
      eliminated_player_id: null,
      eliminated_role: null,
      winner: null,
      mr_white_guess_correct: null,
    }));
    setSelectedVote("");
  }, [params.roomId, realtime.roundStarted, room?.players.length]);

  useEffect(() => {
    if (room?.status !== "playing") return;
    const timer = window.setInterval(() => setClockNow(Date.now()), 1_000);
    return () => window.clearInterval(timer);
  }, [room?.status]);

  useEffect(() => {
    if (roundState?.phase === "VOTING") setSelectedVote("");
  }, [roundState?.cycle, roundState?.phase, roundState?.round]);

  useEffect(() => {
    if (roundState?.current_user_ready && roundState.round === cardRound) {
      setCardOverlay("hidden");
    }
  }, [cardRound, roundState?.current_user_ready, roundState?.round]);

  useEffect(() => {
    if (rosterKey) {
      dealTimersRef.current.forEach((timer) => window.clearTimeout(timer));
      dealTimersRef.current = [];
      setCards([]);
      setCardRound(0);
      setCardOverlay("hidden");
      setDealPhase(room?.status === "playing" ? "ready" : "waiting");
      setDealPending(false);
      setDealError("");
      setDealtPlayerIds([]);
      setActiveDealPlayerId(null);
      setShowNextRoundConfirm(false);
      setRoundState(null);
      setRoundStateError("");
      setSelectedVote("");
      setMrWhiteGuess("");
    }

    return () => {
      dealTimersRef.current.forEach((timer) => window.clearTimeout(timer));
      dealTimersRef.current = [];
    };
  }, [room?.status, rosterKey]);

  useEffect(() => {
    if (cardOverlay === "hidden") return;

    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === "Escape") setCardOverlay("hidden");
    }

    window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [cardOverlay]);

  useEffect(() => {
    if (
      roomState !== "ready" ||
      !room ||
      room.status !== "playing" ||
      room.round <= cardRound ||
      dealPhase === "dealing"
    ) {
      return;
    }

    const activeRoom = room;
    let active = true;

    async function loadCurrentCard() {
      try {
        const activeSession = await restoreSession();
        if (!activeSession) {
          router.replace("/login");
          return;
        }
        const response = await getCurrentRoomCard(
          activeSession.access_token,
          activeRoom.id,
          { trackActivity: false },
        );
        if (!active) return;
        setCards([mapRoundCardResponse(response)]);
        setCardRound(response.round);
        setDealtPlayerIds(
          activeRoom.players
            .filter((player) => !player.eliminated)
            .map((player) => player.id),
        );
        setDealPhase("ready");
        setCardOverlay("facedown");
        setDealError("");
      } catch (error) {
        if (!active || (error instanceof ApiError && error.status === 404))
          return;
        setDealError(
          error instanceof Error ? error.message : "Could not load your card",
        );
      }
    }

    void loadCurrentCard();
    return () => {
      active = false;
    };
  }, [cardRound, dealPhase, room, roomState, router]);

  if (roomState === "loading") {
    return <ApiLoading label="Loading room…" variant="page" />;
  }

  if (roomState === "not-found") {
    return (
      <ErrorScreen
        description="This room may have closed, or the invite link is incorrect."
        eyebrow="ROOM NOT FOUND"
        primaryHref="/dashboard"
        primaryLabel="Back to rooms"
        statusCode="404"
        title="This room has left the table."
      />
    );
  }

  if (roomState === "removed" || realtime.removed) {
    return (
      <ErrorScreen
        description="Chủ phòng đã mời bạn ra. Ghế của bạn hiện đã sẵn sàng cho người chơi khác."
        eyebrow="ROOM ACCESS ENDED"
        primaryHref="/dashboard"
        primaryLabel="Back to rooms"
        statusCode="403"
        title="Bạn không còn ở trong phòng này."
      />
    );
  }

  if (roomState === "error" || !room) {
    return (
      <ErrorScreen
        description="The room server could not be reached. Your room was not intentionally closed."
        eyebrow="ROOM UNAVAILABLE"
        onRetryAction={() => setReloadToken((current) => current + 1)}
        primaryHref="/dashboard"
        primaryLabel="Back to rooms"
        statusCode="500"
        title="We could not load this room."
      />
    );
  }

  const initial = username.trim().charAt(0).toUpperCase() || "P";
  const openSeats = room.maxPlayers - room.players.length;
  const onlinePlayers = room.players.filter((player) => player.online).length;
  const currentPlayer = room.players.find((player) => player.current);
  const isRoomAdmin = Boolean(currentPlayer?.host);
  const activePlayers = room.players.filter((player) => !player.eliminated);
  const activeDealPlayer = activePlayers.find(
    (player) => player.id === activeDealPlayerId,
  );
  const activeDealPosition = activeDealPlayer
    ? activePlayers.findIndex((player) => player.id === activeDealPlayer.id) + 1
    : 0;
  const currentCard = cards.find((card) => card.playerId === currentPlayer?.id);
  const currentCardDealt = Boolean(
    currentPlayer && dealtPlayerIds.includes(currentPlayer.id),
  );
  const activeDeadline =
    roundState?.phase === "DESCRIBING"
      ? roundState.turn_ends_at
      : roundState?.phase_deadline_at;
  const remainingPhaseSeconds = activeDeadline
    ? Math.max(
        0,
        Math.ceil((new Date(activeDeadline).getTime() - clockNow) / 1_000),
      )
    : 0;
  const currentTurnPlayer = roundState?.current_turn_player_id
    ? room.players.find(
        (player) => player.id === String(roundState.current_turn_player_id),
      )
    : undefined;
  const isCurrentTurn = Boolean(
    currentPlayer && currentTurnPlayer?.id === currentPlayer.id,
  );
  const canSendTurnMessage =
    !currentPlayer?.eliminated &&
    (room.status === "waiting" ||
      roundState?.phase === "GAME_FINISHED" ||
      (roundState?.phase === "DESCRIBING" && isCurrentTurn));
  const splitAt = Math.ceil(room.maxPlayers / 2);
  const seats = Array.from<RoomPlayer | null>({ length: room.maxPlayers }).fill(
    null,
  );
  room.players.forEach((player, index) => {
    seats[index] = player;
  });

  async function startGame() {
    if (!room || !isRoomAdmin || dealPhase === "dealing" || dealPending) return;

    dealTimersRef.current.forEach((timer) => window.clearTimeout(timer));
    dealTimersRef.current = [];
    setShowNextRoundConfirm(false);
    setDealPending(true);
    setDealError("");
    setCardOverlay("hidden");

    const playersToDeal =
      roundState?.phase === "GAME_FINISHED" ? room.players : activePlayers;
    try {
      const activeSession = await restoreSession();
      if (!activeSession) {
        router.replace("/login");
        return;
      }
      const response = await dealRoomRound(activeSession.access_token, room.id);
      setCards([mapRoundCardResponse(response)]);
      setCardRound(response.round);
      setRoom((current) =>
        current
          ? {
              ...current,
              status: "playing",
              round: response.round,
              version: current.version + 1,
              players: current.players.map((player) => ({
                ...player,
                eliminated: false,
              })),
            }
          : current,
      );
      setRoundState({
        room_id: room.id,
        round: response.round,
        cycle: 1,
        phase: "REVEALING_ROLE",
        phase_deadline_at: response.phase_deadline_at,
        ready_players: 0,
        eligible_players: playersToDeal.length,
        current_user_ready: false,
        current_turn_player_id: null,
        turn_number: 0,
        total_turns: playersToDeal.length,
        turn_ends_at: null,
        eligible_voters: playersToDeal.length,
        votes_cast: 0,
        current_user_vote_id: null,
        undercover_player_id: null,
        mr_white_player_id: null,
        eliminated_player_id: null,
        eliminated_role: null,
        winner: null,
        mr_white_guess_correct: null,
      });
      setClockNow(Date.now());
    } catch (error) {
      setDealError(
        error instanceof Error
          ? error.message
          : "Could not deal the next round",
      );
      return;
    } finally {
      setDealPending(false);
    }

    setDealtPlayerIds([]);
    setActiveDealPlayerId(null);
    setDealPhase("dealing");

    const firstDealDelay = 300;
    const dealStepDuration = 700;
    playersToDeal.forEach((player, index) => {
      const timer = window.setTimeout(
        () => {
          setActiveDealPlayerId(player.id);
          setDealtPlayerIds((current) => [...current, player.id]);
        },
        firstDealDelay + index * dealStepDuration,
      );
      dealTimersRef.current.push(timer);
    });

    const completionTimer = window.setTimeout(
      () => {
        setActiveDealPlayerId(null);
        setDealPhase("ready");
        setCardOverlay("facedown");
      },
      firstDealDelay + playersToDeal.length * dealStepDuration,
    );
    dealTimersRef.current.push(completionTimer);
  }

  function revealCurrentCard() {
    if (!currentCard || dealPhase !== "ready") return;
    setCardOverlay("revealed");
  }

  function handleGameAction() {
    if (!isRoomAdmin || dealPhase === "dealing" || dealPending) return;
    if (dealPhase === "ready") {
      if (roundState?.phase !== "GAME_FINISHED") return;
      setShowNextRoundConfirm(true);
      return;
    }
    startGame();
  }

  async function saveDiscussionTime() {
    if (!room || !isRoomAdmin || room.status !== "waiting" || settingsPending)
      return;
    setSettingsPending(true);
    setSettingsError("");
    try {
      const activeSession = await restoreSession();
      if (!activeSession) {
        router.replace("/login");
        return;
      }
      const response = await updateRoomDiscussion(
        activeSession.access_token,
        room.id,
        discussionDraft,
      );
      setRoom(mapRoomResponse(response));
    } catch (error) {
      setSettingsError(
        error instanceof Error
          ? error.message
          : "Could not save discussion time",
      );
    } finally {
      setSettingsPending(false);
    }
  }

  async function kickPlayer(player: RoomPlayer) {
    if (
      !room ||
      !isRoomAdmin ||
      room.status !== "waiting" ||
      player.host ||
      player.current
    )
      return;
    if (!window.confirm(`Remove ${player.name} from this room?`)) return;

    setKickPendingId(player.id);
    setKickError("");
    try {
      const activeSession = await restoreSession();
      if (!activeSession) {
        router.replace("/login");
        return;
      }
      const response = await kickRoomMember(
        activeSession.access_token,
        room.id,
        Number(player.id),
      );
      setRoom(mapRoomResponse(response));
    } catch (error) {
      setKickError(
        error instanceof Error ? error.message : "Could not remove this player",
      );
    } finally {
      setKickPendingId(null);
    }
  }

  async function submitVote() {
    if (!room || !selectedVote || votePending || roundState?.phase !== "VOTING")
      return;
    await gameCommands.castVote(Number(selectedVote));
  }

  async function confirmRoleCard() {
    if (!room || roleReadyPending) return;

    // Closing the private card is a local UI action. After a reload the round
    // may already be past REVEALING_ROLE, where the backend correctly rejects
    // a late "ready" command; that rejection must not keep the overlay open.
    setCardOverlay("hidden");

    if (
      roundState?.phase === "REVEALING_ROLE" &&
      !roundState.current_user_ready
    ) {
      await gameCommands.confirmRole();
    }
  }

  async function finishCurrentTurn() {
    if (
      !room ||
      !isCurrentTurn ||
      turnPending ||
      roundState?.phase !== "DESCRIBING"
    )
      return;
    await gameCommands.finishTurn();
  }

  async function submitWhiteGuess() {
    if (!room || !mrWhiteGuess.trim() || mrWhiteGuessPending) return;
    await gameCommands.submitGuess(mrWhiteGuess);
  }

  return (
    <main className={styles.page}>
      <header className={styles.header}>
        <Link className={styles.brand} href="/">
          <span aria-hidden="true">F</span>
          falzo
        </Link>

        <Link className={styles.backLink} href="/dashboard">
          <span aria-hidden="true">←</span>
          All rooms
        </Link>

        <div className={styles.account}>
          <span className={styles.avatar} aria-hidden="true">
            {initial}
          </span>
          <span>{username}</span>
          <LogoutButton />
        </div>
      </header>

      <GameGuide mrWhiteEnabled={room.mrWhiteEnabled} />

      {dealPhase === "dealing" && activeDealPlayer && (
        <div className={styles.dealOverlay} role="status" aria-live="assertive">
          <div className={styles.dealShowcase} key={activeDealPlayer.id}>
            <div className={styles.dealRecipient}>
              <span
                className={`${styles.dealAvatar} ${styles[activeDealPlayer.color]}`}
                aria-hidden="true"
              >
                {playerInitials(activeDealPlayer.name)}
              </span>
              <div>
                <small>DEALING CARD TO</small>
                <strong>
                  {activeDealPlayer.current
                    ? `${activeDealPlayer.name} · You`
                    : activeDealPlayer.name}
                </strong>
              </div>
            </div>

            <div
              className={styles.largeDealCard}
              aria-label={`Private card for ${activeDealPlayer.name}`}
            >
              <span>FALZO</span>
              <strong>?</strong>
              <small>KEEP IT SECRET</small>
            </div>

            <div className={styles.dealCounter}>
              <span>
                Card {activeDealPosition} of {activePlayers.length}
              </span>
              <div aria-hidden="true">
                {activePlayers.map((player) => (
                  <span
                    className={
                      dealtPlayerIds.includes(player.id)
                        ? styles.progressDone
                        : ""
                    }
                    key={player.id}
                  />
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {cardOverlay === "facedown" && currentCard && (
        <div
          aria-label="Your new secret card is ready"
          aria-modal="true"
          className={`${styles.cardRevealOverlay} ${styles.cardPromptOverlay}`}
          role="dialog"
        >
          <button
            autoFocus
            className={styles.facedownSecretCard}
            onClick={revealCurrentCard}
            type="button"
          >
            <span>YOUR NEW CARD</span>
            <strong>?</strong>
            <small>Tap to reveal</small>
          </button>
          <p>Your private card is ready</p>
        </div>
      )}

      {cardOverlay === "revealed" && currentCard && (
        <div
          aria-label="Your revealed secret card"
          aria-modal="true"
          className={styles.cardRevealOverlay}
          role="dialog"
        >
          <div className={styles.revealedSecretCard}>
            <span>YOUR SECRET WORD</span>
            <strong>{currentCard.word || "Không có từ khóa"}</strong>
            <small>{currentCard.role}</small>
          </div>
          <button
            autoFocus
            className={styles.confirmRoleButton}
            disabled={
              roundState?.phase === "REVEALING_ROLE" && roleReadyPending
            }
            onClick={confirmRoleCard}
            type="button"
          >
            {roundState?.phase === "REVEALING_ROLE" && roleReadyPending
              ? "Đang xác nhận…"
              : roundState?.phase === "REVEALING_ROLE" &&
                  !roundState.current_user_ready
                ? "Đã hiểu"
                : "Ẩn thẻ"}
          </button>
          <p>Ghi nhớ vai trò và không để người khác nhìn thấy.</p>
        </div>
      )}

      <section className={styles.content} data-mobile-view={mobileView}>
        <div className={styles.roomHeading}>
          <div className={styles.roomIntro}>
            <div className={styles.eyebrow}>
              <span className={`${styles.statusDot} ${styles[room.status]}`} />
              {room.status === "waiting"
                ? "WAITING FOR PLAYERS"
                : `ROUND ${room.round} IN PROGRESS`}
            </div>
            <h1>{room.name}</h1>
            <div className={styles.roomSummary}>
              <div className={styles.inviteSummary}>
                <small>MÃ PHÒNG</small>
                <strong>#{room.code}</strong>
                <span>Gửi mã này để mời bạn bè</span>
              </div>
              <div>
                <small>NGƯỜI CHƠI</small>
                <strong>
                  {room.players.length}
                  <i>/{room.maxPlayers}</i>
                </strong>
                <span>{onlinePlayers} đang online</span>
              </div>
              <div>
                <small>CHỖ TRỐNG</small>
                <strong>{openSeats}</strong>
                <span>
                  {openSeats > 0 ? "Vẫn có thể tham gia" : "Phòng đã đủ người"}
                </span>
              </div>
              <div>
                <small>NGÔN NGỮ</small>
                <strong className={styles.languageValue}>
                  {room.language === "vi" ? "VI" : "EN"}
                </strong>
                <span>{room.language === "vi" ? "Tiếng Việt" : "English"}</span>
              </div>
            </div>
          </div>

          <div className={styles.gameControl}>
            {room.status === "waiting" && isRoomAdmin && (
              <div className={styles.discussionSetting}>
                <label htmlFor="discussion-time">
                  Thời gian mỗi lượt mô tả
                </label>
                <div>
                  <select
                    id="discussion-time"
                    onChange={(event) =>
                      setDiscussionDraft(Number(event.target.value))
                    }
                    value={discussionDraft}
                  >
                    <option value={10}>10 giây</option>
                    <option value={15}>15 giây</option>
                    <option value={20}>20 giây</option>
                    <option value={30}>30 giây</option>
                  </select>
                  <button
                    disabled={
                      settingsPending ||
                      discussionDraft === room.discussionSeconds
                    }
                    onClick={saveDiscussionTime}
                    type="button"
                  >
                    {settingsPending ? "Saving…" : "Save"}
                  </button>
                </div>
                <small>
                  Mỗi người có một lượt; hết giờ hệ thống tự chuyển sang người
                  tiếp theo.
                </small>
                {settingsError && (
                  <small className={styles.dealError}>{settingsError}</small>
                )}
              </div>
            )}

            <button
              aria-describedby={!isRoomAdmin ? "admin-action-hint" : undefined}
              className={styles.dealButton}
              disabled={
                dealPhase === "dealing" ||
                dealPending ||
                !isRoomAdmin ||
                (dealPhase === "ready" && roundState?.phase !== "GAME_FINISHED")
              }
              onClick={handleGameAction}
              title={
                !isRoomAdmin
                  ? "Only the room admin can control rounds"
                  : undefined
              }
              type="button"
            >
              <span aria-hidden="true">
                {!isRoomAdmin
                  ? "◇"
                  : dealPending
                    ? "…"
                    : dealPhase === "waiting"
                      ? "▶"
                      : dealPhase === "dealing"
                        ? "…"
                        : "↻"}
              </span>
              {dealPending
                ? "Preparing cards…"
                : dealPhase === "waiting"
                  ? "Start game"
                  : dealPhase === "dealing"
                    ? `Dealing ${dealtPlayerIds.length}/${activePlayers.length}`
                    : roundState?.phase === "GAME_FINISHED"
                      ? "Chơi lại"
                      : "Round in progress"}
            </button>

            {!isRoomAdmin && (
              <small className={styles.adminHint} id="admin-action-hint">
                Only the room admin can control rounds
              </small>
            )}

            {dealError && (
              <small className={styles.dealError} role="alert">
                {dealError}
              </small>
            )}

            {showNextRoundConfirm && (
              <div
                aria-labelledby="next-round-confirm-title"
                className={styles.roundConfirm}
                role="alertdialog"
              >
                <strong id="next-round-confirm-title">
                  Chơi lại với bộ từ mới?
                </strong>
                <p>Mọi người vẫn ở nguyên trong phòng và được chia vai lại.</p>
                <div>
                  <button
                    autoFocus
                    className={styles.confirmNo}
                    onClick={() => setShowNextRoundConfirm(false)}
                    type="button"
                  >
                    No
                  </button>
                  <button
                    className={styles.confirmYes}
                    onClick={startGame}
                    type="button"
                  >
                    Yes
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>

        <nav
          className={styles.mobileRoomNav}
          aria-label="Điều hướng trong phòng"
        >
          <button
            aria-controls="game-table-panel"
            aria-pressed={mobileView === "table"}
            onClick={() => setMobileView("table")}
            type="button"
          >
            <span aria-hidden="true">◆</span>
            <span>
              <strong>Bàn chơi</strong>
              <small>{activePlayers.length} người còn lại</small>
            </span>
          </button>
          <button
            aria-controls="action-chat-panel"
            aria-pressed={mobileView === "chat"}
            onClick={() => setMobileView("chat")}
            type="button"
          >
            <span aria-hidden="true">◌</span>
            <span>
              <strong>Hành động</strong>
              <small>
                {room.status === "playing"
                  ? "Diễn biến & chat"
                  : `${realtime.messages.length} tin nhắn`}
              </small>
            </span>
          </button>
          <button
            aria-controls="room-ranking-panel"
            aria-pressed={mobileView === "ranking"}
            onClick={() => setMobileView("ranking")}
            type="button"
          >
            <span aria-hidden="true">↗</span>
            <span>
              <strong>Xếp hạng</strong>
              <small>Điểm người chơi</small>
            </span>
          </button>
        </nav>

        <div className={styles.roomLayout}>
          <section
            className={styles.board}
            id="game-table-panel"
            aria-label={`${room.name} seating table`}
          >
            <div className={styles.seatRow}>
              {seats.slice(0, splitAt).map((player, index) => (
                <PlayerSeat
                  card={cards.find((card) => card.playerId === player?.id)}
                  dealPhase={dealPhase}
                  dealt={Boolean(player && dealtPlayerIds.includes(player.id))}
                  key={player?.id ?? `empty-top-${index}`}
                  onReveal={revealCurrentCard}
                  player={player}
                  rank={player ? playerRanks.get(player.id) : undefined}
                />
              ))}
            </div>

            <div
              className={`${styles.table} ${dealPhase === "dealing" ? styles.dealingTable : ""}`}
            >
              <div className={styles.tableMark} aria-hidden="true">
                ?
              </div>
              <p>UNDERCOVER</p>
              <h2>
                {dealPhase === "waiting"
                  ? "Ready to play?"
                  : dealPhase === "dealing"
                    ? "Dealing cards…"
                    : "Read the room."}
              </h2>
              <span>
                {dealPhase === "waiting"
                  ? "Start the game to deal one private card to every seated player."
                  : dealPhase === "dealing"
                    ? "Cards are moving around the table one player at a time."
                    : "Everyone has a card. Your secret stays private until you reveal it."}
              </span>
              <div className={styles.tableStatus} aria-live="polite">
                <span>
                  {dealPhase === "waiting"
                    ? "Waiting to start"
                    : `${dealtPlayerIds.length}/${activePlayers.length} cards dealt`}
                </span>
                <span>{openSeats} open seats</span>
              </div>
            </div>

            <div className={styles.seatRow}>
              {seats.slice(splitAt).map((player, index) => (
                <PlayerSeat
                  card={cards.find((card) => card.playerId === player?.id)}
                  dealPhase={dealPhase}
                  dealt={Boolean(player && dealtPlayerIds.includes(player.id))}
                  key={player?.id ?? `empty-bottom-${index}`}
                  onReveal={revealCurrentCard}
                  player={player}
                  rank={player ? playerRanks.get(player.id) : undefined}
                />
              ))}
            </div>
          </section>

          <div className={styles.rightRail} id="action-chat-panel">
            {room.status === "playing" && (
              <RoomPhasePanel
                currentCardAvailable={Boolean(currentCard)}
                mrWhiteGuess={mrWhiteGuess}
                mrWhiteGuessPending={mrWhiteGuessPending}
                onFinishTurn={finishCurrentTurn}
                onMrWhiteGuessChange={setMrWhiteGuess}
                onRevealCard={revealCurrentCard}
                onSelectedVoteChange={setSelectedVote}
                onSubmitVote={submitVote}
                onSubmitWhiteGuess={submitWhiteGuess}
                remainingSeconds={remainingPhaseSeconds}
                room={room}
                selectedVote={selectedVote}
                state={roundState}
                stateError={roundStateError}
                turnPending={turnPending}
                votePending={votePending}
              />
            )}

            <ChatPanel
              className={styles.roomChat}
              connected={realtime.connected && canSendTurnMessage}
              contextLabel={`#${room.code}`}
              disabledPlaceholder={
                currentPlayer?.eliminated
                  ? "Spectators cannot chat"
                  : roundState?.phase === "DESCRIBING" && !isCurrentTurn
                    ? "Chờ đến lượt mô tả của bạn"
                    : roundState?.phase !== "GAME_FINISHED" &&
                        room.status === "playing"
                      ? "Chat đang tạm khóa ở giai đoạn này"
                      : "Reconnecting…"
              }
              inputPlaceholder={
                isCurrentTurn ? "Nhập lời mô tả…" : "Message the room…"
              }
              key={room.id}
              messages={realtime.messages}
              onSendMessageAction={realtime.sendChat}
              presence={
                currentPlayer?.eliminated
                  ? "room"
                  : realtime.connected
                    ? "online"
                    : "offline"
              }
              subtitle={
                currentPlayer?.eliminated
                  ? `${onlinePlayers}/${room.players.length} online · spectating`
                  : `${onlinePlayers}/${room.players.length} online · ${realtime.status}`
              }
              title="Room chat"
            />

            <aside
              className={styles.myAccount}
              aria-label="Your player account"
            >
              <div className={styles.myIdentity}>
                <span className={styles.myAvatar} aria-hidden="true">
                  {initial}
                </span>
                <div>
                  <small>
                    {isRoomAdmin ? "YOUR ACCOUNT · ROOM ADMIN" : "YOUR ACCOUNT"}
                  </small>
                  <strong>{username}</strong>
                  <span>
                    {currentPlayer?.eliminated
                      ? "Spectating this round"
                      : dealPhase === "ready"
                        ? "Ready to give a clue"
                        : dealPhase === "dealing"
                          ? "Your card is on the way"
                          : `Seated in ${room.name}`}
                  </span>
                </div>
              </div>

              <button
                aria-label="Reveal your secret card"
                className={`${styles.myCard} ${currentCardDealt ? styles.myCardReady : ""}`}
                disabled={
                  !currentCard || !currentCardDealt || dealPhase !== "ready"
                }
                onClick={revealCurrentCard}
                type="button"
              >
                {!currentCardDealt || !currentCard ? (
                  <>
                    <small>
                      {dealPhase === "dealing" ? "INCOMING" : "NO CARD"}
                    </small>
                    <strong>—</strong>
                    <span>
                      {dealPhase === "dealing" ? "Please wait" : "Start game"}
                    </span>
                  </>
                ) : (
                  <>
                    <small>YOUR CARD</small>
                    <strong>?</strong>
                    <span>Tap to reveal</span>
                  </>
                )}
              </button>
            </aside>
          </div>
        </div>

        <section
          className={styles.leaderboard}
          id="room-ranking-panel"
          aria-labelledby="leaderboard-title"
        >
          <div className={styles.leaderboardHeading}>
            <div>
              <p>ROOM STANDINGS</p>
              <h2 id="leaderboard-title">Leaderboard</h2>
            </div>
            <span>
              {room.players.length} players · Round {room.round}
            </span>
          </div>

          <div className={styles.leaderboardTable}>
            <table>
              <thead>
                <tr>
                  <th scope="col">Rank</th>
                  <th scope="col">Player</th>
                  <th scope="col">Status</th>
                  <th scope="col">Score</th>
                  {room.status === "waiting" && isRoomAdmin && (
                    <th scope="col">Action</th>
                  )}
                </tr>
              </thead>
              <tbody>
                {rankedPlayers.map((player, index) => (
                  <tr
                    className={
                      player.current ? styles.currentRanking : undefined
                    }
                    key={player.id}
                  >
                    <td>
                      <span
                        className={`${styles.rankNumber} ${index < 3 ? styles.topRank : ""}`}
                        key={`${player.id}-${index + 1}`}
                      >
                        {String(index + 1).padStart(2, "0")}
                      </span>
                    </td>
                    <td>
                      <div className={styles.rankedPlayer}>
                        <PlayerAvatar compact player={player} />
                        <strong title={player.name}>{player.name}</strong>
                        {player.current && <small>YOU</small>}
                      </div>
                    </td>
                    <td>
                      <span
                        className={`${styles.playerStatus} ${
                          player.online
                            ? styles.onlineStatus
                            : styles.offlineStatus
                        }`}
                      >
                        <span
                          className={styles.statusIndicator}
                          aria-hidden="true"
                        />
                        {player.eliminated
                          ? "Spectator"
                          : player.host
                            ? "Admin"
                            : "Player"}{" "}
                        · {player.online ? "Online" : "Offline"}
                      </span>
                    </td>
                    <td className={styles.scoreCell}>
                      <span className={styles.scoreBadge}>
                        <strong>{player.score}</strong>
                        <small>PTS</small>
                      </span>
                    </td>
                    {room.status === "waiting" && isRoomAdmin && (
                      <td className={styles.kickCell}>
                        {!player.host && !player.current ? (
                          <button
                            className={styles.kickButton}
                            disabled={kickPendingId !== null}
                            onClick={() => void kickPlayer(player)}
                            type="button"
                          >
                            {kickPendingId === player.id
                              ? "Removing…"
                              : "Remove"}
                          </button>
                        ) : (
                          <span aria-label="Not removable">—</span>
                        )}
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {kickError && (
            <p className={styles.kickError} role="alert">
              {kickError}
            </p>
          )}
        </section>

        <footer className={styles.roomNote}>
          <div>
            <span aria-hidden="true">i</span>
            <p>
              Room membership, rounds, and private cards are synced with the
              backend. Chat and online presence are delivered live through
              WebSocket and Redis.
            </p>
          </div>
          <span>
            {realtime.connected
              ? "REALTIME CONNECTED"
              : "REALTIME RECONNECTING"}
          </span>
        </footer>
      </section>
    </main>
  );
}

type PlayerSeatProps = {
  player: RoomPlayer | null;
  card?: PlayerCard;
  dealPhase: DealPhase;
  dealt: boolean;
  onReveal: () => void;
  rank?: number;
};

type PlayerAvatarProps = {
  compact?: boolean;
  player: RoomPlayer;
  rank?: number;
};

function PlayerAvatar({ compact = false, player, rank }: PlayerAvatarProps) {
  return (
    <span
      aria-label={`${player.name}${rank ? `, rank ${rank}` : ""}, ${player.online ? "online" : "offline"}`}
      className={`${styles.playerAvatarFrame} ${compact ? styles.compactAvatarFrame : ""} ${
        player.current ? styles.currentPlayerAvatar : ""
      } ${rank ? styles.rankedAvatarFrame : ""}`}
      role="img"
      title={`${player.name} - ${player.online ? "Online" : "Offline"}`}
    >
      <span
        aria-hidden="true"
        className={`${styles.playerAvatar} ${styles[player.color]} ${
          player.online ? "" : styles.offlineAvatar
        }`}
      >
        {playerInitials(player.name)}
      </span>
      <span
        aria-hidden="true"
        className={`${styles.avatarPresence} ${
          player.online
            ? styles.avatarPresenceOnline
            : styles.avatarPresenceOffline
        }`}
      />
      {rank && (
        <span
          aria-hidden="true"
          className={`${styles.avatarRank} ${rank <= 3 ? styles.topAvatarRank : ""}`}
          key={rank}
        >
          #{String(rank).padStart(2, "0")}
        </span>
      )}
    </span>
  );
}

function playerInitials(name: string) {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  return parts
    .slice(0, 2)
    .map((part) => part.charAt(0))
    .join("")
    .toUpperCase();
}

function PlayerSeat({
  player,
  card,
  dealPhase,
  dealt,
  onReveal,
  rank,
}: PlayerSeatProps) {
  if (!player) {
    return (
      <article className={`${styles.seat} ${styles.emptySeat}`}>
        <div className={styles.playerInfo}>
          <span className={styles.emptyAvatar} aria-hidden="true">
            +
          </span>
          <div>
            <strong>Open seat</strong>
            <small>Waiting for player</small>
          </div>
        </div>
        <div className={styles.emptyCard} aria-hidden="true" />
      </article>
    );
  }

  const canReveal = player.current && card && dealt && dealPhase === "ready";

  return (
    <article
      className={`${styles.seat} ${player.current ? styles.currentSeat : ""}`}
    >
      <div className={styles.playerInfo}>
        <PlayerAvatar player={player} rank={rank} />
        <div>
          <strong title={player.name}>{player.name}</strong>
          <small>
            {player.eliminated
              ? player.current
                ? "You · Spectator"
                : "Spectator"
              : player.current
                ? player.host
                  ? "You · Admin"
                  : "You"
                : player.host
                  ? "Room admin"
                  : "Player"}
            {` · ${player.online ? "Online" : "Offline"}`}
          </small>
        </div>
      </div>

      {player.eliminated ? (
        <div
          className={`${styles.playerCard} ${styles.spectatorCard}`}
          aria-label={`${player.name} is spectating`}
        >
          <small>SPECTATOR</small>
          <strong>×</strong>
          <span>Watching only</span>
        </div>
      ) : !dealt ? (
        <div
          className={`${styles.playerCard} ${styles.undealtCard}`}
          aria-label={`${player.name} is waiting for a card`}
        >
          <small>{dealPhase === "dealing" ? "DEALING" : "NO CARD"}</small>
          <strong>—</strong>
          <span>{dealPhase === "dealing" ? "Waiting" : "Start game"}</span>
        </div>
      ) : canReveal ? (
        <button
          aria-label="Reveal your secret card"
          className={`${styles.playerCard} ${styles.dealtCard}`}
          onClick={onReveal}
          type="button"
        >
          <small>YOUR CARD</small>
          <strong>?</strong>
          <span>Tap to reveal</span>
        </button>
      ) : (
        <div
          className={`${styles.playerCard} ${styles.dealtCard}`}
          aria-label={`${player.name} has a private card`}
        >
          <small>SECRET CARD</small>
          <strong>?</strong>
          <span>Private</span>
        </div>
      )}
    </article>
  );
}
