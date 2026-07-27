"use client";

import { useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { LogoutButton } from "@/components/logout-button";
import { useSession } from "@/components/session-guard";
import { ChatPanel, type ChatMessage } from "@/features/chat/chat-panel";
import { mapRoomResponse, type GameRoom } from "@/features/rooms/data";
import { ApiError, createRoom, joinRoom, listRooms } from "@/lib/api";
import { restoreSession } from "@/lib/auth";
import type { RoomLanguage } from "@/types/room";
import styles from "./dashboard.module.css";

type Friend = {
  id: string;
  name: string;
  status: "online" | "offline";
  activity: string;
  color: "lime" | "coral" | "blue" | "sand" | "violet" | "mint";
};

const friends: readonly Friend[] = [
  { id: "minh", name: "Minh", status: "online", activity: "In Night Owls", color: "coral" },
  { id: "lan", name: "Lan", status: "online", activity: "Looking for a room", color: "blue" },
  { id: "khoa", name: "Khoa", status: "online", activity: "In round 2", color: "mint" },
  { id: "vy", name: "Vy", status: "offline", activity: "Last seen 18m ago", color: "violet" },
  { id: "bao", name: "Bảo", status: "offline", activity: "Last seen yesterday", color: "sand" },
];

export default function DashboardPage() {
  const router = useRouter();
  const session = useSession();
  const username = session.username;
  const [selectedFriendId, setSelectedFriendId] = useState<string | null>(null);
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

  const selectedFriend = friends.find((friend) => friend.id === selectedFriendId);

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

  function retryRooms() {
    setRoomsError("");
    setRoomsLoaded(false);
    setRoomsReloadToken((current) => current + 1);
  }

  function openRoomAction(action: "create" | "join") {
    setRoomAction(action);
    setActionError("");
  }

  async function handleCreateRoom(event: FormEvent<HTMLFormElement>) {
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

  async function handleJoinRoom(event: FormEvent<HTMLFormElement>) {
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
                <span>{friends.filter((friend) => friend.status === "online").length} online</span>
              </div>

              <div className={styles.friendList}>
                {friends.map((friend) => (
                  <button
                    className={`${styles.friend} ${
                      selectedFriendId === friend.id ? styles.selectedFriend : ""
                    }`}
                    key={friend.id}
                    onClick={() => setSelectedFriendId(friend.id)}
                    type="button"
                  >
                    <span className={`${styles.friendAvatar} ${styles[friend.color]}`}>
                      {friend.name.charAt(0).toUpperCase()}
                      <span
                        className={`${styles.friendStatus} ${styles[friend.status]}`}
                        aria-label={friend.status}
                      />
                    </span>
                    <span>
                      <strong>{friend.name}</strong>
                      <small>{friend.activity}</small>
                    </span>
                    <span className={styles.chatIcon} aria-hidden="true">···</span>
                  </button>
                ))}
              </div>
            </section>
          </div>

          <div className={styles.sidebarNote}>
            <span aria-hidden="true">i</span>
            <div>
              <strong>Live room lobby</strong>
              <p>Rooms and private cards are synced by the server. Chat is still a local preview.</p>
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

      {selectedFriend && (
        <div className={styles.directChat}>
          <ChatPanel
            currentUsername={username}
            initialMessages={createFriendMessages(selectedFriend)}
            inputPlaceholder={`Message ${selectedFriend.name}…`}
            key={selectedFriend.id}
            onClose={() => setSelectedFriendId(null)}
            presence={selectedFriend.status}
            subtitle={selectedFriend.status === "online" ? selectedFriend.activity : "Offline"}
            title={selectedFriend.name}
          />
        </div>
      )}
    </main>
  );
}

function roomApiErrorMessage(error: unknown) {
  if (error instanceof ApiError) return error.message;
  return "Could not reach the room server. Please try again.";
}

function createFriendMessages(friend: Friend): ChatMessage[] {
  if (friend.status === "offline") {
    return [
      {
        id: `${friend.id}-offline`,
        sender: "Falzo",
        text: `${friend.name} is offline. Your messages will stay in this demo chat.`,
        time: "Now",
        system: true,
      },
      {
        id: `${friend.id}-last-message`,
        sender: friend.name,
        text: "Let’s play another round later.",
        time: "Yesterday",
      },
    ];
  }

  return [
    {
      id: `${friend.id}-online`,
      sender: "Falzo",
      text: `${friend.name} is online now.`,
      time: "Now",
      system: true,
    },
    {
      id: `${friend.id}-message`,
      sender: friend.name,
      text: friend.activity.startsWith("In ")
        ? "I’m already in a room. Join when you’re ready!"
        : "Want to start an Undercover room?",
      time: "1m",
    },
  ];
}
