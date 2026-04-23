import { cn } from "@/lib/utils";

export function TagBadge({ label, className }: { label: string; className?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border border-[#d6e3f1] bg-[#f7fbff] px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.05em] text-[#446487]",
        className,
      )}
    >
      {label}
    </span>
  );
}
