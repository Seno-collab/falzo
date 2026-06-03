"use client";

import {
  createContext,
  useContext,
  useCallback,
  useEffect,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";
import { messages, type AppMessages, type SupportedLocale } from "@/i18n/messages";
import { defaultLocale, persistLocale, readCurrentLocale } from "@/i18n/locale";

type LocaleContextValue = {
  locale: SupportedLocale;
  messages: AppMessages;
  setLocale: (locale: SupportedLocale) => void;
};

const LocaleContext = createContext<LocaleContextValue | null>(null);

export function LocaleProvider({ children }: Readonly<PropsWithChildren>) {
  const [locale, setLocaleState] = useState<SupportedLocale>(defaultLocale);

  useEffect(() => {
    setLocaleState(readCurrentLocale());
  }, []);

  useEffect(() => {
    document.documentElement.lang = locale;
    persistLocale(locale);
  }, [locale]);

  const setLocale = useCallback((nextLocale: SupportedLocale) => {
    persistLocale(nextLocale);
    setLocaleState(nextLocale);
  }, []);

  const value = useMemo<LocaleContextValue>(
    () => ({
      locale,
      messages: messages[locale],
      setLocale,
    }),
    [locale, setLocale],
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
