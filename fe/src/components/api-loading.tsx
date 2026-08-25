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
      <div className="api-loading-panel">
        <div className="api-loading-brand" aria-hidden="true">
          <span>F</span>
          <div>
            <strong>falzo</strong>
            <small>GAME LOBBY</small>
          </div>
          <i />
        </div>

        <div className="api-loading-status">
          <span aria-hidden="true">?</span>
          <div>
            <strong>{label}</strong>
            <small>Keeping your seat ready</small>
          </div>
        </div>

        <div className="api-loading-progress" aria-hidden="true">
          <span />
        </div>
      </div>
    </div>
  );

  if (variant === "page") {
    return (
      <main
        className="api-loading api-loading-page"
        aria-live="polite"
        aria-busy="true"
      >
        {content}
      </main>
    );
  }

  return (
    <div
      className="api-loading api-loading-overlay"
      role="status"
      aria-live="polite"
    >
      {content}
    </div>
  );
}

export function ApiLoadingProvider({
  children,
}: Readonly<{ children: ReactNode }>) {
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
