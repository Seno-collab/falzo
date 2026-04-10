import {
  createContext,
  useContext,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";

export type AppLanguage = "vi" | "en";

type LanguageContextValue = {
  language: AppLanguage;
  isVietnamese: boolean;
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
  const hasVietnamTimezone = timezone
    ? VIETNAM_TIMEZONES.has(timezone)
    : false;

  return hasVietnameseLocale || hasVietnamTimezone ? "vi" : "en";
}

export function LanguageProvider({ children }: PropsWithChildren) {
  const [language] = useState<AppLanguage>(() => detectLanguageFromBrowser());

  const value = useMemo(
    () => ({
      language,
      isVietnamese: language === "vi",
    }),
    [language],
  );

  return (
    <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>
  );
}

export function useLanguage() {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error("useLanguage must be used within LanguageProvider");
  }

  return context;
}
