"use client";

import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";
import { messages, type AppMessages, type SupportedLocale } from "@/i18n/messages";

const localeStorageKey = "falzo.locale";
const defaultLocale: SupportedLocale = "en";

type LocaleContextValue = {
  locale: SupportedLocale;
  messages: AppMessages;
  setLocale: (locale: SupportedLocale) => void;
};

const LocaleContext = createContext<LocaleContextValue | null>(null);

function isSupportedLocale(value: string | null): value is SupportedLocale {
  return value === "en" || value === "vi";
}

function readInitialLocale(): SupportedLocale {
  if (typeof window === "undefined") {
    return defaultLocale;
  }

  const storedLocale = window.localStorage.getItem(localeStorageKey);
  if (isSupportedLocale(storedLocale)) {
    return storedLocale;
  }

  const browserLocale = window.navigator.language.toLowerCase();
  return browserLocale.startsWith("vi") ? "vi" : defaultLocale;
}

export function LocaleProvider({ children }: Readonly<PropsWithChildren>) {
  const [locale, setLocaleState] = useState<SupportedLocale>(defaultLocale);

  useEffect(() => {
    setLocaleState(readInitialLocale());
  }, []);

  useEffect(() => {
    document.documentElement.lang = locale;
    window.localStorage.setItem(localeStorageKey, locale);
  }, [locale]);

  const value = useMemo<LocaleContextValue>(
    () => ({
      locale,
      messages: messages[locale],
      setLocale: setLocaleState,
    }),
    [locale],
  );

  return (
    <LocaleContext.Provider value={value}>{children}</LocaleContext.Provider>
  );
}

export function useI18n() {
  const context = useContext(LocaleContext);
  if (!context) {
    throw new Error("useI18n must be used inside LocaleProvider.");
  }

  return context;
}
