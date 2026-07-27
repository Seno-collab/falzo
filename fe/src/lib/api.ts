import type {
  ApiErrorPayload,
  ApiResponse,
  AuthTokens,
  GoogleLoginResult,
} from "@/types/auth";
import type { RoomLanguage, RoomResponse, RoundCardResponse } from "@/types/room";
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

function roomRequest<T>(
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
  return roomRequest<RoomResponse[]>(accessToken, "/v1/rooms", {
    method: "GET",
  }, options);
}

export function getRoom(accessToken: string, roomId: string, options?: RequestOptions) {
  return roomRequest<RoomResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}`,
    { method: "GET" },
    options,
  );
}

export function createRoom(
  accessToken: string,
  input: { name: string; maxPlayers: number; languageCode: RoomLanguage },
) {
  return roomRequest<RoomResponse>(accessToken, "/v1/rooms", {
    method: "POST",
    body: JSON.stringify({
      name: input.name,
      max_players: input.maxPlayers,
      language_code: input.languageCode,
    }),
  });
}

export function joinRoom(accessToken: string, inviteCode: string) {
  return roomRequest<RoomResponse>(accessToken, "/v1/rooms/join", {
    method: "POST",
    body: JSON.stringify({ invite_code: inviteCode }),
  });
}

export function dealRoomRound(accessToken: string, roomId: string) {
  return roomRequest<RoundCardResponse>(
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
  return roomRequest<RoundCardResponse>(
    accessToken,
    `/v1/rooms/${encodeURIComponent(roomId)}/rounds/current/card`,
    { method: "GET" },
    options,
  );
}
