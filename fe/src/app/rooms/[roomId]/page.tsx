"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { ApiLoading } from "@/components/api-loading";
import { LogoutButton } from "@/components/logout-button";
import { useSession } from "@/components/session-guard";
import { ErrorScreen } from "@/components/error-screen";
import { ChatPanel, type ChatMessage } from "@/features/chat/chat-panel";
import {
  mapRoomResponse,
  mapRoundCardResponse,
  type GameRoom,
  type PlayerCard,
  type RoomPlayer,
} from "@/features/rooms/data";
import { ApiError, dealRoomRound, getCurrentRoomCard, getRoom } from "@/lib/api";
import { restoreSession } from "@/lib/auth";
import styles from "./room.module.css";

type DealPhase = "waiting" | "dealing" | "ready";
type CardOverlay = "hidden" | "facedown" | "revealed";

export default function RoomDetailPage() {
  const params = useParams<{ roomId: string }>();
  const router = useRouter();
  const session = useSession();
  const username = session.username;
  const [room, setRoom] = useState<GameRoom | null>(null);
  const [roomState, setRoomState] = useState<"loading" | "ready" | "not-found" | "error">("loading");
  const [reloadToken, setReloadToken] = useState(0);
  const [cards, setCards] = useState<PlayerCard[]>([]);
  const [cardRound, setCardRound] = useState(0);
  const [cardOverlay, setCardOverlay] = useState<CardOverlay>("hidden");
  const [dealPhase, setDealPhase] = useState<DealPhase>("waiting");
  const [dealPending, setDealPending] = useState(false);
  const [dealError, setDealError] = useState("");
  const [dealtPlayerIds, setDealtPlayerIds] = useState<string[]>([]);
  const [activeDealPlayerId, setActiveDealPlayerId] = useState<string | null>(null);
  const [showNextRoundConfirm, setShowNextRoundConfirm] = useState(false);
  const dealTimersRef = useRef<number[]>([]);

  const initialRoomMessages = useMemo(
    () => (room ? createInitialMessages(room.players, room.status) : []),
    [room],
  );
  const rankedPlayers = useMemo(
    () => room
      ? [...room.players].sort((left, right) => right.score - left.score)
      : [],
    [room],
  );

  const rosterKey = room
    ? `${room.id}:${room.players.map((player) => player.id).join(",")}`
    : "";

  useEffect(() => {
    let active = true;

    async function loadRoom(trackActivity: boolean) {
      try {
        const activeSession = await restoreSession();
        if (!activeSession) {
          router.replace("/login");
          return;
        }
        const response = await getRoom(activeSession.access_token, params.roomId, { trackActivity });
        if (!active) return;
        setRoom(mapRoomResponse(response));
        setRoomState("ready");
      } catch (error) {
        if (!active) return;
        if (error instanceof ApiError && error.status === 404) {
          setRoom(null);
          setRoomState("not-found");
          return;
        }
        if (trackActivity) {
          setRoom(null);
          setRoomState("error");
        }
      }
    }

    setRoomState("loading");
    void loadRoom(true);
    const pollTimer = window.setInterval(() => void loadRoom(false), 3000);

    return () => {
      active = false;
      window.clearInterval(pollTimer);
    };
  }, [params.roomId, reloadToken, router]);

  useEffect(() => {
    if (rosterKey) {
      dealTimersRef.current.forEach((timer) => window.clearTimeout(timer));
      dealTimersRef.current = [];
      setCards([]);
      setCardRound(0);
      setCardOverlay("hidden");
      setDealPhase("waiting");
      setDealPending(false);
      setDealError("");
      setDealtPlayerIds([]);
      setActiveDealPlayerId(null);
      setShowNextRoundConfirm(false);
    }

    return () => {
      dealTimersRef.current.forEach((timer) => window.clearTimeout(timer));
      dealTimersRef.current = [];
    };
  }, [rosterKey]);

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
      roomState !== "ready"
      || !room
      || room.status !== "playing"
      || room.round <= cardRound
      || dealPhase === "dealing"
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
        setDealtPlayerIds(activeRoom.players.map((player) => player.id));
        setDealPhase("ready");
        setCardOverlay("facedown");
        setDealError("");
      } catch (error) {
        if (!active || (error instanceof ApiError && error.status === 404)) return;
        setDealError(error instanceof Error ? error.message : "Could not load your card");
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

  if (roomState === "error" || !room) {
    return (
      <ErrorScreen
        description="The room server could not be reached. Your room was not intentionally closed."
        eyebrow="ROOM UNAVAILABLE"
        onRetry={() => setReloadToken((current) => current + 1)}
        primaryHref="/dashboard"
        primaryLabel="Back to rooms"
        statusCode="500"
        title="We could not load this room."
      />
    );
  }

  const initial = username.trim().charAt(0).toUpperCase() || "P";
  const openSeats = room.maxPlayers - room.players.length;
  const currentPlayer = room.players.find((player) => player.current);
  const isRoomAdmin = Boolean(currentPlayer?.host);
  const activeDealPlayer = room.players.find((player) => player.id === activeDealPlayerId);
  const activeDealPosition = activeDealPlayer
    ? room.players.findIndex((player) => player.id === activeDealPlayer.id) + 1
    : 0;
  const currentCard = cards.find((card) => card.playerId === currentPlayer?.id);
  const currentCardDealt = Boolean(
    currentPlayer && dealtPlayerIds.includes(currentPlayer.id),
  );
  const splitAt = Math.ceil(room.maxPlayers / 2);
  const seats = Array.from<RoomPlayer | null>({ length: room.maxPlayers }).fill(null);
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

    try {
      const activeSession = await restoreSession();
      if (!activeSession) {
        router.replace("/login");
        return;
      }
      const response = await dealRoomRound(activeSession.access_token, room.id);
      setCards([mapRoundCardResponse(response)]);
      setCardRound(response.round);
      setRoom((current) => current ? {
        ...current,
        status: "playing",
        round: response.round,
        version: current.version + 1,
      } : current);
    } catch (error) {
      setDealError(error instanceof Error ? error.message : "Could not deal the next round");
      return;
    } finally {
      setDealPending(false);
    }

    setDealtPlayerIds([]);
    setActiveDealPlayerId(null);
    setDealPhase("dealing");

    const firstDealDelay = 300;
    const dealStepDuration = 700;
    room.players.forEach((player, index) => {
      const timer = window.setTimeout(() => {
        setActiveDealPlayerId(player.id);
        setDealtPlayerIds((current) => [...current, player.id]);
      }, firstDealDelay + index * dealStepDuration);
      dealTimersRef.current.push(timer);
    });

    const completionTimer = window.setTimeout(() => {
      setActiveDealPlayerId(null);
      setDealPhase("ready");
      setCardOverlay("facedown");
    }, firstDealDelay + room.players.length * dealStepDuration);
    dealTimersRef.current.push(completionTimer);
  }

  function revealCurrentCard() {
    if (!currentCard || dealPhase !== "ready") return;
    setCardOverlay("revealed");
  }

  function handleGameAction() {
    if (!isRoomAdmin || dealPhase === "dealing" || dealPending) return;
    if (dealPhase === "ready") {
      setShowNextRoundConfirm(true);
      return;
    }
    startGame();
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
          <span className={styles.avatar} aria-hidden="true">{initial}</span>
          <span>{username}</span>
          <LogoutButton />
        </div>
      </header>

      {dealPhase === "dealing" && activeDealPlayer && (
        <div className={styles.dealOverlay} role="status" aria-live="assertive">
          <div className={styles.dealShowcase} key={activeDealPlayer.id}>
            <div className={styles.dealRecipient}>
              <span
                className={`${styles.dealAvatar} ${styles[activeDealPlayer.color]}`}
                aria-hidden="true"
              >
                {activeDealPlayer.name.charAt(0).toUpperCase()}
              </span>
              <div>
                <small>DEALING CARD TO</small>
                <strong>
                  {activeDealPlayer.current ? `${activeDealPlayer.name} · You` : activeDealPlayer.name}
                </strong>
              </div>
            </div>

            <div className={styles.largeDealCard} aria-label={`Private card for ${activeDealPlayer.name}`}>
              <span>FALZO</span>
              <strong>?</strong>
              <small>KEEP IT SECRET</small>
            </div>

            <div className={styles.dealCounter}>
              <span>Card {activeDealPosition} of {room.players.length}</span>
              <div aria-hidden="true">
                {room.players.map((player) => (
                  <span
                    className={dealtPlayerIds.includes(player.id) ? styles.progressDone : ""}
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
          onClick={() => setCardOverlay("hidden")}
          role="dialog"
        >
          <button
            autoFocus
            className={styles.revealedSecretCard}
            onClick={() => setCardOverlay("hidden")}
            type="button"
          >
            <span>YOUR SECRET WORD</span>
            <strong>{currentCard.word}</strong>
            <small>{currentCard.role}</small>
          </button>
          <p>Click anywhere to hide your card</p>
        </div>
      )}

      <section className={styles.content}>
        <div className={styles.roomHeading}>
          <div>
            <div className={styles.eyebrow}>
              <span className={`${styles.statusDot} ${styles[room.status]}`} />
              {room.status === "waiting" ? "WAITING FOR PLAYERS" : `ROUND ${room.round} IN PROGRESS`}
            </div>
            <h1>{room.name}</h1>
            <p>
              Room <strong>#{room.code}</strong>
              <span aria-hidden="true">·</span>
              {room.players.length}/{room.maxPlayers} players
              <span aria-hidden="true">·</span>
              {openSeats} seats open
              <span aria-hidden="true">·</span>
              {room.language === "vi" ? "Tiếng Việt" : "English"}
            </p>
          </div>

          <div className={styles.gameControl}>
            <button
              aria-describedby={!isRoomAdmin ? "admin-action-hint" : undefined}
              className={styles.dealButton}
              disabled={dealPhase === "dealing" || dealPending || !isRoomAdmin}
              onClick={handleGameAction}
              title={!isRoomAdmin ? "Only the room admin can control rounds" : undefined}
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
                  ? `Dealing ${dealtPlayerIds.length}/${room.players.length}`
                  : "Deal next round"}
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
                <strong id="next-round-confirm-title">Deal the next round?</strong>
                <p>Current cards will be replaced for every player.</p>
                <div>
                  <button
                    autoFocus
                    className={styles.confirmNo}
                    onClick={() => setShowNextRoundConfirm(false)}
                    type="button"
                  >
                    No
                  </button>
                  <button className={styles.confirmYes} onClick={startGame} type="button">
                    Yes
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>

        <div className={styles.roomLayout}>
          <section className={styles.board} aria-label={`${room.name} seating table`}>
            <div className={styles.seatRow}>
              {seats.slice(0, splitAt).map((player, index) => (
                <PlayerSeat
                  card={cards.find((card) => card.playerId === player?.id)}
                  dealPhase={dealPhase}
                  dealt={Boolean(player && dealtPlayerIds.includes(player.id))}
                  key={player?.id ?? `empty-top-${index}`}
                  onReveal={revealCurrentCard}
                  player={player}
                />
              ))}
            </div>

            <div className={`${styles.table} ${dealPhase === "dealing" ? styles.dealingTable : ""}`}>
              <div className={styles.tableMark} aria-hidden="true">?</div>
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
                    : `${dealtPlayerIds.length}/${room.players.length} cards dealt`}
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
                />
              ))}
            </div>
          </section>

          <div className={styles.rightRail}>
            <ChatPanel
              className={styles.roomChat}
              contextLabel={`#${room.code}`}
              currentUsername={username}
              initialMessages={initialRoomMessages}
              inputPlaceholder="Message the room…"
              key={room.id}
              presence="room"
              subtitle={`${room.players.length} players here`}
              title="Room chat"
            />

            <aside className={styles.myAccount} aria-label="Your player account">
              <div className={styles.myIdentity}>
                <span className={styles.myAvatar} aria-hidden="true">{initial}</span>
                <div>
                  <small>{isRoomAdmin ? "YOUR ACCOUNT · ROOM ADMIN" : "YOUR ACCOUNT"}</small>
                  <strong>{username}</strong>
                  <span>
                    {dealPhase === "ready"
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
                disabled={!currentCard || !currentCardDealt || dealPhase !== "ready"}
                onClick={revealCurrentCard}
                type="button"
              >
                {!currentCardDealt || !currentCard ? (
                  <>
                    <small>{dealPhase === "dealing" ? "INCOMING" : "NO CARD"}</small>
                    <strong>—</strong>
                    <span>{dealPhase === "dealing" ? "Please wait" : "Start game"}</span>
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

        <section className={styles.leaderboard} aria-labelledby="leaderboard-title">
          <div className={styles.leaderboardHeading}>
            <div>
              <p>ROOM STANDINGS</p>
              <h2 id="leaderboard-title">Leaderboard</h2>
            </div>
            <span>{room.players.length} players · Round {room.round}</span>
          </div>

          <div className={styles.leaderboardTable}>
            <table>
              <thead>
                <tr>
                  <th scope="col">Rank</th>
                  <th scope="col">Player</th>
                  <th scope="col">Status</th>
                  <th scope="col">Score</th>
                </tr>
              </thead>
              <tbody>
                {rankedPlayers.map((player, index) => (
                  <tr className={player.current ? styles.currentRanking : undefined} key={player.id}>
                    <td>
                      <span className={`${styles.rankNumber} ${index < 3 ? styles.topRank : ""}`}>
                        {String(index + 1).padStart(2, "0")}
                      </span>
                    </td>
                    <td>
                      <div className={styles.rankedPlayer}>
                        <span className={`${styles.rankedAvatar} ${styles[player.color]}`} aria-hidden="true">
                          {player.name.charAt(0).toUpperCase()}
                        </span>
                        <strong title={player.name}>{player.name}</strong>
                        {player.current && <small>YOU</small>}
                      </div>
                    </td>
                    <td>
                      <span
                        className={`${styles.playerStatus} ${
                          player.host ? styles.adminStatus : styles.memberStatus
                        }`}
                      >
                        {player.host ? "Room admin" : "Player"}
                      </span>
                    </td>
                    <td className={styles.scoreCell}>
                      <span className={styles.scoreBadge}>
                        <strong>{player.score}</strong>
                        <small>PTS</small>
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        <footer className={styles.roomNote}>
          <div>
            <span aria-hidden="true">i</span>
            <p>
              Room membership, rounds, and private cards are synced with the backend.
              Chat and scores remain local previews until their APIs are available.
            </p>
          </div>
          <span>ROOM + CARD API CONNECTED</span>
        </footer>
      </section>
    </main>
  );
}

function createInitialMessages(players: RoomPlayer[], status: "waiting" | "playing"): ChatMessage[] {
  const host = players.find((player) => player.host);
  const guest = players.find((player) => !player.host && !player.current);

  return [
    {
      id: "joined",
      sender: "Falzo",
      text: status === "waiting"
        ? "You joined the room. Say hello while everyone takes a seat."
        : "You joined during a live round. Keep secret words out of chat.",
      time: "Now",
      system: true,
    },
    ...(host ? [{
      id: "host-message",
      sender: host.name,
      text: status === "waiting"
        ? "We’ll start when everyone is ready."
        : "One clue each. Keep it short.",
      time: "2m",
    }] : []),
    ...(guest ? [{
      id: "guest-message",
      sender: guest.name,
      text: status === "waiting" ? "Ready when you are 👀" : "Who wants to go first?",
      time: "1m",
    }] : []),
  ];
}

type PlayerSeatProps = {
  player: RoomPlayer | null;
  card?: PlayerCard;
  dealPhase: DealPhase;
  dealt: boolean;
  onReveal: () => void;
};

function PlayerSeat({
  player,
  card,
  dealPhase,
  dealt,
  onReveal,
}: PlayerSeatProps) {
  if (!player) {
    return (
      <article className={`${styles.seat} ${styles.emptySeat}`}>
        <span className={styles.emptyAvatar} aria-hidden="true">+</span>
        <div>
          <strong>Open seat</strong>
          <small>Waiting for player</small>
        </div>
        <div className={styles.emptyCard} aria-hidden="true" />
      </article>
    );
  }

  const canReveal = player.current && card && dealt && dealPhase === "ready";

  return (
    <article className={`${styles.seat} ${player.current ? styles.currentSeat : ""}`}>
      <div className={styles.playerInfo}>
        <span className={`${styles.playerAvatar} ${styles[player.color]}`} aria-hidden="true">
          {player.name.charAt(0).toUpperCase()}
        </span>
        <div>
          <strong title={player.name}>{player.name}</strong>
          <small>
            {player.current
              ? player.host ? "You · Admin" : "You"
              : player.host ? "Room admin" : "Player"}
          </small>
        </div>
      </div>

      {!dealt ? (
        <div className={`${styles.playerCard} ${styles.undealtCard}`} aria-label={`${player.name} is waiting for a card`}>
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
        <div className={`${styles.playerCard} ${styles.dealtCard}`} aria-label={`${player.name} has a private card`}>
          <small>SECRET CARD</small>
          <strong>?</strong>
          <span>Private</span>
        </div>
      )}
    </article>
  );
}
