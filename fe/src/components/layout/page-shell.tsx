import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export function PageShell({
  children,
  topbar,
  className,
  contentClassName,
}: {
  children: ReactNode;
  topbar?: ReactNode;
  className?: string;
  contentClassName?: string;
}) {
  return (
    <div className={cn("app-shell", className)}>
      {topbar ? (
        <header className="app-topbar">
          <div className="app-container">{topbar}</div>
        </header>
      ) : null}
      <main
        className={cn(
          "app-container",
          topbar ? "pt-2" : "pt-8",
          contentClassName,
        )}
      >
        {children}
      </main>
    </div>
  );
}
