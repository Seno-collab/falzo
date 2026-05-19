import { CircleAlert } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export function EmptyState({
  title,
  description,
  icon,
  action,
  className,
}: Readonly<{
  title: string;
  description: string;
  icon?: ReactNode;
  action?: ReactNode;
  className?: string;
}>) {
  return (
    <div className={cn("app-empty-state", className)}>
      <span className="app-empty-icon">
        {icon ?? <CircleAlert className="size-5" />}
      </span>
      <div className="space-y-1">
        <h3 className="text-lg font-semibold tracking-normal text-[#1c3b61]">
          {title}
        </h3>
        <p className="text-sm leading-6 text-[#567396]">{description}</p>
      </div>
      {action ? <div>{action}</div> : null}
    </div>
  );
}
