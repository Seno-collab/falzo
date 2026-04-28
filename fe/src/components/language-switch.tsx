import { Globe } from "lucide-react";
import { useLanguage, type AppLanguage } from "@/app/language-provider";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

function LanguageOption({
  target,
  active,
  onSelect,
}: Readonly<{
  target: AppLanguage;
  active: boolean;
  onSelect: (language: AppLanguage) => void;
}>) {
  return (
    <Button
      aria-pressed={active}
      className={cn(
        "h-7 min-w-10 rounded-md px-2 text-[11px] font-semibold shadow-none",
        active
          ? "border-[#7fa8d3] bg-[#255c99] text-white hover:bg-[#215188]"
          : "border-[#c2d7ec] bg-white/90 text-[#355883] hover:bg-[#eff6ff]",
      )}
      onClick={() => onSelect(target)}
      size="xs"
      type="button"
      variant="outline"
    >
      {target.toUpperCase()}
    </Button>
  );
}

export function LanguageSwitch() {
  const { language, setLanguage } = useLanguage();

  return (
    <div className="fixed top-20 right-3 z-[70] sm:top-4 sm:right-4">
      <div className="inline-flex items-center gap-1 rounded-xl border border-white/70 bg-white/86 px-2 py-1 shadow-[0_20px_36px_-24px_rgba(20,50,85,0.7)] backdrop-blur-md">
        <Globe className="mx-0.5 size-3.5 text-[#3f648f]" />
        <LanguageOption
          active={language === "vi"}
          onSelect={setLanguage}
          target="vi"
        />
        <LanguageOption
          active={language === "en"}
          onSelect={setLanguage}
          target="en"
        />
      </div>
    </div>
  );
}
