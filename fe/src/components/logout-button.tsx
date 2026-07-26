"use client";

import { useRouter } from "next/navigation";
import { logout } from "@/lib/api";
import { clearSession, getStoredSession } from "@/lib/auth";

export function LogoutButton() {
  const router = useRouter();

  async function handleLogout() {
    const session = getStoredSession();
    clearSession();

    if (session?.refresh_token) {
      await logout(session.refresh_token).catch(() => undefined);
    }

    router.push("/");
  }

  return (
    <button className="button button-secondary" onClick={handleLogout} type="button">
      Sign out
    </button>
  );
}
