import { Menu } from "lucide-react";
import type { ReactNode } from "react";
import Link from "next/link";
import type { VariantProps } from "class-variance-authority";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { cn } from "@/lib/utils";

type ActionVariant = VariantProps<typeof buttonVariants>["variant"];

type TopbarAction = {
  id: string;
  label: string;
  to?: string;
  onClick?: () => void;
  icon?: ReactNode;
  variant?: ActionVariant;
  disabled?: boolean;
};

function TopbarActionButton({
  action,
  fullWidth = false,
}: {
  action: TopbarAction;
  fullWidth?: boolean;
}) {
  const classes = cn(fullWidth ? "w-full justify-start" : "");

  if (action.to && !action.disabled) {
    return (
      <Button
        asChild
        className={classes}
        size={fullWidth ? "default" : "sm"}
        variant={action.variant ?? "outline"}
      >
        <Link href={action.to}>
          {action.icon}
          {action.label}
        </Link>
      </Button>
    );
  }

  return (
    <Button
      className={classes}
      disabled={action.disabled}
      onClick={action.onClick}
      size={fullWidth ? "default" : "sm"}
      type="button"
      variant={action.variant ?? "outline"}
    >
      {action.icon}
      {action.label}
    </Button>
  );
}

export function AppTopbar({
  brand,
  brandIcon,
  subtitle,
  meta,
  actions,
  mobileMenuTitle,
}: {
  brand: string;
  brandIcon?: ReactNode;
  subtitle?: string;
  meta?: ReactNode;
  actions: TopbarAction[];
  mobileMenuTitle: string;
}) {
  return (
    <div className="app-topbar-panel">
      <div className="min-w-0 space-y-1">
        <div className="app-brand">
          <span className="app-brand-dot">{brandIcon}</span>
          <span className="truncate">{brand}</span>
        </div>
        {subtitle ? (
          <p className="mt-1 hidden text-xs text-[#5a7aa2] sm:block">
            {subtitle}
          </p>
        ) : null}
        {meta ? <div className="hidden sm:block">{meta}</div> : null}
      </div>

      <div className="hidden items-center gap-2 sm:flex">
        {actions.map((action) => (
          <TopbarActionButton action={action} key={action.id} />
        ))}
      </div>

      <Sheet>
        <SheetTrigger asChild>
          <Button
            className="sm:hidden"
            size="icon-sm"
            type="button"
            variant="outline"
          >
            <Menu className="size-4" />
            <span className="sr-only">Open menu</span>
          </Button>
        </SheetTrigger>
        <SheetContent side="right">
          <SheetHeader>
            <SheetTitle>{mobileMenuTitle}</SheetTitle>
            <SheetDescription>{brand}</SheetDescription>
          </SheetHeader>
          {meta ? <div className="px-5">{meta}</div> : null}
          <div className="space-y-2 px-5">
            {actions.map((action) => (
              <TopbarActionButton
                action={action}
                fullWidth
                key={`mobile-${action.id}`}
              />
            ))}
          </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}
