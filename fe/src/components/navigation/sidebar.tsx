"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { navigationItems } from "@/components/navigation/nav-config";
import { cn } from "@/lib/utils";

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="surface fixed top-4 bottom-4 left-4 hidden w-64 flex-col p-4 md:flex">
      <div className="mb-6">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[#6e88aa]">Travel Discovery</p>
        <h1 className="mt-1 font-[var(--font-sora)] text-xl font-semibold tracking-tight text-[#152f4f]">
          Explore Places
        </h1>
      </div>

      <nav className="space-y-1.5">
        {navigationItems.map((item) => {
          const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
          const Icon = item.icon;

          return (
            <Link
              className={cn(
                "flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-semibold transition",
                active ? "bg-[#eaf2ff] text-[#1f6fe5]" : "text-[#587493] hover:bg-[#f2f7ff]",
              )}
              href={item.href}
              key={item.href}
            >
              <Icon className="size-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="mt-auto rounded-xl border border-[#d6e3f1] bg-[#f7fbff] p-3">
        <p className="text-xs font-semibold text-[#355b87]">Tips</p>
        <p className="mt-1 text-xs leading-5 text-[#6a84a4]">
          Use map view on desktop for split layout and faster place comparison.
        </p>
      </div>
    </aside>
  );
}
