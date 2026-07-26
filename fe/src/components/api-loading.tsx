"use client";

import {
  useEffect,
  useState,
  useSyncExternalStore,
  type ReactNode,
} from "react";
import {
  getActiveApiRequestCount,
  subscribeToApiActivity,
} from "@/lib/api-activity";

type ApiLoadingProps = {
  label?: string;
  variant?: "overlay" | "page";
};

export function ApiLoading({
  label = "Getting things ready…",
  variant = "overlay",
}: ApiLoadingProps) {
  const content = (
    <div className="api-loading-content">
      <div className="api-loading-card" aria-hidden="true">
        <span>FALZO</span>
        <strong>?</strong>
        <small>PLEASE WAIT</small>
      </div>
      <div className="api-loading-copy">
        <strong>{label}</strong>
        <span>
          <i />
          <i />
          <i />
        </span>
      </div>
    </div>
  );

  if (variant === "page") {
    return (
      <main className="api-loading api-loading-page" aria-live="polite" aria-busy="true">
        {content}
      </main>
    );
  }

  return (
    <div className="api-loading api-loading-overlay" role="status" aria-live="polite">
      {content}
    </div>
  );
}

export function ApiLoadingProvider({ children }: Readonly<{ children: ReactNode }>) {
  const activeRequestCount = useSyncExternalStore(
    subscribeToApiActivity,
    getActiveApiRequestCount,
    () => 0,
  );
  const hasActiveRequest = activeRequestCount > 0;
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (!hasActiveRequest) {
      setVisible(false);
      return;
    }

    const timer = window.setTimeout(() => setVisible(true), 140);
    return () => window.clearTimeout(timer);
  }, [hasActiveRequest]);

  return (
    <>
      {children}
      {visible && <ApiLoading />}
    </>
  );
}
