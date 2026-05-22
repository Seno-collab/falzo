import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

type RainbowAvatarSize = "xs" | "sm" | "md" | "lg";

const avatarSizeClass: Record<RainbowAvatarSize, string> = {
  xs: "size-6 p-0.5 text-[10px]",
  sm: "size-8 p-0.5 text-[11px]",
  md: "size-9 p-0.5 text-xs",
  lg: "size-20 p-0.75 text-lg",
};

export function RainbowAvatar({
  alt,
  className,
  fallback,
  size = "md",
  src,
}: Readonly<{
  alt: string;
  className?: string;
  fallback: ReactNode;
  size?: RainbowAvatarSize;
  src?: string | null;
}>) {
  return (
    <span
      className={cn(
        "relative inline-flex shrink-0 rounded-full font-semibold text-white shadow-[0_18px_38px_-22px_rgb(22_58_95/0.82),0_0_0_1px_rgb(158_189_221/0.45)]",
        avatarSizeClass[size],
        className,
      )}
    >
      <span
        aria-hidden="true"
        className="absolute inset-0 rounded-full bg-[conic-gradient(from_0deg,#ff3b30,#ff9500,#ffcc00,#34c759,#00c7be,#007aff,#af52de,#ff2d55,#ff3b30)] motion-safe:animate-[spin_4s_linear_infinite]"
      />
      <span className="relative z-10 flex h-full w-full items-center justify-center overflow-hidden rounded-full bg-[radial-gradient(circle_at_30%_25%,#7db8ff_0%,#2f6da8_42%,#17395c_100%)]">
        {src ? (
          <img alt={alt} className="h-full w-full object-cover" src={src} />
        ) : (
          fallback
        )}
      </span>
    </span>
  );
}
