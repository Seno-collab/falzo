import type { AuthUser } from "@/features/auth/types";

export function readAuthUserText(
  user: AuthUser | null | undefined,
  keys: string[],
) {
  if (!user) {
    return null;
  }

  for (const key of keys) {
    const value = user[key];
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }

    if (typeof value === "number" && Number.isFinite(value)) {
      return String(value);
    }
  }

  return null;
}

export function getAuthUserDisplayName(
  user: AuthUser | null | undefined,
  fallback = "Falzo traveler",
) {
  return (
    readAuthUserText(user, [
      "fullName",
      "name",
      "displayName",
      "user_name",
      "userName",
      "username",
      "email",
    ]) ?? fallback
  );
}

export function getAuthUserInitials(name: string) {
  const parts = name
    .split(/\s+/)
    .map((part) => part[0])
    .filter(Boolean)
    .slice(0, 2);

  return parts.join("").toUpperCase() || "F";
}
