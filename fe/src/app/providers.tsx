"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import dynamic from "next/dynamic";
import { useEffect, useState, type PropsWithChildren } from "react";
import { initializeAuthHeader } from "@/features/auth/api";

const Toaster = dynamic(() => import("sonner").then((mod) => mod.Toaster), {
  ssr: false,
});

export function AppProviders({ children }: Readonly<PropsWithChildren>) {
  const [isToasterReady, setIsToasterReady] = useState(false);
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
      {children}
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
