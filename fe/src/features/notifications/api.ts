import { getAuthAccessToken, initializeAuthHeader } from "@/features/auth/api";
import type { AppNotification } from "@/features/notifications/types";
import { buildApiUrl, NOTIFICATIONS_ENDPOINT } from "@/lib/api-config";
import { apiGet, apiPost, endpointPath } from "@/lib/api-utils";

function parseNotificationPayload(data: string): AppNotification | null {
  try {
    const payload = JSON.parse(data) as Partial<AppNotification>;
    if (
      typeof payload.id !== "string" ||
      typeof payload.type !== "string" ||
      typeof payload.title !== "string" ||
      typeof payload.body !== "string" ||
      typeof payload.created_at !== "string"
    ) {
      return null;
    }

    return payload as AppNotification;
  } catch {
    return null;
  }
}

function parseSSEBlock(block: string) {
  let event = "message";
  const data: string[] = [];

  for (const line of block.split("\n")) {
    if (!line || line.startsWith(":")) {
      continue;
    }

    const separatorIndex = line.indexOf(":");
    const field = separatorIndex >= 0 ? line.slice(0, separatorIndex) : line;
    const value =
      separatorIndex >= 0 ? line.slice(separatorIndex + 1).trimStart() : "";

    if (field === "event") {
      event = value;
    }
    if (field === "data") {
      data.push(value);
    }
  }

  return { event, data: data.join("\n") };
}

function trimText(value: unknown) {
  return typeof value === "string" ? value.trim() : "";
}

export function cleanNotificationIds(ids: unknown[]) {
  return Array.from(
    new Set(
      ids
        .map((id) => {
          if (typeof id === "string") {
            return id.trim();
          }
          if (typeof id === "number" && Number.isFinite(id)) {
            return String(id);
          }
          return "";
        })
        .filter(Boolean),
    ),
  );
}

export function createPostUploadNotification(post: {
  id: number;
  user_id: number;
  user_name?: string | null;
  caption?: string | null;
  location_name?: string | null;
  created_at: string;
}): AppNotification {
  const actor = trimText(post.user_name) || "Someone";
  const detail = trimText(post.caption) || trimText(post.location_name);

  return {
    id: `post.created:${post.id}`,
    actor_name: actor,
    actor_user_id: post.user_id,
    body: detail ? `${actor} uploaded ${detail}.` : `${actor} uploaded a new post.`,
    created_at: post.created_at,
    post_id: post.id,
    resource: "post",
    resource_id: String(post.id),
    title: "New upload",
    type: "post.created",
  };
}

export function getNotificationsApi(limit = 30) {
  initializeAuthHeader();
  return apiGet<AppNotification[]>(
    `${NOTIFICATIONS_ENDPOINT}?limit=${encodeURIComponent(String(limit))}`,
  );
}

export async function markNotificationsReadApi(ids: string[]) {
  initializeAuthHeader();
  const cleanIds = cleanNotificationIds(ids);
  if (cleanIds.length === 0) {
    return;
  }

  await apiPost(endpointPath(NOTIFICATIONS_ENDPOINT, "read"), {
    ids: cleanIds,
  });
}

export function subscribeNotificationEvents({
  onNotification,
  onError,
}: {
  onNotification: (notification: AppNotification) => void;
  onError?: (error: unknown) => void;
}) {
  initializeAuthHeader();

  const token = getAuthAccessToken();
  const controller = new AbortController();
  if (!token) {
    return () => controller.abort();
  }

  void (async () => {
    const response = await fetch(buildApiUrl(`${NOTIFICATIONS_ENDPOINT}/events`), {
      headers: {
        Accept: "text/event-stream",
        Authorization: `Bearer ${token}`,
      },
      signal: controller.signal,
    });

    if (!response.ok || !response.body) {
      throw new Error("Notification stream is unavailable.");
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";

    for (;;) {
      const { done, value } = await reader.read();
      if (done) {
        break;
      }

      buffer += decoder.decode(value, { stream: true });
      let boundaryIndex = buffer.indexOf("\n\n");
      while (boundaryIndex >= 0) {
        const block = buffer.slice(0, boundaryIndex);
        buffer = buffer.slice(boundaryIndex + 2);
        boundaryIndex = buffer.indexOf("\n\n");

        const event = parseSSEBlock(block);
        if (event.event !== "notification.created") {
          continue;
        }

        const notification = parseNotificationPayload(event.data);
        if (notification) {
          onNotification(notification);
        }
      }
    }
  })().catch((error) => {
    if (!controller.signal.aborted) {
      onError?.(error);
    }
  });

  return () => controller.abort();
}
