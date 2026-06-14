"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import dynamic from "next/dynamic";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useRef, useState, type PropsWithChildren } from "react";
import { toast } from "sonner";
import { CharacterCursor } from "@/components/layout/character-cursor";
import { hasAuthSession, initializeAuthHeader } from "@/features/auth/api";
import { LocaleProvider } from "@/i18n/locale-provider";
import { ROUTES } from "@/lib/routes";

const Toaster = dynamic(() => import("sonner").then((mod) => mod.Toaster), {
  ssr: false,
});

const guestLoginPromptDelayMs = 90_000;
const guestLoginRedirectDelayMs = 1_800;
const guestUsageStartedAtKey = "falzo.guest_usage_started_at";
const authRoutes = new Set<string>([ROUTES.login, ROUTES.register]);

function getGuestUsageStartedAt() {
  const storedValue = sessionStorage.getItem(guestUsageStartedAtKey);
  const startedAt = storedValue ? Number(storedValue) : NaN;

  if (Number.isFinite(startedAt) && startedAt > 0) {
    return startedAt;
  }

  const now = Date.now();
  sessionStorage.setItem(guestUsageStartedAtKey, String(now));
  return now;
}

function clearGuestUsageStartedAt() {
  sessionStorage.removeItem(guestUsageStartedAtKey);
}

export function AppProviders({ children }: Readonly<PropsWithChildren>) {
  const pathname = usePathname();
  const router = useRouter();
  const [isToasterReady, setIsToasterReady] = useState(false);
  const guestPromptTimerRef = useRef<ReturnType<typeof setTimeout> | null>(
    null,
  );
  const guestRedirectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(
    null,
  );
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 30_000,
            refetchOnWindowFocus: false,
            retry: 1,
          },
        },
      }),
  );

  useEffect(() => {
    initializeAuthHeader();
  }, []);

  useEffect(() => {
    const clearGuestTimers = () => {
      if (guestPromptTimerRef.current) {
        clearTimeout(guestPromptTimerRef.current);
        guestPromptTimerRef.current = null;
      }

      if (guestRedirectTimerRef.current) {
        clearTimeout(guestRedirectTimerRef.current);
        guestRedirectTimerRef.current = null;
      }
    };

    if (hasAuthSession()) {
      clearGuestTimers();
      clearGuestUsageStartedAt();
      return clearGuestTimers;
    }

    if (authRoutes.has(pathname)) {
      clearGuestTimers();
      return clearGuestTimers;
    }

    if (!guestPromptTimerRef.current) {
      const startedAt = getGuestUsageStartedAt();
      const remainingDelay = Math.max(
        guestLoginPromptDelayMs - (Date.now() - startedAt),
        0,
      );

      guestPromptTimerRef.current = setTimeout(() => {
        guestPromptTimerRef.current = null;

        if (hasAuthSession() || authRoutes.has(globalThis.location.pathname)) {
          return;
        }

        toast.info("Please login to continue exploring.", {
          description:
            "You have used Falzo for 1 minute 30 seconds as a guest.",
        });

        guestRedirectTimerRef.current = setTimeout(() => {
          if (!hasAuthSession()) {
            router.replace(ROUTES.login);
          }
        }, guestLoginRedirectDelayMs);
      }, remainingDelay);
    }

    return clearGuestTimers;
  }, [pathname, router]);

  useEffect(() => {
    const scheduleIdle =
      globalThis.requestIdleCallback ??
      ((callback: IdleRequestCallback) =>
        globalThis.setTimeout(() => callback({} as IdleDeadline), 1200));
    const cancelIdle =
      globalThis.cancelIdleCallback ??
      ((handle: number) => globalThis.clearTimeout(handle));
    const handle = scheduleIdle(() => setIsToasterReady(true));

    return () => cancelIdle(handle);
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <LocaleProvider>
        {children}
        <CharacterCursor />
      </LocaleProvider>
      {isToasterReady ? (
        <Toaster
          closeButton
          position="top-center"
          toastOptions={{
            className:
              "!rounded-xl !border !border-[#d7e2ef] !bg-white !text-[#143052]",
          }}
        />
      ) : null}
    </QueryClientProvider>
  );
}
