import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { Slot } from "radix-ui"

import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex w-fit shrink-0 items-center justify-center gap-1 overflow-hidden rounded-full border px-2.5 py-1 text-[11px] font-semibold tracking-wide whitespace-nowrap transition-colors focus-visible:ring-[3px] focus-visible:ring-ring/40 [&>svg]:pointer-events-none [&>svg]:size-3",
  {
    variants: {
      variant: {
        default:
          "border-[#bfd5ed] bg-[#ecf5ff] text-[#25598f] [a&]:hover:bg-[#dfedfd]",
        secondary:
          "border-border/90 bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/84",
        destructive:
          "border-transparent bg-destructive text-white [a&]:hover:bg-destructive/92",
        outline:
          "border-border bg-white text-foreground [a&]:hover:bg-accent/65",
        ghost: "border-transparent bg-transparent text-muted-foreground [a&]:hover:bg-accent/65",
        link: "border-transparent bg-transparent px-0 text-primary underline-offset-4 [a&]:hover:underline",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
)

function Badge({
  className,
  variant,
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & {
    asChild?: boolean
  }) {
  const Comp = asChild ? Slot.Root : "span"

  return (
    <Comp
      data-slot="badge"
      data-variant={variant}
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  )
}

export { Badge, badgeVariants }
