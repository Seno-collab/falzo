import type { AuthSession, AuthTokens } from "@/types/auth";

const SESSION_KEY = "falzo.session";

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

export function getSession(): AuthSession | null {
  if (!isBrowser()) return null;

  const raw = localStorage.getItem(SESSION_KEY);
  if (!raw) return null;

  try {
    const session = JSON.parse(raw) as AuthSession;
    const expiresAt = session.expires_at
      ?? readAccessTokenExpiration(session.access_token);

    if (!expiresAt || expiresAt <= Date.now()) {
      clearSession();
      return null;
    }

    if (!session.expires_at) {
      session.expires_at = expiresAt;
      localStorage.setItem(SESSION_KEY, JSON.stringify(session));
    }
    return session;
  } catch {
    clearSession();
    return null;
  }
}

export function updateTokens(tokens: AuthTokens) {
  const session = getSession();
  if (session) saveSession(session.username, tokens);
}

export function clearSession() {
  if (isBrowser()) localStorage.removeItem(SESSION_KEY);
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
