import { AxiosError } from "axios";
import type { AxiosRequestConfig } from "axios";
import { messages } from "@/i18n/messages";
import { http } from "@/lib/http";
import type {
  AuthSession,
  AuthUser,
  LoginRequest,
  RegisterRequest,
} from "@/features/auth/types";
import type { ApiEnvelope } from "@/types/api/response";

const ACCESS_TOKEN_KEY = "falzo.access_token";
const REFRESH_TOKEN_KEY = "falzo.refresh_token";
const AUTH_EXCLUDED_RETRY =
  /\/auth\/(login|register|refresh(?:-token)?|logout)\b/i;
const AUTH_ENDPOINTS = {
  login:
    process.env.NEXT_PUBLIC_AUTH_LOGIN_ENDPOINT ??
    process.env.VITE_AUTH_LOGIN_ENDPOINT ??
    "/auth/login",
  register:
    process.env.NEXT_PUBLIC_AUTH_REGISTER_ENDPOINT ??
    process.env.VITE_AUTH_REGISTER_ENDPOINT ??
    "/auth/register",
  me:
    process.env.NEXT_PUBLIC_AUTH_ME_ENDPOINT ??
    process.env.VITE_AUTH_ME_ENDPOINT ??
    "/auth/me",
  refresh:
    process.env.NEXT_PUBLIC_AUTH_REFRESH_ENDPOINT ??
    process.env.VITE_AUTH_REFRESH_ENDPOINT ??
    "/auth/refresh-token",
  logout:
    process.env.NEXT_PUBLIC_AUTH_LOGOUT_ENDPOINT ??
    process.env.VITE_AUTH_LOGOUT_ENDPOINT ??
    "/auth/logout",
} as const;

type StorageScope = "local" | "session";
type RetryableRequestConfig = AxiosRequestConfig & {
  _retry?: boolean;
  skipAuthRefresh?: boolean;
};

let authInterceptorInstalled = false;
let refreshInFlight: Promise<AuthSession> | null = null;

function asRecord(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }

  return value as Record<string, unknown>;
}

function findFirstString(
  value: unknown,
  keys: string[],
  depth = 0,
): string | null {
  if (depth > 5) {
    return null;
  }

  const node = asRecord(value);
  if (!node) {
    return null;
  }

  for (const key of keys) {
    const candidate = node[key];
    if (typeof candidate === "string" && candidate.trim()) {
      return candidate;
    }
  }

  for (const nested of Object.values(node)) {
    const token = findFirstString(nested, keys, depth + 1);
    if (token) {
      return token;
    }
  }

  return null;
}

function extractDataNode(value: unknown): unknown {
  const node = asRecord(value);
  if (!node || !("data" in node)) {
    return value;
  }

  return node.data;
}

function unwrapResponseData<T>(value: ApiEnvelope<T> | T): T {
  const node = asRecord(value);
  if (!node || !("data" in node) || !("success" in node)) {
    return value as T;
  }

  return node.data as T;
}

function readMessage(data: unknown): string | null {
  const payload = asRecord(data);
  if (!payload) {
    return null;
  }

  const firstError = Array.isArray(payload.errors) ? payload.errors[0] : null;
  const errorNode = asRecord(firstError);
  const nestedMessage = errorNode?.message;
  const message =
    payload.message ?? nestedMessage ?? payload.error ?? payload.detail;
  return typeof message === "string" && message.trim() ? message : null;
}

function writeStorage(scope: StorageScope, key: string, value: string) {
  if (globalThis.window === undefined) {
    return;
  }

  const primary = scope === "local" ? localStorage : sessionStorage;
  const secondary = scope === "local" ? sessionStorage : localStorage;

  primary.setItem(key, value);
  secondary.removeItem(key);
}

function clearStorageKey(key: string) {
  if (globalThis.window === undefined) {
    return;
  }

  localStorage.removeItem(key);
  sessionStorage.removeItem(key);
}

function getStoredValue(key: string): string | null {
  if (globalThis.window === undefined) {
    return null;
  }

  return localStorage.getItem(key) ?? sessionStorage.getItem(key);
}

function getStorageScopeForKey(key: string): StorageScope | null {
  if (globalThis.window === undefined) {
    return null;
  }

  if (localStorage.getItem(key)) {
    return "local";
  }

  if (sessionStorage.getItem(key)) {
    return "session";
  }

  return null;
}

function resolveStorageScope(remember?: boolean): StorageScope {
  if (remember === true) {
    return "local";
  }

  if (remember === false) {
    return "session";
  }

  return (
    getStorageScopeForKey(ACCESS_TOKEN_KEY) ??
    getStorageScopeForKey(REFRESH_TOKEN_KEY) ??
    "local"
  );
}

function persistSession(session: AuthSession, scope: StorageScope) {
  writeStorage(scope, ACCESS_TOKEN_KEY, session.accessToken);

  if (session.refreshToken) {
    writeStorage(scope, REFRESH_TOKEN_KEY, session.refreshToken);
  } else {
    clearStorageKey(REFRESH_TOKEN_KEY);
  }

  http.defaults.headers.common.Authorization = `Bearer ${session.accessToken}`;
}

