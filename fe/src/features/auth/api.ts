import { AxiosError } from "axios";
import type { AxiosRequestConfig } from "axios";
import { readCurrentLocale } from "@/i18n/locale";
import { messages } from "@/i18n/messages";
import { AUTH_ENDPOINTS } from "@/lib/api-config";
import { http } from "@/lib/http";
import { apiGet, apiPatch, apiPost } from "@/lib/api-utils";
import type {
  AuthSession,
  AuthUser,
  ChangePasswordRequest,
  LoginRequest,
  RegisterRequest,
} from "@/features/auth/types";

const ACCESS_TOKEN_KEY = "falzo.access_token";
const REFRESH_TOKEN_KEY = "falzo.refresh_token";
const authSessionChangedEvent = "falzo:auth-session-changed";
const AUTH_EXCLUDED_RETRY =
  /\/auth\/(login|register|refresh(?:-token)?|logout)\b/i;

type StorageScope = "local" | "session";
type RetryableRequestConfig = AxiosRequestConfig & {
  _retry?: boolean;
  skipAuthRefresh?: boolean;
};

let authInterceptorInstalled = false;
let refreshInFlight: Promise<AuthSession> | null = null;

function canUseBrowserStorage() {
  return (
    globalThis.localStorage !== undefined &&
    globalThis.sessionStorage !== undefined
  );
}

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

function readMessage(data: unknown): string | null {
  const payload = asRecord(data);
  if (!payload) {
    return null;
  }

  const firstError = Array.isArray(payload.errors) ? payload.errors[0] : null;
  const errorNode = asRecord(firstError);
  const nestedMessage = errorNode?.message;
  const message = payload.message ?? nestedMessage ?? payload.error ?? payload.detail;
  return typeof message === "string" && message.trim() ? message : null;
}

type ApiErrorDetailPayload = {
  code?: string;
  field?: string;
  message?: string;
};

function readFirstErrorDetail(data: unknown): ApiErrorDetailPayload | null {
  const payload = asRecord(data);
  const firstError = Array.isArray(payload?.errors) ? payload.errors[0] : null;
  const detail = asRecord(firstError);
  if (!detail) {
    return null;
  }

  return {
    code: typeof detail.code === "string" ? detail.code : undefined,
    field: typeof detail.field === "string" ? detail.field : undefined,
    message: typeof detail.message === "string" ? detail.message : undefined,
  };
}

function getCurrentLocale() {
  return readCurrentLocale();
}

function normalizeBackendMessage(value: string) {
  return value.trim().replaceAll(/\s+/g, " ");
}

function translateBackendMessage(
  message: string | null | undefined,
  locale = getCurrentLocale(),
) {
  if (!message?.trim()) {
    return null;
  }

  const normalizedMessage = normalizeBackendMessage(message);
  const localizedMessages = messages[locale].apiErrors
    .backendMessages as Record<string, string>;
  const englishMessages = messages.en.apiErrors.backendMessages as Record<
    string,
    string
  >;
  const localized =
    localizedMessages[normalizedMessage];
  if (localized) {
    return localized;
  }

  return englishMessages[normalizedMessage] ?? normalizedMessage;
}

export function getApiFieldErrors(error: unknown) {
  if (!(error instanceof AxiosError)) {
    return [];
  }

  const payload = asRecord(error.response?.data);
  const errors = Array.isArray(payload?.errors) ? payload.errors : [];

  return errors
    .map((item) => {
      const detail = asRecord(item);
      const field = detail?.field;
      const message = detail?.message;
      const locale = getCurrentLocale();

      if (typeof field !== "string" || typeof message !== "string") {
        return null;
      }

      return {
        field,
        message: translateBackendMessage(message, locale) ?? message,
      };
    })
    .filter((item): item is { field: string; message: string } =>
      Boolean(item),
    );
}

function writeStorage(scope: StorageScope, key: string, value: string) {
  if (!canUseBrowserStorage()) {
    return;
  }

  const primary =
    scope === "local" ? globalThis.localStorage : globalThis.sessionStorage;
  const secondary =
    scope === "local" ? globalThis.sessionStorage : globalThis.localStorage;

  primary.setItem(key, value);
  secondary.removeItem(key);
}

