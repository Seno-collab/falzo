"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { LogoutButton } from "@/components/logout-button";
import { getRoomsForPlayer } from "@/features/rooms/data";
import { getSession } from "@/lib/auth";
import styles from "./dashboard.module.css";

export default function DashboardPage() {
  const router = useRouter();
  const [username, setUsername] = useState<string | null>(null);

  useEffect(() => {
    const session = getSession();
    if (!session) {
      router.replace("/login");
      return;
    }
    setUsername(session.username);
  }, [router]);

  const rooms = useMemo(
    () => (username ? getRoomsForPlayer(username) : []),
    [username],
  );

  if (!username) {
    return <main className={styles.loading}>Loading rooms…</main>;
  }

  const initial = username.trim().charAt(0).toUpperCase() || "P";
  const openRooms = rooms.filter((room) => room.status === "waiting").length;

  return (
    <main className={styles.page}>
      <header className={styles.header}>
        <Link className={styles.brand} href="/">
          <span aria-hidden="true">F</span>
          falzo
        </Link>

        <div className={styles.lobbyLabel}>
          <span aria-hidden="true" />
          Undercover rooms
        </div>

        <div className={styles.account}>
          <span className={styles.avatar} aria-hidden="true">
            {initial}
          </span>
          <div className={styles.accountName}>
            <small>PLAYER</small>
            <span>{username}</span>
          </div>
          <LogoutButton />
        </div>
      </header>

      <div className={styles.workspace}>
        <aside className={styles.sidebar}>
          <div>
            <p className={styles.sidebarTitle}>YOUR GAMES</p>
            <div className={styles.gameTab}>
              <span className={styles.tabMark} aria-hidden="true">?</span>
              <span>
                <strong>Undercover</strong>
                <small>Social deduction</small>
              </span>
              <span className={styles.activeDot} aria-label="Selected" />
            </div>
          </div>

          <div className={styles.sidebarNote}>
            <span aria-hidden="true">i</span>
            <div>
              <strong>Frontend room preview</strong>
              <p>Rooms and cards are local demo data until the game API is ready.</p>
            </div>
          </div>
        </aside>

        <section className={styles.content}>
          <div className={styles.pageHeading}>
            <div>
              <p>WELCOME BACK, {username.toUpperCase()}</p>
              <h1>Choose a room.<br />Take a seat.</h1>
            </div>
            <div className={styles.roomSummary}>
              <strong>{openRooms}</strong>
              <span>rooms waiting</span>
            </div>
          </div>

          <div className={styles.roomToolbar}>
            <div>
              <span className={styles.liveDot} aria-hidden="true" />
              <p><strong>{rooms.length} rooms</strong> available in this preview</p>
            </div>
            <span>Invite code and create-room actions will connect to the backend later.</span>
          </div>

          <div className={styles.roomGrid}>
            {rooms.map((room) => {
              const isFull = room.players.length >= room.maxPlayers;
              return (
                <article className={styles.roomCard} key={room.id}>
                  <div className={styles.roomTopline}>
                    <span className={`${styles.status} ${styles[room.status]}`}>
                      {room.status === "waiting" ? "Waiting" : `Round ${room.round}`}
                    </span>
                    <span className={styles.roomCode}>#{room.code}</span>
                  </div>

                  <div>
                    <p className={styles.roomLabel}>UNDERCOVER ROOM</p>
                    <h2>{room.name}</h2>
                  </div>

                  <div className={styles.players}>
                    <div className={styles.playerStack} aria-label={`${room.players.length} players`}>
                      {room.players.slice(0, 5).map((player) => (
                        <span
                          className={`${styles.miniAvatar} ${styles[player.color]}`}
                          key={player.id}
                          title={player.name}
                        >
                          {player.name.charAt(0).toUpperCase()}
                        </span>
                      ))}
                      {room.players.length > 5 && (
                        <span className={styles.morePlayers}>+{room.players.length - 5}</span>
                      )}
                    </div>
                    <span>{room.players.length}/{room.maxPlayers} seated</span>
                  </div>

                  <div className={styles.roomFooter}>
                    <span>{isFull ? "Room full" : `${room.maxPlayers - room.players.length} seats open`}</span>
                    <Link href={`/rooms/${room.id}`}>
                      View room <span aria-hidden="true">→</span>
                    </Link>
                  </div>
                </article>
              );
            })}
          </div>
        </section>
      </div>
    </main>
  );
}