function parseSession(data: unknown): AuthSession | null {
  const dataNode = extractDataNode(data);
  const accessToken =
    findFirstString(dataNode, [
      "accessToken",
      "access_token",
      "token",
      "jwt",
      "id_token",
    ]) ??
    findFirstString(data, [
      "accessToken",
      "access_token",
      "token",
      "jwt",
      "id_token",
    ]);

  if (!accessToken) {
    return null;
  }

  const refreshToken =
    findFirstString(dataNode, ["refreshToken", "refresh_token"]) ??
    findFirstString(data, ["refreshToken", "refresh_token"]);
  return {
    accessToken,
    refreshToken: refreshToken ?? undefined,
  };
}

function isExcludedFromRetry(url?: string): boolean {
  if (!url) {
    return false;
  }

  return AUTH_EXCLUDED_RETRY.test(url);
}

async function refreshTokenApi(): Promise<AuthSession> {
  const endpoint = AUTH_ENDPOINTS.refresh;
  const storageScope = resolveStorageScope();
  const currentRefreshToken = getStoredValue(REFRESH_TOKEN_KEY);

  const response = await http.post(
    endpoint,
    currentRefreshToken ? { refreshToken: currentRefreshToken } : {},
    { skipAuthRefresh: true } as RetryableRequestConfig,
  );

  const refreshedSession = parseSession(response.data);
  if (!refreshedSession) {
    throw new Error("Refresh token API did not return an access token.");
  }

  if (!refreshedSession.refreshToken && currentRefreshToken) {
    refreshedSession.refreshToken = currentRefreshToken;
  }

  persistSession(refreshedSession, storageScope);
  return refreshedSession;
}

function getRefreshPromise() {
  refreshInFlight ??= refreshTokenApi().finally(() => {
    refreshInFlight = null;
  });

  return refreshInFlight;
}

function installAuthInterceptor() {
  if (authInterceptorInstalled) {
    return;
  }

  authInterceptorInstalled = true;

  http.interceptors.response.use(
    (response) => response,
    async (error: unknown) => {
      if (!(error instanceof AxiosError)) {
        throw error;
      }

      const status = error.response?.status;
      const requestConfig = error.config as RetryableRequestConfig | undefined;
      if (!requestConfig) {
        throw error;
      }

      if (
        status !== 401 ||
        requestConfig._retry ||
        requestConfig.skipAuthRefresh ||
        isExcludedFromRetry(requestConfig.url)
      ) {
        throw error;
      }

      if (!getStoredValue(REFRESH_TOKEN_KEY)) {
        clearAuthSession();
        throw error;
      }

      requestConfig._retry = true;

      try {
        const session = await getRefreshPromise();
        requestConfig.headers = {
          ...requestConfig.headers,
          Authorization: `Bearer ${session.accessToken}`,
        };

        return http.request(requestConfig);
      } catch {
        clearAuthSession();
        throw error;
      }
    },
  );
}

export function initializeAuthHeader() {
  if (globalThis.window !== undefined) {
    installAuthInterceptor();
  }

  if (globalThis.window === undefined) {
    return;
  }

  const accessToken = getStoredValue(ACCESS_TOKEN_KEY);
  if (!accessToken) {
    delete http.defaults.headers.common.Authorization;
    return;
  }

  http.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
}

export function hasAuthSession() {
  return Boolean(getStoredValue(ACCESS_TOKEN_KEY));
}

export function clearAuthSession() {
  clearStorageKey(ACCESS_TOKEN_KEY);
  clearStorageKey(REFRESH_TOKEN_KEY);
  delete http.defaults.headers.common.Authorization;
}

export function getApiErrorMessage(error: unknown): string {
  const errorMessages = messages.en.apiErrors;

  if (error instanceof AxiosError) {
    const serverMessage = readMessage(error.response?.data);
    if (serverMessage) {
      return serverMessage;
    }

    const status = error.response?.status;
    if (status === 401) {
      return errorMessages.unauthorized;
    }

    if (status === 409) {
      return errorMessages.conflict;
    }

    if (status && status >= 500) {
      return errorMessages.server;
    }

    return errorMessages.unreachable;
  }

  if (error instanceof Error && error.message.trim()) {
    return error.message;
  }

  return errorMessages.generic;
}

export async function loginApi(payload: LoginRequest): Promise<AuthSession> {
  const endpoint = AUTH_ENDPOINTS.login;
  const response = await http.post(endpoint, {
    email: payload.email,
    password: payload.password,
  });

  const session = parseSession(response.data);
  if (!session) {
    throw new Error("Login API did not return an access token.");
  }

  persistSession(session, resolveStorageScope(payload.remember));
  return session;
}

export async function registerApi(payload: RegisterRequest) {
  const endpoint = AUTH_ENDPOINTS.register;

  const response = await http.post(endpoint, {
    user_name: payload.fullName,
    email: payload.email,
    password: payload.password,
  });

  const session = parseSession(response.data);
  if (session) {
    persistSession(session, "local");
  }

  return response.data;
}

export async function getMeApi<
  TUser extends AuthUser = AuthUser,
>(): Promise<TUser> {
  const endpoint = AUTH_ENDPOINTS.me;
  const response = await http.get<ApiEnvelope<TUser> | TUser>(endpoint);
  return unwrapResponseData(response.data);
}

export async function logoutApi(): Promise<void> {
  const endpoint = AUTH_ENDPOINTS.logout;
  const refreshToken = getStoredValue(REFRESH_TOKEN_KEY);

  try {
    await http.post(endpoint, refreshToken ? { refreshToken } : {}, {
      skipAuthRefresh: true,
    } as RetryableRequestConfig);
  } finally {
    clearAuthSession();
  }
}
