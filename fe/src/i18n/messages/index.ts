import { enMessages } from "@/i18n/messages/en";

export const messages = {
  en: enMessages,
} as const;

export type SupportedLocale = keyof typeof messages;
export type AppMessages = (typeof messages)[SupportedLocale];
