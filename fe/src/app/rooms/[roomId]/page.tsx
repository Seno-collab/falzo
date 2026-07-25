"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { LogoutButton } from "@/components/logout-button";
import {
  dealCards,
  getRoomForPlayer,
  type PlayerCard,
  type RoomPlayer,
} from "@/features/rooms/data";
import { getSession } from "@/lib/auth";
import styles from "./room.module.css";

export default function RoomDetailPage() {
  const params = useParams<{ roomId: string }>();
  const router = useRouter();
  const [username, setUsername] = useState<string | null>(null);
  const [cards, setCards] = useState<PlayerCard[]>([]);
  const [cardRevealed, setCardRevealed] = useState(false);

  const room = useMemo(
    () => (username ? getRoomForPlayer(params.roomId, username) : undefined),
    [params.roomId, username],
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
      setCards(dealCards(room.players));
      setCardRevealed(false);
    }
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
  const splitAt = Math.ceil(room.maxPlayers / 2);
  const seats = Array.from<RoomPlayer | null>({ length: room.maxPlayers }).fill(null);
  room.players.forEach((player, index) => {
    seats[index] = player;
  });

  function dealAgain() {
    if (!room) return;
    setCards(dealCards(room.players));
    setCardRevealed(false);
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

          <button className={styles.dealButton} onClick={dealAgain} type="button">
            <span aria-hidden="true">↻</span>
            Deal demo cards
          </button>
        </div>

        <section className={styles.board} aria-label={`${room.name} seating table`}>
          <div className={styles.seatRow}>
            {seats.slice(0, splitAt).map((player, index) => (
              <PlayerSeat
                card={cards.find((card) => card.playerId === player?.id)}
                cardRevealed={cardRevealed}
                key={player?.id ?? `empty-top-${index}`}
                onReveal={() => setCardRevealed((visible) => !visible)}
                player={player}
              />
            ))}
          </div>

          <div className={styles.table}>
            <div className={styles.tableMark} aria-hidden="true">?</div>
            <p>UNDERCOVER</p>
            <h2>Read the room.</h2>
            <span>
              Every seated player has one card. Your card stays private until you reveal it.
            </span>
            <div className={styles.tableStatus}>
              <span>{cards.length} cards dealt</span>
              <span>{openSeats} open seats</span>
            </div>
          </div>

          <div className={styles.seatRow}>
            {seats.slice(splitAt).map((player, index) => (
              <PlayerSeat
                card={cards.find((card) => card.playerId === player?.id)}
                cardRevealed={cardRevealed}
                key={player?.id ?? `empty-bottom-${index}`}
                onReveal={() => setCardRevealed((visible) => !visible)}
                player={player}
              />
            ))}
          </div>
        </section>

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

type PlayerSeatProps = {
  player: RoomPlayer | null;
  card?: PlayerCard;
  cardRevealed: boolean;
  onReveal: () => void;
};

function PlayerSeat({ player, card, cardRevealed, onReveal }: PlayerSeatProps) {
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

  const canReveal = player.current && card;
  const showCard = Boolean(canReveal && cardRevealed);

  return (
    <article className={`${styles.seat} ${player.current ? styles.currentSeat : ""}`}>
      <div className={styles.playerInfo}>
        <span className={`${styles.playerAvatar} ${styles[player.color]}`} aria-hidden="true">
          {player.name.charAt(0).toUpperCase()}
        </span>
        <div>
          <strong>{player.name}</strong>
          <small>{player.current ? "You" : player.host ? "Host" : "Player"}</small>
        </div>
      </div>

      {canReveal ? (
        <button
          aria-label={showCard ? "Hide your secret card" : "Reveal your secret card"}
          className={`${styles.playerCard} ${showCard ? styles.revealedCard : ""}`}
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
        <div className={styles.playerCard} aria-label={`${player.name} has a private card`}>
          <small>SECRET CARD</small>
          <strong>?</strong>
          <span>Private</span>
        </div>
      )}
    </article>
  );
}
