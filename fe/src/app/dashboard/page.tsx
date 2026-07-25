"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { LogoutButton } from "@/components/logout-button";
import { ChatPanel, type ChatMessage } from "@/features/chat/chat-panel";
import { getRoomsForPlayer } from "@/features/rooms/data";
import { getSession } from "@/lib/auth";
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
  const [username, setUsername] = useState<string | null>(null);
  const [selectedFriendId, setSelectedFriendId] = useState<string | null>(null);

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
  const selectedFriend = friends.find((friend) => friend.id === selectedFriendId);

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
