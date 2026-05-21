"use client";

import { Bookmark, Camera, LogIn, MapIcon, Plus, UserRound } from "lucide-react";
import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import type { VariantProps } from "class-variance-authority";
import { Button, buttonVariants } from "@/components/ui/button";
import { hasAuthSession } from "@/features/auth/api";
import { ROUTES } from "@/lib/routes";
import { cn } from "@/lib/utils";

type ActionVariant = VariantProps<typeof buttonVariants>["variant"];

type TopbarAction = {
  id: string;
  label: string;
  to?: string;
  onClick?: () => void;
  icon?: ReactNode;
  variant?: ActionVariant;
  disabled?: boolean;
};

function requiresAuthAction(action: TopbarAction) {
  return (
    action.id === "logout" ||
    action.to === ROUTES.upload ||
    action.to === ROUTES.saved ||
    action.to === ROUTES.profile ||
    action.to === ROUTES.dashboard
  );
}

function TopbarActionButton({
  action,
  fullWidth = false,
}: Readonly<{
  action: TopbarAction;
  fullWidth?: boolean;
}>) {
  const classes = cn(fullWidth ? "w-full justify-start" : "");

  if (action.to && !action.disabled) {
    return (
      <Button
        asChild
        className={classes}
        size={fullWidth ? "default" : "sm"}
        variant={action.variant ?? "outline"}
      >
        <Link href={action.to}>
          {action.icon}
          {action.label}
        </Link>
      </Button>
    );
  }

  return (
    <Button
      className={classes}
      disabled={action.disabled}
      onClick={action.onClick}
      size={fullWidth ? "default" : "sm"}
      type="button"
      variant={action.variant ?? "outline"}
    >
      {action.icon}
      {action.label}
    </Button>
  );
}

export function AppTopbar({
  brand,
  brandIcon,
  subtitle,
  meta,
  actions,
  mobileMenuTitle,
}: Readonly<{
  brand: string;
  brandIcon?: ReactNode;
  subtitle?: string;
  meta?: ReactNode;
  actions: TopbarAction[];
  mobileMenuTitle: string;
}>) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const visibleActions = useMemo(
    () =>
      actions.filter(
        (action) => isAuthenticated || !requiresAuthAction(action),
      ),
    [actions, isAuthenticated],
  );

  useEffect(() => {
    setIsAuthenticated(hasAuthSession());
  }, []);

  return (
    <>
      <div className="app-topbar-panel hidden sm:flex">
        <div className="min-w-0 space-y-1">
          <div className="app-brand">
            <span className="app-brand-dot">{brandIcon}</span>
            <span className="truncate">{brand}</span>
          </div>
          {subtitle ? (
            <p className="mt-1 hidden text-xs text-[#5a7aa2] sm:block">
              {subtitle}
            </p>
          ) : null}
          {meta ? <div className="hidden sm:block">{meta}</div> : null}
        </div>

        <div className="hidden items-center gap-2 sm:flex">
          {visibleActions.map((action) => (
            <TopbarActionButton action={action} key={action.id} />
          ))}
        </div>
      </div>

      <nav className="fixed inset-x-0 bottom-0 z-50 border-t border-black/8 bg-[#f7f7f5]/94 px-3 pb-[calc(env(safe-area-inset-bottom)+0.65rem)] pt-2 shadow-[0_-18px_48px_-34px_rgb(0_0_0/0.72)] backdrop-blur-2xl sm:hidden">
        <div className="mx-auto max-w-md">
          <div className="mb-2 flex items-center justify-between gap-3 rounded-2xl border border-black/6 bg-white px-3 py-2">
            <div className="min-w-0">
              <p className="truncate text-xs font-semibold uppercase tracking-[0.14em] text-[#777]">
                {mobileMenuTitle}
              </p>
              <p className="truncate text-sm font-semibold text-[#111]">
                {brand}
              </p>
            </div>
            {meta ? <div className="shrink-0">{meta}</div> : null}
          </div>

          <div
            className={cn(
              "grid items-center gap-1",
              isAuthenticated ? "grid-cols-5" : "grid-cols-3",
            )}
          >
            <Link
              aria-label="Explore"
              className="flex flex-col items-center gap-1 rounded-2xl bg-[#111] px-2 py-2 text-[11px] font-semibold text-white"
              href={ROUTES.explore}
            >
              <Camera className="size-4" />
              Explore
            </Link>
            <Link
              aria-label="Destinations"
              className="flex flex-col items-center gap-1 rounded-2xl px-2 py-2 text-[11px] font-semibold text-[#555] transition hover:bg-white"
              href={ROUTES.locations}
            >
              <MapIcon className="size-4" />
              Places
            </Link>
            {isAuthenticated ? (
              <>
                <Link
                  aria-label="Upload"
                  className="mx-auto flex size-12 items-center justify-center rounded-full bg-[#ff385c] text-white shadow-[0_16px_34px_-22px_rgb(255_56_92/0.85)]"
                  href={ROUTES.upload}
                >
                  <Plus className="size-5" />
                </Link>
                <Link
                  aria-label="Saved"
                  className="flex flex-col items-center gap-1 rounded-2xl px-2 py-2 text-[11px] font-semibold text-[#555] transition hover:bg-white"
                  href={ROUTES.saved}
                >
                  <Bookmark className="size-4" />
                  Saved
                </Link>
              </>
            ) : null}
            <Link
              aria-label={isAuthenticated ? "Profile" : "Login"}
              className="flex flex-col items-center gap-1 rounded-2xl px-2 py-2 text-[11px] font-semibold text-[#555] transition hover:bg-white"
              href={isAuthenticated ? ROUTES.profile : ROUTES.login}
            >
              {isAuthenticated ? (
                <UserRound className="size-4" />
              ) : (
                <LogIn className="size-4" />
              )}
              {isAuthenticated ? "Account" : "Login"}
            </Link>
          </div>
        </div>
      </nav>
    </>
  );
}
