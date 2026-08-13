import type {
  ApiErrorPayload,
  ApiResponse,
  AuthTokens,
  GoogleLoginResult,
} from "@/types/auth";
import type {
  RoomLanguage,
  RoomResponse,
  RoundCardResponse,
  RoundStateResponse,
} from "@/types/room";
import type {
  Friend,
  FriendNotification,
  FriendRequest,
  SocialUser,
} from "@/types/social";
import { beginApiRequest } from "@/lib/api-activity";

const API_PREFIX = "/api";

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

type RequestOptions = {
  trackActivity?: boolean;
};

async function request<T>(
  path: string,
  init: RequestInit,
  options: RequestOptions = {},
): Promise<T> {
  const finishRequest = options.trackActivity === false
    ? () => undefined
    : beginApiRequest();

  try {
    const response = await fetch(`${API_PREFIX}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...init.headers,
      },
    });

    if (response.status === 204) {
      return undefined as T;
    }

    const payload = (await response.json().catch(() => null)) as
      | ApiResponse<T>
      | ApiErrorPayload
      | null;

    if (!response.ok) {
      throw new ApiError(
        payload?.message ?? "Something went wrong",
        response.status,
        payload?.code,
      );
    }

    return (payload as ApiResponse<T>).data;
  } finally {
    finishRequest();
  }
}

function post<T>(path: string, body: unknown) {
  return request<T>(path, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function login(username: string, password: string) {
  return post<AuthTokens>("/v1/auth/login", { username, password });
}

export function googleLogin(credential: string) {
  return post<GoogleLoginResult>("/v1/auth/google/credential", { credential });
}

export function refresh(refreshToken: string) {
  return post<AuthTokens>("/v1/auth/refresh", {
    refresh_token: refreshToken,
  });
}

export function logout(refreshToken: string) {
  return post<void>("/v1/auth/logout", { refresh_token: refreshToken });
}

function authenticatedRequest<T>(
  accessToken: string,
  path: string,
  init: RequestInit,
  options?: RequestOptions,
) {
  return request<T>(path, {
    ...init,
    headers: {
      Authorization: `Bearer ${accessToken}`,
      ...init.headers,
    },
  }, options);
}

export function listRooms(accessToken: string, options?: RequestOptions) {
  return authenticatedRequest<RoomResponse[]>(accessToken, "/v1/rooms", {
    method: "GET",
  }, options);
}

export function getRoom(accessToken: string, roomId: string, options?: RequestOptions) {
  return authenticatedRequest<RoomResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}`,
    { method: "GET" },
    options,
  );
}

export function createRoom(
  accessToken: string,
  input: { name: string; maxPlayers: number; languageCode: RoomLanguage; mrWhiteEnabled: boolean },
) {
  return authenticatedRequest<RoomResponse>(accessToken, "/v1/rooms", {
    method: "POST",
    body: JSON.stringify({
      name: input.name,
      max_players: input.maxPlayers,
      language_code: input.languageCode,
      mr_white_enabled: input.mrWhiteEnabled,
    }),
  });
}

export function joinRoom(accessToken: string, inviteCode: string) {
  return authenticatedRequest<RoomResponse>(accessToken, "/v1/rooms/join", {
    method: "POST",
    body: JSON.stringify({ invite_code: inviteCode }),
  });
}

export function dealRoomRound(accessToken: string, roomId: string) {
  return authenticatedRequest<RoundCardResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}/rounds`,
    { method: "POST" },
  );
}

export function getCurrentRoomCard(
  accessToken: string,
  roomId: string,
  options?: RequestOptions,
) {
  return authenticatedRequest<RoundCardResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}/rounds/current/card`,
    { method: "GET" },
    options,
  );
}

export function updateRoomDiscussion(
  accessToken: string,
  roomId: string,
  discussionSeconds: number,
) {
  return authenticatedRequest<RoomResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}/settings/discussion`,
    {
      method: "PATCH",
      body: JSON.stringify({ discussion_seconds: discussionSeconds }),
    },
  );
}

export function getCurrentRoundState(
  accessToken: string,
  roomId: string,
  options?: RequestOptions,
) {
  return authenticatedRequest<RoundStateResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}/rounds/current`,
    { method: "GET" },
    options,
  );
}