function clearStorageKey(key: string) {
  if (!canUseBrowserStorage()) {
    return;
  }

  globalThis.localStorage.removeItem(key);
  globalThis.sessionStorage.removeItem(key);
}

function getStoredValue(key: string): string | null {
  if (!canUseBrowserStorage()) {
    return null;
  }

  return (
    globalThis.localStorage.getItem(key) ??
    globalThis.sessionStorage.getItem(key)
  );
}

function getStorageScopeForKey(key: string): StorageScope | null {
  if (!canUseBrowserStorage()) {
    return null;
  }

  if (globalThis.localStorage.getItem(key)) {
    return "local";
  }

  if (globalThis.sessionStorage.getItem(key)) {
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
      "access_token",
      "token",
      "jwt",
      "id_token",
    ]) ??
    findFirstString(data, [
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
    currentRefreshToken ? { refresh_token: currentRefreshToken } : {},
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
  if (canUseBrowserStorage()) {
    installAuthInterceptor();
  }

  if (!canUseBrowserStorage()) {
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

export function getAuthAccessToken() {
  return getStoredValue(ACCESS_TOKEN_KEY);
}

export function clearAuthSession() {
  clearStorageKey(ACCESS_TOKEN_KEY);
  clearStorageKey(REFRESH_TOKEN_KEY);
  delete http.defaults.headers.common.Authorization;
}

function notifyAuthSessionChanged() {
  if (!canUseBrowserStorage()) {
    return;
  }

  globalThis.dispatchEvent(new Event(authSessionChangedEvent));
}

export function subscribeAuthSessionChanged(listener: () => void) {
  if (!canUseBrowserStorage()) {
    return () => undefined;
  }

  globalThis.addEventListener(authSessionChangedEvent, listener);
  return () => globalThis.removeEventListener(authSessionChangedEvent, listener);
}

export function getApiErrorMessage(error: unknown): string {
  const errorMessages = messages[getCurrentLocale()].apiErrors;

  if (error instanceof AxiosError) {
    if (process.env.NODE_ENV !== "production") {
      console.debug("API error response", {
        method: error.config?.method,
        url: error.config?.url,
        status: error.response?.status,
        data: error.response?.data,
      });
    }

    const firstErrorDetail = readFirstErrorDetail(error.response?.data);
    const translatedDetail = translateBackendMessage(firstErrorDetail?.message);
    if (translatedDetail) {
      return translatedDetail;
    }

    const serverMessage = translateBackendMessage(
      readMessage(error.response?.data),
    );
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
  notifyAuthSessionChanged();
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
    notifyAuthSessionChanged();
  }

  return response.data;
}

export async function getMeApi<
  TUser extends AuthUser = AuthUser,
>(config?: AxiosRequestConfig): Promise<TUser> {
  initializeAuthHeader();

  const endpoint = AUTH_ENDPOINTS.me;
  return apiGet<TUser>(endpoint, config);
}

export async function logoutApi(): Promise<void> {
  initializeAuthHeader();

  const endpoint = AUTH_ENDPOINTS.logout;
  const refreshToken = getStoredValue(REFRESH_TOKEN_KEY);

  try {
    await apiPost(
      endpoint,
      refreshToken ? { refresh_token: refreshToken } : {},
      { skipAuthRefresh: true } as RetryableRequestConfig,
    );
  } finally {
    clearAuthSession();
    notifyAuthSessionChanged();
  }
}

export async function changePasswordApi(
  payload: ChangePasswordRequest,
): Promise<void> {
  initializeAuthHeader();

  const endpoint = AUTH_ENDPOINTS.changePassword;
  await apiPost(
    endpoint,
    {
      current_password: payload.currentPassword,
      new_password: payload.newPassword,
    },
    { skipAuthRefresh: false } as RetryableRequestConfig,
  );
}

export async function updateAvatarApi(avatarUrl: string): Promise<AuthUser> {
  initializeAuthHeader();

  return apiPatch<AuthUser>(
    AUTH_ENDPOINTS.updateAvatar,
    {
      avatar_url: avatarUrl,
    },
    { skipAuthRefresh: false } as RetryableRequestConfig,
  );
}
