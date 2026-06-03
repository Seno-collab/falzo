"use client";

import { Languages } from "lucide-react";
import { useI18n } from "@/i18n/locale-provider";
import type { SupportedLocale } from "@/i18n/messages";
import { cn } from "@/lib/utils";

const localeOptions: Array<{ locale: SupportedLocale; label: string }> = [
  { locale: "en", label: "English" },
  { locale: "vi", label: "Tiếng Việt" },
];

export function LanguageSwitcher({
  className,
}: Readonly<{ className?: string }>) {
  const { locale, messages, setLocale } = useI18n();

  return (
    <div
      aria-label={messages.common.language}
      className={cn(
        "inline-flex h-9 shrink-0 items-center gap-1 rounded-full border border-black/8 bg-white/86 p-1 text-xs font-bold text-[#50647d] shadow-sm",
        className,
      )}
      role="group"
    >
      <Languages className="ml-1 size-3.5 text-[#7892ad]" />
      {localeOptions.map((option) => {
        const active = locale === option.locale;

        return (
          <button
            aria-pressed={active}
            className={cn(
              "h-7 rounded-full px-2.5 transition",
              active
                ? "bg-[#111] text-white"
                : "text-[#5f7894] hover:bg-[#f2f7fd] hover:text-[#143052]",
            )}
            key={option.locale}
            onClick={() => setLocale(option.locale)}
            title={
              option.locale === "en"
                ? messages.common.english
                : messages.common.vietnamese
            }
            type="button"
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}
