"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { LogoutButton } from "@/components/logout-button";
import { ChatPanel, type ChatMessage } from "@/features/chat/chat-panel";
import {
  dealCards,
  getRoomForPlayer,
  type PlayerCard,
  type RoomPlayer,
} from "@/features/rooms/data";
import { getSession } from "@/lib/auth";
import styles from "./room.module.css";

type DealPhase = "waiting" | "dealing" | "ready";

export default function RoomDetailPage() {
  const params = useParams<{ roomId: string }>();
  const router = useRouter();
  const [username, setUsername] = useState<string | null>(null);
  const [cards, setCards] = useState<PlayerCard[]>([]);
  const [cardRevealed, setCardRevealed] = useState(false);
  const [dealPhase, setDealPhase] = useState<DealPhase>("waiting");
  const [dealtPlayerIds, setDealtPlayerIds] = useState<string[]>([]);
  const [activeDealPlayerId, setActiveDealPlayerId] = useState<string | null>(null);
  const [showNextRoundConfirm, setShowNextRoundConfirm] = useState(false);
  const dealTimersRef = useRef<number[]>([]);

  const room = useMemo(
    () => (username ? getRoomForPlayer(params.roomId, username) : undefined),
    [params.roomId, username],
  );
  const initialRoomMessages = useMemo(
    () => (room ? createInitialMessages(room.players, room.status) : []),
    [room],
  );

  useEffect(() => {
    const session = getSession();
    if (!session) {
      router.replace("/login");
      return;
    }
    setUsername(session.username);
  }, [router]);

  useEffect(() => {
    if (room) {
      dealTimersRef.current.forEach((timer) => window.clearTimeout(timer));
      dealTimersRef.current = [];
      setCards([]);
      setCardRevealed(false);
      setDealPhase("waiting");
      setDealtPlayerIds([]);
      setActiveDealPlayerId(null);
      setShowNextRoundConfirm(false);
    }

    return () => {
      dealTimersRef.current.forEach((timer) => window.clearTimeout(timer));
      dealTimersRef.current = [];
    };
  }, [room]);

  if (!username) {
    return <main className={styles.loading}>Preparing the room…</main>;
  }

  if (!room) {
    return (
      <main className={styles.notFound}>
        <span aria-hidden="true">?</span>
        <h1>Room not found</h1>
        <p>This demo room may have closed or the link is incorrect.</p>
        <Link href="/dashboard">Back to rooms</Link>
      </main>
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

  function startGame() {
    if (!room || !isRoomAdmin || dealPhase === "dealing") return;

    dealTimersRef.current.forEach((timer) => window.clearTimeout(timer));
    dealTimersRef.current = [];
    setShowNextRoundConfirm(false);
    setCards(dealCards(room.players));
    setCardRevealed(false);
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
    }, firstDealDelay + room.players.length * dealStepDuration);
    dealTimersRef.current.push(completionTimer);
  }

  function handleGameAction() {
    if (!isRoomAdmin || dealPhase === "dealing") return;
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
            </p>
          </div>

          <div className={styles.gameControl}>
            <button
              aria-describedby={!isRoomAdmin ? "admin-action-hint" : undefined}
              className={styles.dealButton}
              disabled={dealPhase === "dealing" || !isRoomAdmin}
              onClick={handleGameAction}
              title={!isRoomAdmin ? "Only the room admin can control rounds" : undefined}
              type="button"
            >
              <span aria-hidden="true">
                {!isRoomAdmin
                  ? "◇"
                  : dealPhase === "waiting"
                    ? "▶"
                    : dealPhase === "dealing"
                      ? "…"
                      : "↻"}
              </span>
              {dealPhase === "waiting"
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
                  cardRevealed={cardRevealed}
                  dealPhase={dealPhase}
                  dealt={Boolean(player && dealtPlayerIds.includes(player.id))}
                  key={player?.id ?? `empty-top-${index}`}
                  onReveal={() => setCardRevealed((visible) => !visible)}
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
                  cardRevealed={cardRevealed}
                  dealPhase={dealPhase}
                  dealt={Boolean(player && dealtPlayerIds.includes(player.id))}
                  key={player?.id ?? `empty-bottom-${index}`}
                  onReveal={() => setCardRevealed((visible) => !visible)}
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
                aria-label={cardRevealed ? "Hide your secret card" : "Reveal your secret card"}
                className={`${styles.myCard} ${
                  currentCardDealt ? styles.myCardReady : ""
                } ${cardRevealed ? styles.myCardRevealed : ""}`}
                disabled={!currentCard || !currentCardDealt || dealPhase !== "ready"}
                onClick={() => setCardRevealed((visible) => !visible)}
                type="button"
              >
                {!currentCardDealt || !currentCard ? (
                  <>
                    <small>{dealPhase === "dealing" ? "INCOMING" : "NO CARD"}</small>
                    <strong>—</strong>
                    <span>{dealPhase === "dealing" ? "Please wait" : "Start game"}</span>
                  </>
                ) : cardRevealed ? (
                  <>
                    <small>{currentCard.role}</small>
                    <strong>{currentCard.word}</strong>
                    <span>Tap to hide</span>
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

        <footer className={styles.roomNote}>
          <div>
            <span aria-hidden="true">i</span>
            <p>
              This is frontend-only room state. The backend must own room membership,
              card assignment, and secret words before real multiplayer play.
            </p>
          </div>
          <span>UI PREVIEW</span>
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
  cardRevealed: boolean;
  dealPhase: DealPhase;
  dealt: boolean;
  onReveal: () => void;
};

function PlayerSeat({
  player,
  card,
  cardRevealed,
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
  const showCard = Boolean(canReveal && cardRevealed);

  return (
    <article className={`${styles.seat} ${player.current ? styles.currentSeat : ""}`}>
      <div className={styles.playerInfo}>
        <span className={`${styles.playerAvatar} ${styles[player.color]}`} aria-hidden="true">
          {player.name.charAt(0).toUpperCase()}
        </span>
        <div>
          <strong>{player.name}</strong>
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
          aria-label={showCard ? "Hide your secret card" : "Reveal your secret card"}
          className={`${styles.playerCard} ${styles.dealtCard} ${showCard ? styles.revealedCard : ""}`}
          onClick={onReveal}
          type="button"
        >
          {showCard ? (
            <>
              <small>{card.role}</small>
              <strong>{card.word}</strong>
              <span>Tap to hide</span>
            </>
          ) : (
            <>
              <small>YOUR CARD</small>
              <strong>?</strong>
              <span>Tap to reveal</span>
            </>
          )}
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
