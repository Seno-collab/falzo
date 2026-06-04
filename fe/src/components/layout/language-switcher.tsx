"use client";

import { Languages } from "lucide-react";
import { useI18n } from "@/i18n/locale-provider";
import type { SupportedLocale } from "@/i18n/messages";
import { cn } from "@/lib/utils";

const localeOptions: Array<{
  locale: SupportedLocale;
  label: string;
  shortLabel: string;
}> = [
  { locale: "en", label: "English", shortLabel: "EN" },
  { locale: "vi", label: "Tiếng Việt", shortLabel: "VI" },
];

export function LanguageSwitcher({
  className,
  compact = false,
}: Readonly<{ className?: string; compact?: boolean }>) {
  const { locale, messages, setLocale } = useI18n();

  return (
    <div
      aria-label={messages.common.language}
      className={cn(
        "inline-flex h-9 max-w-full shrink-0 items-center gap-1 rounded-full border border-black/8 bg-white/86 p-1 text-xs font-bold text-[#50647d] shadow-sm",
        compact ? "w-fit" : "",
        className,
      )}
      role="group"
    >
      <Languages
        className={cn("size-3.5 text-[#7892ad]", compact ? "ml-0.5" : "ml-1")}
      />
      {localeOptions.map((option) => {
        const active = locale === option.locale;

        return (
          <button
            aria-pressed={active}
            className={cn(
              "h-7 rounded-full transition",
              compact ? "min-w-8 px-2 text-center" : "px-2.5",
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
            {compact ? option.shortLabel : option.label}
          </button>
        );
      })}
    </div>
  );
}
