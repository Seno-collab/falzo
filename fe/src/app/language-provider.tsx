"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";

export type AppLanguage = "vi" | "en";
const LANGUAGE_STORAGE_KEY = "falzo.language";
const SUPPORTED_LANGUAGES = new Set<AppLanguage>(["vi", "en"]);
type LanguageContextValue = {
  appLanguage: AppLanguage;
  isVietnamese: boolean;
  setAppLanguage: (language: AppLanguage) => void;
  toggleAppLanguage: () => void;
};

const VIETNAM_TIMEZONES = new Set(["Asia/Ho_Chi_Minh", "Asia/Saigon"]);
const LanguageContext = createContext<LanguageContextValue | null>(null);

function isVietnameseLocale(locale: string) {
  const normalized = locale.toLowerCase();
  return (
    normalized.startsWith("vi") ||
    normalized.includes("-vn") ||
    normalized.includes("_vn")
  );
}

function detectLanguageFromBrowser(): AppLanguage {
  if (globalThis.window === undefined) {
    return "en";
  }

  const browserLocales = Array.from(
    new Set([
      ...(globalThis.navigator.languages ?? []),
      globalThis.navigator.language,
    ]),
  ).filter((value): value is string => Boolean(value));

  const hasVietnameseLocale = browserLocales.some(isVietnameseLocale);
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
  const hasVietnamTimezone = timezone ? VIETNAM_TIMEZONES.has(timezone) : false;

  return hasVietnameseLocale || hasVietnamTimezone ? "vi" : "en";
}

function normalizeLanguage(value: unknown): AppLanguage | null {
  if (typeof value !== "string") {
    return null;
  }

  const normalized = value.toLowerCase();
  return SUPPORTED_LANGUAGES.has(normalized as AppLanguage)
    ? (normalized as AppLanguage)
    : null;
}

function getStoredLanguage(): AppLanguage | null {
  if (globalThis.window === undefined) {
    return null;
  }

  return normalizeLanguage(
    globalThis.window.localStorage.getItem(LANGUAGE_STORAGE_KEY),
  );
}

export function LanguageProvider({ children }: Readonly<PropsWithChildren>) {
  const [appLanguageState, setAppLanguageState] = useState<AppLanguage>(() => {
    return getStoredLanguage() ?? detectLanguageFromBrowser();
  });

  useEffect(() => {
    if (globalThis.window === undefined) {
      return;
    }

    globalThis.window.localStorage.setItem(
      LANGUAGE_STORAGE_KEY,
      appLanguageState,
    );
    document.documentElement.lang = appLanguageState;
  }, [appLanguageState]);

  const setAppLanguage = useCallback((nextLanguage: AppLanguage) => {
    setAppLanguageState(nextLanguage);
  }, []);

  const toggleAppLanguage = useCallback(() => {
    setAppLanguageState((previous) => (previous === "vi" ? "en" : "vi"));
  }, []);

  const value = useMemo(
    () => ({
      appLanguage: appLanguageState,
      isVietnamese: appLanguageState === "vi",
      setAppLanguage,
      toggleAppLanguage,
    }),
    [appLanguageState, setAppLanguage, toggleAppLanguage],
  );

  return (
    <LanguageContext.Provider value={value}>
      {children}
    </LanguageContext.Provider>
  );
}

export function useLanguage() {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error("useLanguage must be used within LanguageProvider");
  }

  return context;
}
