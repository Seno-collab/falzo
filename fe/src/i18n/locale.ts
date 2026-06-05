import type { SupportedLocale } from "@/i18n/messages";

export const localeStorageKey = "falzo.locale";
export const defaultLocale: SupportedLocale = "vi";

export function isSupportedLocale(
  value: string | null | undefined,
): value is SupportedLocale {
  return value === "en" || value === "vi";
}

export function readCurrentLocale(): SupportedLocale {
  if (globalThis.localStorage === undefined) {
    return defaultLocale;
  }

  const storedLocale = globalThis.localStorage.getItem(localeStorageKey);
  if (isSupportedLocale(storedLocale)) {
    return storedLocale;
  }

  const browserLocale = globalThis.navigator.language.toLowerCase();
  return browserLocale.startsWith("vi") ? "vi" : defaultLocale;
}

export function persistLocale(locale: SupportedLocale) {
  if (globalThis.localStorage === undefined) {
    return;
  }

  globalThis.localStorage.setItem(localeStorageKey, locale);
}
