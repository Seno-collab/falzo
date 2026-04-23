import * as React from "react"

import { cn } from "@/lib/utils"

function Textarea({ className, ...props }: React.ComponentProps<"textarea">) {
  return (
    <textarea
      data-slot="textarea"
      className={cn(
        "field-sizing-content min-h-24 w-full rounded-lg border border-input bg-white/90 px-3.5 py-2.5 text-sm text-foreground shadow-[0_1px_2px_rgb(12_31_56/0.03)] outline-none transition-[border-color,box-shadow,background-color] placeholder:text-muted-foreground/95 disabled:cursor-not-allowed disabled:opacity-55",
        "hover:border-[#b5cae2] focus-visible:border-ring focus-visible:bg-white focus-visible:ring-[3px] focus-visible:ring-ring/40",
        "aria-invalid:border-destructive aria-invalid:ring-[3px] aria-invalid:ring-destructive/20",
        className,
      )}
      {...props}
    />
  )
}

export { Textarea }
