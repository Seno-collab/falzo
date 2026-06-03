import { enMessages } from "@/i18n/messages/en";
import { viMessages } from "@/i18n/messages/vi";

export const messages = {
  en: enMessages,
  vi: viMessages,
} as const;

export type SupportedLocale = keyof typeof messages;
export type AppMessages = (typeof messages)[SupportedLocale];
