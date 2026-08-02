"use client";

import { useEffect, useState, type SubmitEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { LogoutButton } from "@/components/logout-button";
import { useSession } from "@/components/session-guard";
import { mapRoomResponse, type GameRoom } from "@/features/rooms/data";
import {
  ApiError,
  countUnreadNotifications,
  createRoom,
  joinRoom,
  listFriends,
  listRooms,
} from "@/lib/api";
import { restoreSession } from "@/lib/auth";
import type { RoomLanguage } from "@/types/room";
import type { Friend } from "@/types/social";
import styles from "./dashboard.module.css";

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession();
  const username = session.username;
  const [friends, setFriends] = useState<Friend[]>([]);
  const [friendsLoaded, setFriendsLoaded] = useState(false);
  const [unreadNotifications, setUnreadNotifications] = useState(0);
  const [rooms, setRooms] = useState<GameRoom[]>([]);
  const [roomsLoaded, setRoomsLoaded] = useState(false);
  const [roomsError, setRoomsError] = useState("");
  const [roomsReloadToken, setRoomsReloadToken] = useState(0);
  const [roomAction, setRoomAction] = useState<"create" | "join" | null>(null);
  const [roomName, setRoomName] = useState("");
  const [maxPlayers, setMaxPlayers] = useState(8);
  const [roomLanguage, setRoomLanguage] = useState<RoomLanguage>("vi");
  const [inviteCode, setInviteCode] = useState("");
  const [actionError, setActionError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const initial = username.trim().charAt(0).toUpperCase() || "P";
  const openRooms = rooms.filter((room) => room.status === "waiting").length;

  useEffect(() => {
    let active = true;

    async function loadRooms(trackActivity: boolean) {
      try {
        const activeSession = await restoreSession();
        if (!activeSession) {
          router.replace("/login");
          return;
        }
        const response = await listRooms(activeSession.access_token, { trackActivity });
        if (!active) return;
        setRooms(response.map(mapRoomResponse));
        setRoomsError("");
        setRoomsLoaded(true);
      } catch (error) {
        if (active) {
          setRoomsError(roomApiErrorMessage(error));
          setRoomsLoaded(true);
        }
      }
    }

    void loadRooms(true);
    const pollTimer = window.setInterval(() => void loadRooms(false), 5000);

    return () => {
      active = false;
      window.clearInterval(pollTimer);
    };
  }, [roomsReloadToken, router]);

  useEffect(() => {
    let active = true;

    async function loadSocialData() {
      try {
        const activeSession = await restoreSession();
        if (!activeSession) return;
        const [friendData, unreadData] = await Promise.all([
          listFriends(activeSession.access_token, { trackActivity: false }),
          countUnreadNotifications(activeSession.access_token, { trackActivity: false }),
        ]);
        if (!active) return;
        setFriends(friendData);
        setUnreadNotifications(unreadData.count);
        setFriendsLoaded(true);
      } catch {
        if (active) setFriendsLoaded(true);
      }
    }

    void loadSocialData();
    const pollTimer = window.setInterval(() => void loadSocialData(), 30_000);
    return () => {
      active = false;
      window.clearInterval(pollTimer);
    };
  }, []);

  function retryRooms() {
    setRoomsError("");
    setRoomsLoaded(false);
    setRoomsReloadToken((current) => current + 1);
  }

  function openRoomAction(action: "create" | "join") {
    setRoomAction(action);
    setActionError("");
  }

  async function handleCreateRoom(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    const name = roomName.trim();
    if (!name) {
      setActionError("Enter a room name.");
      return;
    }

    setSubmitting(true);
    setActionError("");
    try {
      const activeSession = await restoreSession();
      if (!activeSession) {
        router.replace("/login");
        return;
      }
      const response = await createRoom(activeSession.access_token, {
        name,
        maxPlayers,
        languageCode: roomLanguage,
      });
      router.push(`/rooms/${response.id}`);
    } catch (error) {
      setActionError(roomApiErrorMessage(error));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleJoinRoom(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    const code = inviteCode.trim().toUpperCase();
    if (code.length < 6) {
      setActionError("Enter a valid invite code.");
      return;
    }

    setSubmitting(true);
    setActionError("");
    try {
      const activeSession = await restoreSession();
      if (!activeSession) {
        router.replace("/login");
        return;
      }
      const response = await joinRoom(activeSession.access_token, code);
      router.push(`/rooms/${response.id}`);
    } catch (error) {
      setActionError(roomApiErrorMessage(error));
    } finally {
      setSubmitting(false);
    }
  }

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
          <Link className={styles.notificationLink} href="/friends#notifications" aria-label={`${unreadNotifications} unread friend notifications`}>
            <span aria-hidden="true">!</span>
            {unreadNotifications > 0 && <strong>{unreadNotifications > 99 ? "99+" : unreadNotifications}</strong>}
          </Link>
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
          <div className={styles.sidebarMain}>
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

            <section className={styles.friends} aria-label="Friends">
              <div className={styles.friendsHeading}>
                <p className={styles.sidebarTitle}>FRIENDS</p>
                <Link href="/friends">Manage</Link>
              </div>

              <div className={styles.friendList}>
                {friends.slice(0, 5).map((friend, index) => (
                  <Link
                    className={styles.friend}
                    href="/friends"
                    key={friend.id}
                  >
                    <span className={`${styles.friendAvatar} ${styles[friendColor(index)]}`}>
                      {friend.username.charAt(0).toUpperCase()}
                    </span>
                    <span>
                      <strong>{friend.username}</strong>
                      <small>Falzo friend</small>
                    </span>
                    <span className={styles.chatIcon} aria-hidden="true">→</span>
                  </Link>
                ))}
                {friendsLoaded && friends.length === 0 && (
                  <Link className={styles.emptyFriendList} href="/friends">+ Find your first friend</Link>
                )}
              </div>
            </section>
          </div>

          <div className={styles.sidebarNote}>
            <span aria-hidden="true">i</span>
            <div>
              <strong>Live room lobby</strong>
              <p>Rooms, chat and friend activity are synced by the server.</p>
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
            <div className={styles.roomCount}>
              <span className={styles.liveDot} aria-hidden="true" />
              <p><strong>{rooms.length} rooms</strong> available now</p>
            </div>
            <div className={styles.roomActions}>
              <button onClick={() => openRoomAction("join")} type="button">
                Join with code
              </button>
              <button className={styles.createRoomButton} onClick={() => openRoomAction("create")} type="button">
                <span aria-hidden="true">+</span> Create room
              </button>
            </div>
          </div>

          {roomAction && (
            <section className={styles.roomActionPanel} aria-labelledby="room-action-title">
              <div className={styles.actionPanelHeading}>
                <div>
                  <p>{roomAction === "create" ? "NEW ROOM" : "ROOM INVITE"}</p>
                  <h2 id="room-action-title">
                    {roomAction === "create" ? "Create an Undercover room" : "Join your friends"}
                  </h2>
                </div>
                <button
                  aria-label="Close room form"
                  className={styles.closeActionPanel}
                  onClick={() => setRoomAction(null)}
                  type="button"
                >
                  ×
                </button>
              </div>

              {roomAction === "create" ? (
                <form className={styles.roomForm} onSubmit={handleCreateRoom}>
                  <label>
                    <span>Room name</span>
                    <input
                      autoFocus
                      maxLength={80}
                      onChange={(event) => setRoomName(event.target.value)}
                      placeholder="Friday night"
                      value={roomName}
                    />
                  </label>
                  <label>
                    <span>Players</span>
                    <select
                      onChange={(event) => setMaxPlayers(Number(event.target.value))}
                      value={maxPlayers}
                    >
                      {[4, 5, 6, 7, 8, 9, 10, 11, 12].map((count) => (
                        <option key={count} value={count}>{count} players</option>
                      ))}
                    </select>
                  </label>
                  <label>
                    <span>Card language</span>
                    <select
                      onChange={(event) => setRoomLanguage(event.target.value as RoomLanguage)}
                      value={roomLanguage}
                    >
                      <option value="vi">Tiếng Việt</option>
                      <option value="en">English</option>
                    </select>
                  </label>
                  <button disabled={submitting} type="submit">
                    {submitting ? "Creating…" : "Create room"} <span aria-hidden="true">→</span>
                  </button>
                </form>
              ) : (
                <form className={styles.roomForm} onSubmit={handleJoinRoom}>
                  <label className={styles.inviteField}>
                    <span>Invite code</span>
                    <input
                      autoFocus
                      maxLength={8}
                      onChange={(event) => setInviteCode(event.target.value.toUpperCase())}
                      placeholder="OWL824"
                      value={inviteCode}
                    />
                  </label>
                  <button disabled={submitting} type="submit">
                    {submitting ? "Joining…" : "Join room"} <span aria-hidden="true">→</span>
                  </button>
                </form>
              )}

              {actionError && (
                <div className={styles.formError} role="alert">
                  <span aria-hidden="true">!</span>
                  <p>{actionError}</p>
                </div>
              )}
            </section>
          )}

          {roomsError && (
            <div className={styles.roomsError} role="alert">
              <span className={styles.errorMark} aria-hidden="true">!</span>
              <div>
                <strong>Room server unavailable</strong>
                <p>{roomsError}</p>
                <small>
                  {rooms.length > 0
                    ? "Showing the last rooms we loaded."
                    : "Your session is safe. Try again in a moment."}
                </small>
              </div>
              <button onClick={retryRooms} type="button">
                Try again <span aria-hidden="true">↻</span>
              </button>
            </div>
          )}

          {rooms.length > 0 ? <div className={styles.roomGrid}>
            {rooms.map((room) => {
              const isFull = room.players.length >= room.maxPlayers;
              return (
                <article className={styles.roomCard} key={room.id}>
                  <div className={styles.roomTopline}>
                    <span className={`${styles.status} ${styles[room.status]}`}>
                      {room.status === "waiting" ? "Waiting" : `Round ${room.round}`}
                    </span>
                    <span className={styles.roomMeta}>
                      <span className={styles.languageBadge}>
                        {room.language === "vi" ? "VI" : "EN"}
                      </span>
                      <span className={styles.roomCode}>#{room.code}</span>
                    </span>
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
                    <span className={styles.seatedCount}>
                      {room.players.length}/{room.maxPlayers} seated
                    </span>
                  </div>

                  <div className={styles.roomFooter}>
                    <span className={!isFull ? styles.openSeats : undefined}>
                      {isFull ? "Room full" : `${room.maxPlayers - room.players.length} seats open`}
                    </span>
                    <Link className={styles.viewRoomButton} href={`/rooms/${room.id}`}>
                      View room <span aria-hidden="true">→</span>
                    </Link>
                  </div>
                </article>
              );
            })}
          </div> : roomsLoaded && !roomsError && (
            <div className={styles.emptyRooms}>
              <span aria-hidden="true">?</span>
              <h2>No rooms are open yet.</h2>
              <p>Create the first room and invite your friends to take a seat.</p>
              <button onClick={() => openRoomAction("create")} type="button">Create room</button>
            </div>
          )}
        </section>
      </div>

    </main>
  );
}

function roomApiErrorMessage(error: unknown) {
  if (error instanceof ApiError) return error.message;
  return "Could not reach the room server. Please try again.";
}

function friendColor(index: number) {
  const colors = ["lime", "coral", "blue", "mint", "violet", "sand"] as const;
  return colors[index % colors.length];
}
