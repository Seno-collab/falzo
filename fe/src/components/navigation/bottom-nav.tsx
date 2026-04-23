"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { navigationItems } from "@/components/navigation/nav-config";
import { cn } from "@/lib/utils";

export function BottomNav() {
  const pathname = usePathname();

  return (
    <nav className="fixed inset-x-3 bottom-3 z-50 md:hidden">
      <div className="surface grid grid-cols-5 gap-1 px-2 py-2">
        {navigationItems.map((item) => {
          const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
          const Icon = item.icon;

          return (
            <Link
              className={cn(
                "flex flex-col items-center justify-center rounded-xl px-1 py-2 text-[11px] font-semibold transition",
                active ? "bg-[#eaf2ff] text-[#1f6fe5]" : "text-[#6481a4] hover:bg-[#f2f7ff]",
              )}
              href={item.href}
              key={item.href}
            >
              <Icon className="mb-1 size-4" />
              {item.label}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
