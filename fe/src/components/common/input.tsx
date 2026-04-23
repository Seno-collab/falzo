import { cn } from "@/lib/utils";

type InputProps = React.InputHTMLAttributes<HTMLInputElement>;

export function Input({ className, ...props }: InputProps) {
  return (
    <input
      className={cn(
        "h-11 w-full rounded-xl border border-[#d2deec] bg-white px-3.5 text-sm text-[#10243e] outline-none transition-all placeholder:text-[#91a2bb]",
        "focus:border-[#7ca2d1] focus:ring-2 focus:ring-[#dbe9ff]",
        className,
      )}
      {...props}
    />
  );
}
