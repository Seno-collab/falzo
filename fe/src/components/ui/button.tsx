import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { Slot } from "radix-ui"

import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex shrink-0 items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-semibold transition-all duration-200 outline-none focus-visible:ring-[3px] focus-visible:ring-ring/45 focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-55 active:translate-y-px [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground shadow-[0_12px_24px_-14px_rgb(35_89_151/0.75)] hover:bg-primary/92",
        gradient:
          "border border-[#2760a2]/20 bg-[linear-gradient(135deg,#245694_0%,#2f6fb8_55%,#4f8ec8_100%)] text-white shadow-[0_16px_30px_-15px_rgb(29_83_144/0.78)] hover:brightness-105",
        destructive:
          "bg-destructive text-white shadow-[0_14px_28px_-16px_rgb(183_28_28/0.8)] hover:bg-destructive/92",
        outline:
          "border border-border bg-white/90 text-foreground shadow-[0_8px_22px_-16px_rgb(18_53_90/0.36)] hover:bg-accent/70",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary/82",
        soft: "border border-transparent bg-muted text-secondary-foreground hover:bg-muted/80",
        ghost: "text-foreground hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-10 px-4 py-2 has-[>svg]:px-3",
        xs: "h-7 gap-1 rounded-md px-2.5 text-xs has-[>svg]:px-2",
        sm: "h-9 gap-1.5 rounded-lg px-3.5 has-[>svg]:px-3",
        lg: "h-11 rounded-lg px-6 has-[>svg]:px-4",
        icon: "size-10",
        "icon-xs": "size-7 rounded-md [&_svg:not([class*='size-'])]:size-3.5",
        "icon-sm": "size-9",
        "icon-lg": "size-11",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
)

function Button({
  className,
  variant,
  size,
  asChild = false,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean
  }) {
  const Comp = asChild ? Slot.Root : "button"

  return (
    <Comp
      data-slot="button"
      data-variant={variant}
      data-size={size}
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    />
  )
}

export { Button, buttonVariants }
