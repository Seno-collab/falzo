import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-xl px-4 py-2 text-sm font-semibold transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        primary:
          "bg-[#1f6fe5] text-white shadow-[0_14px_28px_-16px_rgba(19,78,166,0.72)] hover:bg-[#185bc1] focus-visible:ring-[#1f6fe5]",
        secondary: "bg-white text-[#173353] border border-[#d2deec] hover:bg-[#f7fbff] focus-visible:ring-[#91afd3]",
        ghost: "text-[#1d3f67] hover:bg-[#eaf2ff] focus-visible:ring-[#91afd3]",
      },
      size: {
        sm: "h-9 px-3",
        md: "h-10 px-4",
        lg: "h-11 px-5",
        icon: "h-10 w-10 px-0",
      },
    },
    defaultVariants: {
      variant: "primary",
      size: "md",
    },
  },
);

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> &
  VariantProps<typeof buttonVariants>;

export function Button({ className, variant, size, ...props }: ButtonProps) {
  return <button className={cn(buttonVariants({ variant, size, className }))} {...props} />;
}
