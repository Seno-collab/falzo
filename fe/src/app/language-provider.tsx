"use client";

import {
  useCallback,
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";

export type AppLanguage = "vi" | "en";
const LANGUAGE_STORAGE_KEY = "falzo.language";
const SUPPORTED_LANGUAGES: AppLanguage[] = ["vi", "en"];

type LanguageContextValue = {
  language: AppLanguage;
  isVietnamese: boolean;
  setLanguage: (language: AppLanguage) => void;
  toggleLanguage: () => void;
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
  if (typeof window === "undefined") {
    return "en";
  }

  const browserLocales = Array.from(
    new Set([...(window.navigator.languages ?? []), window.navigator.language]),
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
  return SUPPORTED_LANGUAGES.includes(normalized as AppLanguage)
    ? (normalized as AppLanguage)
    : null;
}

function getStoredLanguage(): AppLanguage | null {
  if (typeof window === "undefined") {
    return null;
  }

  return normalizeLanguage(window.localStorage.getItem(LANGUAGE_STORAGE_KEY));
}

export function LanguageProvider({ children }: PropsWithChildren) {
  const [language, setLanguageState] = useState<AppLanguage>(() => {
    return getStoredLanguage() ?? detectLanguageFromBrowser();
  });

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, language);
    document.documentElement.lang = language;
  }, [language]);

  const setLanguage = useCallback((nextLanguage: AppLanguage) => {
    setLanguageState(nextLanguage);
  }, []);

  const toggleLanguage = useCallback(() => {
    setLanguageState((previous) => (previous === "vi" ? "en" : "vi"));
  }, []);

  const value = useMemo(
    () => ({
      language,
      isVietnamese: language === "vi",
      setLanguage,
      toggleLanguage,
    }),
    [language, setLanguage, toggleLanguage],
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
