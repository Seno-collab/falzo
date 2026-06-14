"use client";

import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import {
  getMeApi,
  hasAuthSession,
  subscribeAuthSessionChanged,
} from "@/features/auth/api";
import type { AuthUser } from "@/features/auth/types";
import { readAuthUserText } from "@/features/auth/user-display";

type CursorState = {
  x: number;
  y: number;
  visible: boolean;
  pressed: boolean;
  hiddenForText: boolean;
};

const editableSelector =
  'input, textarea, select, [contenteditable="true"], [role="textbox"]';

function isEditableTarget(target: EventTarget | null) {
  if (!(target instanceof Element)) {
    return false;
  }

  return Boolean(target.closest(editableSelector));
}

export function CharacterCursor() {
  const pathname = usePathname();
  const [enabled, setEnabled] = useState(false);
  const [authenticated, setAuthenticated] = useState(false);
  const [avatarUrl, setAvatarUrl] = useState<string | null>(null);
  const [avatarLoadFailed, setAvatarLoadFailed] = useState(false);
  const [cursor, setCursor] = useState<CursorState>({
    x: 0,
    y: 0,
    visible: false,
    pressed: false,
    hiddenForText: false,
  });

  useEffect(() => {
    const mediaQuery = globalThis.matchMedia("(pointer: fine)");
    const updateEnabled = () => setEnabled(mediaQuery.matches);

    updateEnabled();
    mediaQuery.addEventListener("change", updateEnabled);

    return () => mediaQuery.removeEventListener("change", updateEnabled);
  }, []);

  useEffect(() => {
    if (!enabled) {
      document.documentElement.classList.remove(
        "falzo-character-cursor-enabled",
      );
      return;
    }

    document.documentElement.classList.add("falzo-character-cursor-enabled");

    return () => {
      document.documentElement.classList.remove(
        "falzo-character-cursor-enabled",
      );
    };
  }, [enabled]);

  useEffect(() => {
    let disposed = false;

    const applyProfile = (profile: AuthUser | null) => {
      if (disposed) {
        return;
      }

      const nextAvatarUrl = readAuthUserText(profile, [
        "avatar_url",
        "avatarUrl",
      ]);
      setAvatarUrl(nextAvatarUrl);
      setAvatarLoadFailed(false);
    };

    const refreshAuthState = () => {
      const nextAuthenticated = hasAuthSession();
      setAuthenticated(nextAuthenticated);

      if (!nextAuthenticated) {
        applyProfile(null);
        return;
      }

      getMeApi<AuthUser>()
        .then((profile) => applyProfile(profile))
        .catch(() => applyProfile(null));
    };

    const handleAvatarUpdated = (event: Event) => {
      applyProfile((event as CustomEvent<AuthUser>).detail ?? null);
    };

    refreshAuthState();
    const unsubscribe = subscribeAuthSessionChanged(refreshAuthState);
    globalThis.addEventListener("falzo:avatar-updated", handleAvatarUpdated);
    globalThis.addEventListener("focus", refreshAuthState);
    globalThis.addEventListener("storage", refreshAuthState);

    return () => {
      disposed = true;
      unsubscribe();
      globalThis.removeEventListener(
        "falzo:avatar-updated",
        handleAvatarUpdated,
      );
      globalThis.removeEventListener("focus", refreshAuthState);
      globalThis.removeEventListener("storage", refreshAuthState);
    };
  }, [pathname]);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    const handlePointerMove = (event: PointerEvent) => {
      if (event.pointerType !== "mouse") {
        return;
      }

      setCursor((current) => ({
        ...current,
        x: event.clientX,
        y: event.clientY,
        visible: true,
        hiddenForText: isEditableTarget(event.target),
      }));
    };
    const handlePointerDown = (event: PointerEvent) => {
      if (event.pointerType === "mouse") {
        setCursor((current) => ({ ...current, pressed: true }));
      }
    };
    const handlePointerUp = () => {
      setCursor((current) => ({ ...current, pressed: false }));
    };
    const handleMouseLeave = () => {
      setCursor((current) => ({ ...current, visible: false }));
    };
    const handleMouseEnter = () => {
      setCursor((current) => ({ ...current, visible: true }));
    };

    globalThis.addEventListener("pointermove", handlePointerMove);
    globalThis.addEventListener("pointerdown", handlePointerDown);
    globalThis.addEventListener("pointerup", handlePointerUp);
    document.documentElement.addEventListener("mouseleave", handleMouseLeave);
    document.documentElement.addEventListener("mouseenter", handleMouseEnter);

    return () => {
      globalThis.removeEventListener("pointermove", handlePointerMove);
      globalThis.removeEventListener("pointerdown", handlePointerDown);
      globalThis.removeEventListener("pointerup", handlePointerUp);
      document.documentElement.removeEventListener(
        "mouseleave",
        handleMouseLeave,
      );
      document.documentElement.removeEventListener(
        "mouseenter",
        handleMouseEnter,
      );
    };
  }, [enabled]);

  if (!enabled || !cursor.visible || cursor.hiddenForText) {
    return null;
  }

  return (
    <div
      aria-hidden="true"
      className={[
        "falzo-character-cursor",
        authenticated
          ? "falzo-character-cursor--traveler"
          : "falzo-character-cursor--guest",
        authenticated && avatarUrl && !avatarLoadFailed
          ? "falzo-character-cursor--photo"
          : "",
        cursor.pressed ? "falzo-character-cursor--pressed" : "",
      ].join(" ")}
      style={{
        transform: `translate3d(${cursor.x}px, ${cursor.y}px, 0)`,
      }}
    >
      <span className="falzo-character-cursor__pointer" />
      <span className="falzo-character-cursor__avatar">
        {authenticated && avatarUrl && !avatarLoadFailed ? (
          <img
            alt=""
            className="falzo-character-cursor__photo"
            decoding="async"
            onError={() => setAvatarLoadFailed(true)}
            src={avatarUrl}
          />
        ) : (
          <>
            <span className="falzo-character-cursor__head" />
            <span className="falzo-character-cursor__body" />
            <span className="falzo-character-cursor__pack" />
          </>
        )}
      </span>
      <span className="falzo-character-cursor__label">
        {authenticated ? "Player" : "Guest"}
      </span>
    </div>
  );
}
