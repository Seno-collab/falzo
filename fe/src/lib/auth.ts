import type { AuthSession, AuthTokens } from "@/types/auth";
import { refresh } from "@/lib/api";

const SESSION_KEY = "falzo.session";
let pendingRefresh: {
  refreshToken: string;
  promise: Promise<AuthSession | null>;
} | null = null;

function isBrowser() {
  return typeof window !== "undefined";
}

export function saveSession(username: string, tokens: AuthTokens) {
  if (!isBrowser()) return;

  const expiresAt = readAccessTokenExpiration(tokens.access_token)
    ?? Date.now() + tokens.expires_in * 1000;
  const session: AuthSession = { username, ...tokens, expires_at: expiresAt };
  localStorage.setItem(SESSION_KEY, JSON.stringify(session));
}

export function getStoredSession(): AuthSession | null {
  if (!isBrowser()) return null;

  try {
    const raw = localStorage.getItem(SESSION_KEY);
    if (!raw) return null;

    const parsed: unknown = JSON.parse(raw);
    if (!isAuthSession(parsed)) {
      clearSession();
      return null;
    }

    const session = parsed;
    const expiresAt = session.expires_at
      ?? readAccessTokenExpiration(session.access_token);

    if (expiresAt && !session.expires_at) {
      session.expires_at = expiresAt;
      localStorage.setItem(SESSION_KEY, JSON.stringify(session));
    }

    return session;
  } catch {
    clearSession();
    return null;
  }
}

export function getSession(): AuthSession | null {
  const session = getStoredSession();
  return session && isAccessTokenValid(session) ? session : null;
}

export async function restoreSession(): Promise<AuthSession | null> {
  const session = getStoredSession();
  if (!session) return null;
  if (isAccessTokenValid(session)) return session;

  if (pendingRefresh?.refreshToken === session.refresh_token) {
    return pendingRefresh.promise;
  }

  const promise = refreshExpiredSession(session);
  pendingRefresh = { refreshToken: session.refresh_token, promise };

  void promise.finally(() => {
    if (pendingRefresh?.promise === promise) pendingRefresh = null;
  });

  return promise;
}

export function updateTokens(tokens: AuthTokens) {
  const session = getStoredSession();
  if (session) saveSession(session.username, tokens);
}

export function clearSession() {
  if (isBrowser()) localStorage.removeItem(SESSION_KEY);
}

async function refreshExpiredSession(session: AuthSession): Promise<AuthSession | null> {
  try {
    const tokens = await refresh(session.refresh_token);
    const currentSession = getStoredSession();

    if (currentSession?.refresh_token !== session.refresh_token) return null;

    saveSession(session.username, tokens);
    return getSession();
  } catch {
    const currentSession = getStoredSession();
    if (currentSession?.refresh_token === session.refresh_token) clearSession();
    return null;
  }
}

function isAccessTokenValid(session: AuthSession) {
  const expiresAt = session.expires_at
    ?? readAccessTokenExpiration(session.access_token);
  return Boolean(expiresAt && expiresAt > Date.now());
}

function isAuthSession(value: unknown): value is AuthSession {
  if (!value || typeof value !== "object") return false;

  const session = value as Partial<AuthSession>;
  return typeof session.username === "string"
    && typeof session.access_token === "string"
    && typeof session.refresh_token === "string"
    && typeof session.token_type === "string"
    && typeof session.expires_in === "number"
    && (session.expires_at === undefined || typeof session.expires_at === "number");
}

function readAccessTokenExpiration(accessToken: string): number | null {
  try {
    const payloadPart = accessToken.split(".")[1];
    if (!payloadPart) return null;

    const base64 = payloadPart.replace(/-/g, "+").replace(/_/g, "/");
    const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, "=");
    const payload = JSON.parse(atob(padded)) as { exp?: unknown };
    return typeof payload.exp === "number" ? payload.exp * 1000 : null;
  } catch {
    return null;
  }
}
