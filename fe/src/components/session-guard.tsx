"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { useRouter } from "next/navigation";
import { ApiLoading } from "@/components/api-loading";
import { restoreSession } from "@/lib/auth";
import type { AuthSession } from "@/types/auth";

const SessionContext = createContext<AuthSession | null>(null);

export function SessionGuard({ children }: Readonly<{ children: ReactNode }>) {
  const router = useRouter();
  const [session, setSession] = useState<AuthSession | null>(null);

  useEffect(() => {
    let active = true;

    void restoreSession().then((restoredSession) => {
      if (!active) return;

      if (!restoredSession) {
        router.replace("/login");
        return;
      }

      setSession(restoredSession);
    });

    return () => {
      active = false;
    };
  }, [router]);

  if (!session) {
    return <ApiLoading label="Checking your session…" variant="page" />;
  }

  return (
    <SessionContext.Provider value={session}>
      {children}
    </SessionContext.Provider>
  );
}

export function useSession() {
  const session = useContext(SessionContext);
  if (!session) throw new Error("useSession must be used inside SessionGuard");
  return session;
}
