"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { useSession } from "@/components/session-guard";
import {
  acceptFriendRequest,
  ApiError,
  cancelFriendRequest,
  countUnreadNotifications,
  listFriendRequests,
  listFriends,
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
  rejectFriendRequest,
  searchUsers,
  sendFriendRequest,
  unfriend,
} from "@/lib/api";
import { restoreSession } from "@/lib/auth";
import type {
  Friend,
  FriendNotification,
  FriendRequest,
  SocialUser,
} from "@/types/social";
import styles from "./friends.module.css";

type Panel = "friends" | "requests" | "find" | "notifications";

export default function FriendsPage() {
  const router = useRouter();
  const session = useSession();
  const [panel, setPanel] = useState<Panel>("friends");
  const [friends, setFriends] = useState<Friend[]>([]);
  const [requests, setRequests] = useState<FriendRequest[]>([]);
  const [notifications, setNotifications] = useState<FriendNotification[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SocialUser[]>([]);
  const [searched, setSearched] = useState(false);
  const [loading, setLoading] = useState(true);
  const [searching, setSearching] = useState(false);
  const [pendingAction, setPendingAction] = useState("");
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");

  const incomingRequests = useMemo(
    () => requests.filter((request) => request.receiver_name === session.username),
    [requests, session.username],
  );
  const outgoingRequests = useMemo(
    () => requests.filter((request) => request.sender_name === session.username),
    [requests, session.username],
  );

  const loadSocialData = useCallback(async (trackActivity: boolean) => {
    const activeSession = await restoreSession();
    if (!activeSession) {
      router.replace("/login");
      return false;
    }
    const [friendData, requestData, notificationData, unreadData] = await Promise.all([
      listFriends(activeSession.access_token, { trackActivity }),
      listFriendRequests(activeSession.access_token, { trackActivity }),
      listNotifications(activeSession.access_token, {}, { trackActivity }),
      countUnreadNotifications(activeSession.access_token, { trackActivity }),
    ]);
    setFriends(friendData);
    setRequests(requestData);
    setNotifications(notificationData);
    setUnreadCount(unreadData.count);
    return true;
  }, [router]);

  useEffect(() => {
    let active = true;
    void loadSocialData(true)
      .catch((loadError) => {
        if (active) setError(socialErrorMessage(loadError));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [loadSocialData]);

  useEffect(() => {
    if (window.location.hash === "#notifications") setPanel("notifications");
  }, []);

  async function activeAccessToken() {
    const activeSession = await restoreSession();
    if (!activeSession) {
      router.replace("/login");
      throw new Error("Session expired");
    }
    return activeSession.access_token;
  }

  async function runAction(key: string, action: (accessToken: string) => Promise<unknown>, success: string) {
    setPendingAction(key);
    setError("");
    setNotice("");
    try {
      const accessToken = await activeAccessToken();
      await action(accessToken);
      await loadSocialData(false);
      setNotice(success);
      return true;
    } catch (actionError) {
      setError(socialErrorMessage(actionError));
      return false;
    } finally {
      setPendingAction("");
    }
  }

  async function handleSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const value = query.trim();
    if (value.length < 2) {
      setError("Enter at least 2 characters to find a player.");
      return;
    }
    setSearching(true);
    setError("");
    setNotice("");
    try {
      const accessToken = await activeAccessToken();
      setResults(await searchUsers(accessToken, value));
      setSearched(true);
    } catch (searchError) {
      setError(socialErrorMessage(searchError));
    } finally {
      setSearching(false);
    }
  }

  async function handleSendRequest(user: SocialUser) {
    const sent = await runAction(
      `send-${user.id}`,
      (accessToken) => sendFriendRequest(accessToken, user.id),
      `Friend request sent to ${user.username}.`,
    );
    if (sent) {
      setResults((current) => current.map((item) => (
        item.id === user.id ? { ...item, relationship: "OUTGOING_REQUEST" } : item
      )));
    }
  }

  async function handleUnfriend(friend: Friend) {
    if (!window.confirm(`Remove ${friend.username} from your friends?`)) return;
    await runAction(
      `unfriend-${friend.id}`,
      (accessToken) => unfriend(accessToken, friend.id),
      `${friend.username} was removed from your friends.`,
    );
  }

  async function openNotification(notification: FriendNotification) {
    setError("");
    if (!notification.read) {
      setPendingAction(`notification-${notification.id}`);
      try {
        const accessToken = await activeAccessToken();
        await markNotificationRead(accessToken, notification.id);
        const readAt = new Date().toISOString();
        setNotifications((current) => current.map((item) => (
          item.id === notification.id ? { ...item, read: true, read_at: readAt } : item
        )));
        setUnreadCount((current) => Math.max(0, current - 1));
      } catch (notificationError) {
        setError(socialErrorMessage(notificationError));
        return;
      } finally {
        setPendingAction("");
      }
    }
    setPanel(notification.type === "FRIEND_REQUEST_RECEIVED" ? "requests" : "friends");
  }

  async function handleMarkAllRead() {
    await runAction(
      "notifications-read-all",
      (accessToken) => markAllNotificationsRead(accessToken),
      "All notifications marked as read.",
    );
  }

  return (
    <main className={styles.page}>
      <header className={styles.header}>
        <Link className={styles.brand} href="/dashboard">
          <span aria-hidden="true">F</span>
          falzo
        </Link>
        <Link className={styles.backLink} href="/dashboard">← Back to rooms</Link>
      </header>

      <div className={styles.shell}>
        <section className={styles.intro}>
          <div>
            <p>SOCIAL</p>
            <h1>Friends</h1>
            <span>Find players and bring your group into the next room.</span>
          </div>
          <div className={styles.summary}>
            <span><strong>{friends.length}</strong> friends</span>
            <span><strong>{incomingRequests.length}</strong> requests</span>
            <span><strong>{unreadCount}</strong> unread</span>
          </div>
        </section>

        <nav className={styles.tabs} aria-label="Friend sections">
          <Tab active={panel === "friends"} label="Friends" onClick={() => setPanel("friends")} />
          <Tab
            active={panel === "requests"}
            count={incomingRequests.length}
            label="Requests"
            onClick={() => setPanel("requests")}
          />
          <Tab active={panel === "find"} label="Find players" onClick={() => setPanel("find")} />
          <Tab
            active={panel === "notifications"}
            count={unreadCount}
            label="Notifications"
            onClick={() => setPanel("notifications")}
          />
        </nav>

        {error && <p className={styles.error} role="alert">{error}</p>}
        {notice && <p className={styles.notice} role="status">{notice}</p>}

        <section className={styles.panel} aria-live="polite">
          {loading ? (
            <EmptyState title="Loading friends…" text="Syncing your social list with the server." />
          ) : panel === "friends" ? (
            <FriendsPanel
              friends={friends}
              onUnfriend={handleUnfriend}
              pendingAction={pendingAction}
              showFind={() => setPanel("find")}
            />
          ) : panel === "requests" ? (
            <RequestsPanel
              incoming={incomingRequests}
              outgoing={outgoingRequests}
              pendingAction={pendingAction}
              runAction={runAction}
            />
          ) : panel === "find" ? (
            <FindPanel
              handleSearch={handleSearch}
              onSend={handleSendRequest}
              pendingAction={pendingAction}
              query={query}
              results={results}
              searched={searched}
              searching={searching}
              setPanel={setPanel}
              setQuery={setQuery}
            />
          ) : (
            <NotificationsPanel
              notifications={notifications}
              onMarkAll={handleMarkAllRead}
              onOpen={openNotification}
              pendingAction={pendingAction}
              unreadCount={unreadCount}
            />
          )}
        </section>
      </div>
    </main>
  );
}

function Tab({ active, count = 0, label, onClick }: {
  active: boolean;
  count?: number;
  label: string;
  onClick: () => void;
}) {
  return (
    <button className={active ? styles.activeTab : ""} onClick={onClick} type="button">
      {label}
      {count > 0 && <span>{count > 99 ? "99+" : count}</span>}
    </button>
  );
}

function FriendsPanel({ friends, onUnfriend, pendingAction, showFind }: {
  friends: Friend[];
  onUnfriend: (friend: Friend) => void;
  pendingAction: string;
  showFind: () => void;
}) {
  if (friends.length === 0) {
    return <EmptyState action="Find players" onAction={showFind} title="No friends yet" text="Search by username and send your first request." />;
  }
  return (
    <div className={styles.list}>
      {friends.map((friend) => (
        <article className={styles.row} key={friend.id}>
          <Avatar name={friend.username} />
          <div className={styles.person}>
            <strong>{friend.username}</strong>
            <span>Friends since {formatDate(friend.friends_at)}</span>
          </div>
          <button
            className={styles.dangerButton}
            disabled={Boolean(pendingAction)}
            onClick={() => void onUnfriend(friend)}
            type="button"
          >
            {pendingAction === `unfriend-${friend.id}` ? "Removing…" : "Unfriend"}
          </button>
        </article>
      ))}
    </div>
  );
}

function RequestsPanel({ incoming, outgoing, pendingAction, runAction }: {
  incoming: FriendRequest[];
  outgoing: FriendRequest[];
  pendingAction: string;
  runAction: (key: string, action: (accessToken: string) => Promise<unknown>, success: string) => Promise<boolean>;
}) {
  if (incoming.length === 0 && outgoing.length === 0) {
    return <EmptyState title="No pending requests" text="New friend requests will appear here." />;
  }
  return (
    <div className={styles.requestGroups}>
      <div>
        <h2>Received</h2>
        {incoming.length === 0 ? <p className={styles.muted}>No incoming requests.</p> : (
          <div className={styles.list}>
            {incoming.map((request) => (
              <article className={styles.row} key={request.id}>
                <Avatar name={request.sender_name} />
                <div className={styles.person}>
                  <strong>{request.sender_name}</strong>
                  <span>Sent {formatDate(request.created_at)}</span>
                </div>
                <div className={styles.actions}>
                  <button
                    disabled={Boolean(pendingAction)}
                    onClick={() => void runAction(
                      `accept-${request.id}`,
                      (token) => acceptFriendRequest(token, request.id),
                      `You and ${request.sender_name} are now friends.`,
                    )}
                    type="button"
                  >
                    {pendingAction === `accept-${request.id}` ? "Accepting…" : "Accept"}
                  </button>
                  <button
                    className={styles.secondaryButton}
                    disabled={Boolean(pendingAction)}
                    onClick={() => void runAction(
                      `reject-${request.id}`,
                      (token) => rejectFriendRequest(token, request.id),
                      `Request from ${request.sender_name} declined.`,
                    )}
                    type="button"
                  >
                    Decline
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}
      </div>
      <div>
        <h2>Sent</h2>
        {outgoing.length === 0 ? <p className={styles.muted}>No outgoing requests.</p> : (
          <div className={styles.list}>
            {outgoing.map((request) => (
              <article className={styles.row} key={request.id}>
                <Avatar name={request.receiver_name} />
                <div className={styles.person}>
                  <strong>{request.receiver_name}</strong>
                  <span>Waiting for a response</span>
                </div>
                <button
                  className={styles.secondaryButton}
                  disabled={Boolean(pendingAction)}
                  onClick={() => void runAction(
                    `cancel-${request.id}`,
                    (token) => cancelFriendRequest(token, request.id),
                    `Request to ${request.receiver_name} canceled.`,
                  )}
                  type="button"
                >
                  {pendingAction === `cancel-${request.id}` ? "Canceling…" : "Cancel request"}
                </button>
              </article>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function FindPanel({
  handleSearch,
  onSend,
  pendingAction,
  query,
  results,
  searched,
  searching,
  setPanel,
  setQuery,
}: {
  handleSearch: (event: FormEvent<HTMLFormElement>) => void;
  onSend: (user: SocialUser) => void;
  pendingAction: string;
  query: string;
  results: SocialUser[];
  searched: boolean;
  searching: boolean;
  setPanel: (panel: Panel) => void;
  setQuery: (query: string) => void;
}) {
  return (
    <div>
      <form className={styles.search} onSubmit={handleSearch}>
        <input
          aria-label="Player username"
          autoFocus
          maxLength={100}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search username…"
          value={query}
        />
        <button disabled={searching} type="submit">{searching ? "Searching…" : "Find player"}</button>
      </form>
      {searched && results.length === 0 ? (
        <EmptyState title="No players found" text="Try the beginning of another username." />
      ) : (
        <div className={styles.list}>
          {results.map((user) => (
            <article className={styles.row} key={user.id}>
              <Avatar name={user.username} />
              <div className={styles.person}>
                <strong>{user.username}</strong>
                <span>{relationshipLabel(user.relationship)}</span>
              </div>
              {user.relationship === "NONE" ? (
                <button
                  disabled={Boolean(pendingAction)}
                  onClick={() => void onSend(user)}
                  type="button"
                >
                  {pendingAction === `send-${user.id}` ? "Sending…" : "Add friend"}
                </button>
              ) : user.relationship === "INCOMING_REQUEST" ? (
                <button className={styles.secondaryButton} onClick={() => setPanel("requests")} type="button">View request</button>
              ) : (
                <span className={styles.relationshipTag}>{relationshipLabel(user.relationship)}</span>
              )}
            </article>
          ))}
        </div>
      )}
    </div>
  );
}

function NotificationsPanel({ notifications, onMarkAll, onOpen, pendingAction, unreadCount }: {
  notifications: FriendNotification[];
  onMarkAll: () => void;
  onOpen: (notification: FriendNotification) => void;
  pendingAction: string;
  unreadCount: number;
}) {
  if (notifications.length === 0) {
    return <EmptyState title="No notifications" text="Friend activity will appear here." />;
  }
  return (
    <div>
      <div className={styles.notificationTools}>
        <span>{unreadCount} unread</span>
        {unreadCount > 0 && (
          <button disabled={Boolean(pendingAction)} onClick={() => void onMarkAll()} type="button">
            {pendingAction === "notifications-read-all" ? "Marking…" : "Mark all read"}
          </button>
        )}
      </div>
      <div className={styles.list}>
        {notifications.map((notification) => (
          <button
            className={`${styles.notification} ${notification.read ? "" : styles.unread}`}
            disabled={pendingAction === `notification-${notification.id}`}
            key={notification.id}
            onClick={() => void onOpen(notification)}
            type="button"
          >
            <span className={styles.notificationMark} aria-hidden="true">
              {notification.type === "FRIEND_REQUEST_RECEIVED" ? "+" : "✓"}
            </span>
            <span>
              <strong>{notificationText(notification)}</strong>
              <small>{formatDate(notification.created_at)}</small>
            </span>
            {!notification.read && <span className={styles.unreadDot} aria-label="Unread" />}
          </button>
        ))}
      </div>
    </div>
  );
}

function Avatar({ name }: { name: string }) {
  return <span className={styles.avatar} aria-hidden="true">{name.trim().charAt(0).toUpperCase() || "?"}</span>;
}

function EmptyState({ action, onAction, text, title }: {
  action?: string;
  onAction?: () => void;
  text: string;
  title: string;
}) {
  return (
    <div className={styles.empty}>
      <span aria-hidden="true">+</span>
      <h2>{title}</h2>
      <p>{text}</p>
      {action && onAction && <button onClick={onAction} type="button">{action}</button>}
    </div>
  );
}

function relationshipLabel(status: SocialUser["relationship"]) {
  switch (status) {
    case "FRIENDS": return "Already friends";
    case "INCOMING_REQUEST": return "Sent you a request";
    case "OUTGOING_REQUEST": return "Request sent";
    default: return "Player";
  }
}

function notificationText(notification: FriendNotification) {
  return notification.type === "FRIEND_REQUEST_RECEIVED"
    ? `${notification.actor_name} sent you a friend request`
    : `${notification.actor_name} accepted your friend request`;
}

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Recently";
  return new Intl.DateTimeFormat("en", {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function socialErrorMessage(error: unknown) {
  if (error instanceof ApiError) return error.message;
  if (error instanceof Error && error.message === "Session expired") return "Your session expired. Please sign in again.";
  return "Could not reach the friend service. Please try again.";
}
