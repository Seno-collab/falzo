"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useEffect, useState, type PropsWithChildren } from "react";
import { Toaster } from "sonner";
import { LanguageProvider } from "@/app/language-provider";
import { initializeAuthHeader } from "@/api/auth.api";

export function AppProviders({ children }: Readonly<PropsWithChildren>) {
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

  return (
    <QueryClientProvider client={queryClient}>
      <LanguageProvider>
        {children}
        <Toaster
          closeButton
          position="top-center"
          toastOptions={{
            className:
              "!rounded-xl !border !border-[#d7e2ef] !bg-white !text-[#143052]",
          }}
        />
      </LanguageProvider>
    </QueryClientProvider>
  );
}
