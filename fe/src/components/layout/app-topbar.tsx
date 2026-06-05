"use client";

import { Bookmark, Camera, LogIn, MapIcon, Plus } from "lucide-react";
import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { VariantProps } from "class-variance-authority";
import { Button, buttonVariants } from "@/components/ui/button";
import { LanguageSwitcher } from "@/components/layout/language-switcher";
import { RainbowAvatar } from "@/components/ui/rainbow-avatar";
import { getMeApi, hasAuthSession } from "@/features/auth/api";
import type { AuthUser } from "@/features/auth/types";
import {
  getAuthUserDisplayName,
  getAuthUserInitials,
  readAuthUserText,
} from "@/features/auth/user-display";
import { useI18n } from "@/i18n/locale-provider";
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

type CommonCopy = ReturnType<typeof useI18n>["messages"]["common"];

function requiresAuthAction(action: TopbarAction) {
  return (
    action.id === "logout" ||
    action.to === ROUTES.upload ||
    action.to === ROUTES.saved ||
    action.to === ROUTES.profile ||
    action.to === ROUTES.dashboard
  );
}

function getActionLabel(action: TopbarAction, commonCopy: CommonCopy) {
  switch (action.id) {
    case "explore":
      return commonCopy.explore;
    case "locations":
      return commonCopy.places;
    case "upload":
      return commonCopy.upload;
    case "saved":
      return commonCopy.saved;
    case "login":
      return commonCopy.login;
    case "register":
      return commonCopy.register;
    case "logout":
      return commonCopy.logout;
    case "profile":
      return commonCopy.profile;
    case "dashboard":
      return commonCopy.dashboard;
    default:
      return action.label;
  }
}

function TopbarActionButton({
  action,
  fullWidth = false,
}: Readonly<{
  action: TopbarAction;
  fullWidth?: boolean;
}>) {
  const { messages } = useI18n();
  const label = getActionLabel(action, messages.common);
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
          {label}
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
      {label}
    </Button>
  );
}

function mobileNavItemClass(isActive: boolean) {
  return cn(
    "flex min-w-0 flex-col items-center gap-1 overflow-hidden rounded-2xl px-2 py-2 text-center text-[11px] font-semibold transition hover:bg-white [&>span]:max-w-full [&>span]:truncate",
    isActive ? "bg-[#111] text-white hover:bg-[#222]" : "text-[#555]",
  );
}

function mobileUploadNavItemClass(isActive: boolean) {
  return cn(
    "flex min-h-14 min-w-0 flex-col items-center justify-center gap-1 overflow-hidden rounded-2xl px-2 py-2 text-center text-[11px] font-bold text-white shadow-[0_14px_32px_-22px_rgb(255_56_92/0.95)] transition [&>span]:max-w-full [&>span]:truncate",
    isActive
      ? "bg-[#111] ring-2 ring-[#ff385c]/35 ring-offset-2 ring-offset-[#f7f7f5]"
      : "bg-[#ff385c] hover:bg-[#e63253]",
  );
}

function isActiveRoute(pathname: string, route: string) {
  if (route === ROUTES.explore) {
    return pathname === route;
  }

  return pathname === route || pathname.startsWith(`${route}/`);
}

