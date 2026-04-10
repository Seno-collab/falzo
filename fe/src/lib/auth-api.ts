import { AxiosError } from "axios";
import { http } from "@/lib/http";

const ACCESS_TOKEN_KEY = "falzo.access_token";
const REFRESH_TOKEN_KEY = "falzo.refresh_token";

type LoginPayload = {
  email: string;
  password: string;
  remember: boolean;
};

type RegisterPayload = {
  fullName: string;
  email: string;
  password: string;
};

type AuthSession = {
  accessToken: string;
  refreshToken?: string;
};

function asRecord(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }

  return value as Record<string, unknown>;
}

function findFirstString(value: unknown, keys: string[], depth = 0): string | null {
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

function readMessage(data: unknown): string | null {
  const payload = asRecord(data);
  if (!payload) {
    return null;
  }

  const message = payload.message ?? payload.error ?? payload.detail;
  return typeof message === "string" && message.trim() ? message : null;
}

function writeStorage(remember: boolean, key: string, value: string) {
  const primary = remember ? localStorage : sessionStorage;
  const secondary = remember ? sessionStorage : localStorage;

  primary.setItem(key, value);
  secondary.removeItem(key);
}

function clearStorageKey(key: string) {
  localStorage.removeItem(key);
  sessionStorage.removeItem(key);
}

function getStoredValue(key: string): string | null {
  return localStorage.getItem(key) ?? sessionStorage.getItem(key);
}

function persistSession(session: AuthSession, remember: boolean) {
  writeStorage(remember, ACCESS_TOKEN_KEY, session.accessToken);

  if (session.refreshToken) {
    writeStorage(remember, REFRESH_TOKEN_KEY, session.refreshToken);
  } else {
    clearStorageKey(REFRESH_TOKEN_KEY);
  }

  http.defaults.headers.common.Authorization = `Bearer ${session.accessToken}`;
}

function parseSession(data: unknown): AuthSession | null {
  const accessToken = findFirstString(data, [
    "accessToken",
    "access_token",
    "token",
    "jwt",
    "id_token",
  ]);

  if (!accessToken) {
    return null;
  }

  const refreshToken = findFirstString(data, ["refreshToken", "refresh_token"]);
  return {
    accessToken,
    refreshToken: refreshToken ?? undefined,
  };
}

export function initializeAuthHeader() {
  if (typeof window === "undefined") {
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
  if (error instanceof AxiosError) {
    const serverMessage = readMessage(error.response?.data);
    if (serverMessage) {
      return serverMessage;
    }

    const status = error.response?.status;
    if (status === 401) {
      return "Email hoặc mật khẩu không đúng.";
    }

    if (status === 409) {
      return "Email đã được đăng ký.";
    }

    if (status && status >= 500) {
      return "Máy chủ đang lỗi, vui lòng thử lại sau.";
    }

    return "Không thể kết nối API xác thực.";
  }

  if (error instanceof Error && error.message.trim()) {
    return error.message;
  }

  return "Có lỗi xảy ra, vui lòng thử lại.";
}

export async function loginApi(payload: LoginPayload): Promise<AuthSession> {
  const endpoint = import.meta.env.VITE_AUTH_LOGIN_ENDPOINT ?? "/auth/login";
  const response = await http.post(endpoint, {
    email: payload.email,
    password: payload.password,
  });

  const session = parseSession(response.data);
  if (!session) {
    throw new Error("API login không trả về access token.");
  }

  persistSession(session, payload.remember);
  return session;
}

export async function registerApi(payload: RegisterPayload) {
  const endpoint = import.meta.env.VITE_AUTH_REGISTER_ENDPOINT ?? "/auth/register";

  const response = await http.post(endpoint, {
    fullName: payload.fullName,
    email: payload.email,
    password: payload.password,
  });

  const session = parseSession(response.data);
  if (session) {
    persistSession(session, true);
  }

  return response.data;
}
