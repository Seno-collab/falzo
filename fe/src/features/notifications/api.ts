import { getAuthAccessToken, initializeAuthHeader } from "@/features/auth/api";
import type { AppNotification } from "@/features/notifications/types";
import { apiGet, apiPost, endpointPath, envEndpoint } from "@/lib/api-utils";

const API_BASE_URL = (
  process.env.NEXT_PUBLIC_API_BASE_URL ??
  process.env.VITE_API_BASE_URL ??
  "/api"
).trim();

const NOTIFICATIONS_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_NOTIFICATIONS_ENDPOINT,
  process.env.VITE_NOTIFICATIONS_ENDPOINT,
  "/notifications",
);

function buildApiUrl(path: string) {
  if (/^https?:\/\//i.test(path)) {
    return path;
  }

  const normalizedBase = API_BASE_URL.replace(/\/+$/, "");
  const normalizedPath = path.replace(/^\/+/, "");
  if (!normalizedBase || normalizedBase === "/") {
    return `/${normalizedPath}`;
  }

  return `${normalizedBase}/${normalizedPath}`;
}

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

export function createPostUploadNotification(post: {
  id: number;
  user_id: number;
  user_name: string;
  caption: string;
  location_name: string;
  created_at: string;
}): AppNotification {
  const actor = post.user_name.trim() || "Someone";
  const detail = post.caption.trim() || post.location_name.trim();

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
  const cleanIds = Array.from(new Set(ids.map((id) => id.trim()).filter(Boolean)));
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
