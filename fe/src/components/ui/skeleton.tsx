import { cn } from "@/lib/utils"

function Skeleton({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="skeleton"
      className={cn(
        "relative overflow-hidden rounded-lg bg-gradient-to-r from-[#edf3fb] via-[#f6f9ff] to-[#edf3fb]",
        "before:absolute before:inset-0 before:-translate-x-full before:animate-[shimmer_1.8s_linear_infinite] before:bg-[linear-gradient(100deg,transparent_20%,rgba(255,255,255,0.65)_50%,transparent_80%)]",
        className,
      )}
      {...props}
    />
  )
}

export { Skeleton }
