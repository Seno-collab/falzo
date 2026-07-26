import type {
  ApiErrorPayload,
  ApiResponse,
  AuthTokens,
  GoogleLoginResult,
} from "@/types/auth";
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

async function request<T>(path: string, init: RequestInit): Promise<T> {
  const finishRequest = beginApiRequest();

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
