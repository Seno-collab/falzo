import { Clock3, Globe } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useLanguage } from "@/app/language-provider";

export function UserPresenceBadge() {
  const { appLanguage } = useLanguage();
  const [now, setNow] = useState<Date | null>(null);
  const [timezone, setTimezone] = useState("");

  useEffect(() => {
    setNow(new Date());
    setTimezone(Intl.DateTimeFormat().resolvedOptions().timeZone);

    const timer = globalThis.setInterval(() => {
      setNow(new Date());
    }, 60_000);

    return () => {
      globalThis.clearInterval(timer);
    };
  }, []);

  const time = useMemo(
    () => {
      if (!now) {
        return "--:--";
      }

      return new Intl.DateTimeFormat(appLanguage === "vi" ? "vi-VN" : "en-US", {
        hour: "2-digit",
        minute: "2-digit",
      }).format(now);
    },
    [appLanguage, now],
  );

  return (
    <div className="inline-flex flex-wrap items-center gap-1.5">
      <span className="inline-flex items-center gap-1 rounded-full border border-[#c9ddf0] bg-[#f5faff] px-2 py-0.5 text-[10px] font-semibold tracking-[0.08em] text-[#446b96] uppercase">
        <Globe className="size-3" />
        {appLanguage.toUpperCase()}
      </span>
      <span className="inline-flex items-center gap-1 rounded-full border border-[#c9ddf0] bg-white px-2 py-0.5 text-[10px] font-semibold tracking-[0.06em] text-[#446b96] uppercase">
        <Clock3 className="size-3" />
        {time}
      </span>
      {timezone ? (
        <span className="hidden text-[10px] font-medium tracking-[0.06em] text-[#6587b1] uppercase sm:inline">
          {timezone.replaceAll("_", " ")}
        </span>
      ) : null}
    </div>
  );
}