export function AppTopbar({
  brand,
  brandIcon,
  subtitle,
  meta,
  actions,
  mobileMenuTitle,
  showMobileNav = true,
}: Readonly<{
  brand: string;
  brandIcon?: ReactNode;
  subtitle?: string;
  meta?: ReactNode;
  actions: TopbarAction[];
  mobileMenuTitle: string;
  showMobileNav?: boolean;
}>) {
  const pathname = usePathname();
  const { messages } = useI18n();
  const commonCopy = messages.common;
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [profile, setProfile] = useState<AuthUser | null>(null);
  const visibleActions = useMemo(
    () =>
      actions.filter(
        (action) => isAuthenticated || !requiresAuthAction(action),
      ),
    [actions, isAuthenticated],
  );

  useEffect(() => {
    let disposed = false;
    const authenticated = hasAuthSession();
    setIsAuthenticated(authenticated);

    if (!authenticated) {
      setProfile(null);
      return;
    }

    const loadProfile = async () => {
      try {
        const user = await getMeApi<AuthUser>();
        if (!disposed) {
          setProfile(user);
        }
      } catch {
        if (!disposed) {
          setProfile(null);
        }
      }
    };

    loadProfile().catch(() => undefined);
    const handleAvatarUpdated = (event: Event) => {
      const updatedProfile = (event as CustomEvent<AuthUser>).detail;
      if (updatedProfile) {
        setProfile(updatedProfile);
        return;
      }

      loadProfile().catch(() => undefined);
    };

    globalThis.addEventListener("falzo:avatar-updated", handleAvatarUpdated);

    return () => {
      disposed = true;
      globalThis.removeEventListener(
        "falzo:avatar-updated",
        handleAvatarUpdated,
      );
    };
  }, []);

  const profileName = getAuthUserDisplayName(profile, commonCopy.account);
  const avatarUrl = readAuthUserText(profile, ["avatar_url", "avatarUrl"]);
  const isExploreActive = isActiveRoute(pathname, ROUTES.explore);
  const isLocationsActive = isActiveRoute(pathname, ROUTES.locations);
  const isUploadActive = isActiveRoute(pathname, ROUTES.upload);
  const isSavedActive = isActiveRoute(pathname, ROUTES.saved);
  const accountRoute = isAuthenticated ? ROUTES.profile : ROUTES.login;
  const isAccountActive = isActiveRoute(pathname, accountRoute);

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
          <LanguageSwitcher />
          {visibleActions.map((action) => (
            <TopbarActionButton action={action} key={action.id} />
          ))}
        </div>
      </div>

      {showMobileNav ? (
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
              <div className="flex shrink-0 items-center gap-2">
                <LanguageSwitcher compact />
                {meta ? <div className="shrink-0">{meta}</div> : null}
              </div>
            </div>

            <div
              className={cn(
                "grid items-center gap-1",
                isAuthenticated ? "grid-cols-5" : "grid-cols-3",
              )}
            >
              <Link
                aria-current={isExploreActive ? "page" : undefined}
                aria-label={commonCopy.explore}
                className={mobileNavItemClass(isExploreActive)}
                href={ROUTES.explore}
              >
                <Camera className="size-4" />
                <span>{commonCopy.explore}</span>
              </Link>
              <Link
                aria-current={isLocationsActive ? "page" : undefined}
                aria-label={commonCopy.places}
                className={mobileNavItemClass(isLocationsActive)}
                href={ROUTES.locations}
              >
                <MapIcon className="size-4" />
                <span>{commonCopy.places}</span>
              </Link>
              {isAuthenticated ? (
                <>
                  <Link
                    aria-current={isUploadActive ? "page" : undefined}
                    aria-label={commonCopy.upload}
                    className={mobileUploadNavItemClass(isUploadActive)}
                    href={ROUTES.upload}
                  >
                    <Plus className="size-4" />
                    <span>{commonCopy.upload}</span>
                  </Link>
                  <Link
                    aria-current={isSavedActive ? "page" : undefined}
                    aria-label={commonCopy.saved}
                    className={mobileNavItemClass(isSavedActive)}
                    href={ROUTES.saved}
                  >
                    <Bookmark className="size-4" />
                    <span>{commonCopy.saved}</span>
                  </Link>
                </>
              ) : null}
              <Link
                aria-current={isAccountActive ? "page" : undefined}
                aria-label={isAuthenticated ? commonCopy.profile : commonCopy.login}
                className={mobileNavItemClass(isAccountActive)}
                href={accountRoute}
              >
                {isAuthenticated ? (
                  <RainbowAvatar
                    alt={profileName}
                    fallback={getAuthUserInitials(profileName)}
                    size="sm"
                    src={avatarUrl}
                  />
                ) : (
                  <LogIn className="size-4" />
                )}
                <span>
                  {isAuthenticated ? commonCopy.account : commonCopy.login}
                </span>
              </Link>
            </div>
          </div>
        </nav>
      ) : null}
    </>
  );
}
