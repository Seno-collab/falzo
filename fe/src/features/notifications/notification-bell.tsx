import { Bell } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import type { AppNotification } from "@/features/notifications/types";
import { cn } from "@/lib/utils";

function formatNotificationTime(value: string) {
  const timestamp = new Date(value).getTime();
  if (!Number.isFinite(timestamp)) {
    return "";
  }

  const diffMs = Date.now() - timestamp;
  const diffMinutes = Math.max(0, Math.floor(diffMs / 60_000));
  if (diffMinutes < 1) {
    return "now";
  }
  if (diffMinutes < 60) {
    return `${diffMinutes}m`;
  }

  const diffHours = Math.floor(diffMinutes / 60);
  if (diffHours < 24) {
    return `${diffHours}h`;
  }

  return new Intl.DateTimeFormat("en", {
    day: "2-digit",
    month: "short",
  }).format(timestamp);
}

export function NotificationBell({
  notifications,
  onOpen,
  onSelectNotification,
  unreadCount,
}: Readonly<{
  notifications: AppNotification[];
  onOpen: () => void;
  onSelectNotification?: (notification: AppNotification) => void;
  unreadCount: number;
}>) {
  const [open, setOpen] = useState(false);
  const displayCount = unreadCount > 9 ? "9+" : String(unreadCount);

  function toggleOpen() {
    const nextOpen = !open;
    setOpen(nextOpen);
    if (nextOpen) {
      onOpen();
    }
  }

  function selectNotification(notification: AppNotification) {
    onSelectNotification?.(notification);
    setOpen(false);
  }

  return (
    <div className="relative">
      <Button
        aria-expanded={open}
        aria-label="Notifications"
        className="relative rounded-full"
        onClick={toggleOpen}
        size="icon-sm"
        type="button"
        variant="ghost"
      >
        <Bell className="size-4" />
        {unreadCount > 0 ? (
          <span className="-right-1 -top-1 absolute inline-flex min-w-4 items-center justify-center rounded-full bg-[#e23b3b] px-1 text-[10px] font-bold leading-4 text-white shadow-[0_8px_18px_-10px_rgb(226_59_59/0.8)]">
            {displayCount}
          </span>
        ) : null}
      </Button>

      {open ? (
        <div className="absolute right-0 top-11 z-50 w-[min(22rem,calc(100vw-1.5rem))] overflow-hidden rounded-2xl border border-black/8 bg-white shadow-[0_26px_80px_-42px_rgb(0_0_0/0.64)]">
          <div className="flex items-center justify-between border-black/6 border-b px-4 py-3">
            <div>
              <p className="text-sm font-semibold text-[#111]">Notifications</p>
              <p className="text-xs text-[#6f7782]">
                Uploads and comment messages
              </p>
            </div>
            {unreadCount > 0 ? (
              <span className="rounded-full bg-[#111] px-2 py-1 text-[11px] font-semibold text-white">
                {unreadCount}
              </span>
            ) : null}
          </div>

          <div className="max-h-96 overflow-y-auto p-2">
            {notifications.length > 0 ? (
              notifications.map((notification, index) => (
                <button
                  aria-label={`Open notification: ${notification.title}`}
                  className={cn(
                    "block w-full rounded-xl px-3 py-2.5 text-left transition",
                    notification.post_id
                      ? "cursor-pointer"
                      : "cursor-default",
                    index === 0 ? "bg-[#f3f6f8]" : "hover:bg-[#f7f7f5]",
                  )}
                  key={notification.id}
                  onClick={() => selectNotification(notification)}
                  type="button"
                >
                  <div className="flex items-start justify-between gap-3">
                    <p className="min-w-0 text-sm font-semibold text-[#171717]">
                      {notification.title}
                    </p>
                    <span className="shrink-0 text-[11px] text-[#7b8490]">
                      {formatNotificationTime(notification.created_at)}
                    </span>
                  </div>
                  <p className="mt-1 line-clamp-2 text-sm text-[#4a5664]">
                    {notification.body}
                  </p>
                </button>
              ))
            ) : (
              <div className="px-4 py-8 text-center">
                <p className="text-sm font-semibold text-[#1f1f1f]">
                  No notifications yet
                </p>
                <p className="mt-1 text-xs text-[#7b8490]">
                  New uploads and comments will appear here.
                </p>
              </div>
            )}
          </div>
        </div>
      ) : null}
    </div>
  );
}