export function castRoomVote(accessToken: string, roomId: string, targetPlayerId: number) {
  return authenticatedRequest<RoundStateResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}/rounds/current/votes`,
    {
      method: "POST",
      body: JSON.stringify({ target_player_id: targetPlayerId }),
    },
  );
}

export function confirmRoomRole(accessToken: string, roomId: string) {
  return authenticatedRequest<RoundStateResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}/rounds/current/ready`,
    { method: "POST" },
  );
}

export function finishRoomTurn(accessToken: string, roomId: string) {
  return authenticatedRequest<RoundStateResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}/rounds/current/turn/finish`,
    { method: "POST" },
  );
}

export function submitMrWhiteGuess(accessToken: string, roomId: string, guess: string) {
  return authenticatedRequest<RoundStateResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}/rounds/current/mr-white/guess`,
    { method: "POST", body: JSON.stringify({ guess }) },
  );
}

export function searchUsers(accessToken: string, query: string, limit = 20) {
  const params = new URLSearchParams({ q: query, limit: String(limit) });
  return authenticatedRequest<SocialUser[]>(
    accessToken,
    `/v1/users/search?${params.toString()}`,
    { method: "GET" },
  );
}

export function sendFriendRequest(accessToken: string, receiverId: number) {
  return authenticatedRequest<FriendRequest>(accessToken, "/v1/friend-requests", {
    method: "POST",
    body: JSON.stringify({ receiver_id: receiverId }),
  });
}

export function listFriendRequests(accessToken: string, options?: RequestOptions) {
  return authenticatedRequest<FriendRequest[]>(accessToken, "/v1/friend-requests", {
    method: "GET",
  }, options);
}

export function acceptFriendRequest(accessToken: string, requestId: number) {
  return authenticatedRequest<FriendRequest>(
    accessToken,
    `/v1/friend-requests/${requestId}/accept`,
    { method: "POST" },
  );
}

export function rejectFriendRequest(accessToken: string, requestId: number) {
  return authenticatedRequest<FriendRequest>(
    accessToken,
    `/v1/friend-requests/${requestId}/reject`,
    { method: "POST" },
  );
}

export function cancelFriendRequest(accessToken: string, requestId: number) {
  return authenticatedRequest<void>(
    accessToken,
    `/v1/friend-requests/${requestId}`,
    { method: "DELETE" },
  );
}

export function listFriends(accessToken: string, options?: RequestOptions) {
  return authenticatedRequest<Friend[]>(accessToken, "/v1/friends", {
    method: "GET",
  }, options);
}

export function unfriend(accessToken: string, friendUserId: number) {
  return authenticatedRequest<void>(accessToken, `/v1/friends/${friendUserId}`, {
    method: "DELETE",
  });
}

export function listNotifications(
  accessToken: string,
  input: { unreadOnly?: boolean; limit?: number; offset?: number } = {},
  options?: RequestOptions,
) {
  const params = new URLSearchParams({
    unread_only: String(input.unreadOnly ?? false),
    limit: String(input.limit ?? 30),
    offset: String(input.offset ?? 0),
  });
  return authenticatedRequest<FriendNotification[]>(
    accessToken,
    `/v1/notifications?${params.toString()}`,
    { method: "GET" },
    options,
  );
}

export function countUnreadNotifications(accessToken: string, options?: RequestOptions) {
  return authenticatedRequest<{ count: number }>(
    accessToken,
    "/v1/notifications/unread-count",
    { method: "GET" },
    options,
  );
}

export function markNotificationRead(accessToken: string, notificationId: number) {
  return authenticatedRequest<void>(
    accessToken,
    `/v1/notifications/${notificationId}/read`,
    { method: "PATCH" },
  );
}

export function markAllNotificationsRead(accessToken: string) {
  return authenticatedRequest<{ updated: number }>(
    accessToken,
    "/v1/notifications/read-all",
    { method: "PATCH" },
  );
}
