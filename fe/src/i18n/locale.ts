import type { SupportedLocale } from "@/i18n/messages";

export const localeStorageKey = "falzo.locale";
export const defaultLocale: SupportedLocale = "vi";

export function isSupportedLocale(
  value: string | null | undefined,
): value is SupportedLocale {
  return value === "en" || value === "vi";
}

export function readCurrentLocale(): SupportedLocale {
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

export function persistLocale(locale: SupportedLocale) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(localeStorageKey, locale);
}
