import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export function SectionHeading({
  kicker,
  title,
  description,
  action,
  className,
}: {
  kicker?: string;
  title: string;
  description?: string;
  action?: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex flex-wrap items-end justify-between gap-3",
        className,
      )}
    >
      <div className="space-y-1.5">
        {kicker ? <p className="app-kicker">{kicker}</p> : null}
        <h2 className="falzo-display text-2xl font-semibold tracking-tight text-[#173457] sm:text-3xl">
          {title}
        </h2>
        {description ? (
          <p className="app-subtitle max-w-3xl">{description}</p>
        ) : null}
      </div>
      {action ? <div className="shrink-0">{action}</div> : null}
    </div>
  );
}
