import { cn } from "@/lib/utils";

export function SkeletonLoader({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-xl bg-[#e6eef8]",
        "before:absolute before:inset-0 before:-translate-x-full before:animate-[shimmer_1.6s_linear_infinite] before:bg-[linear-gradient(90deg,transparent,rgba(255,255,255,0.7),transparent)]",
        className,
      )}
    />
  );
}
