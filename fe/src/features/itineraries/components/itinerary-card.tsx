import { CalendarDays, MapPinned, Route, Wallet } from "lucide-react";
import Link from "next/link";
import type { ItineraryListItem } from "@/features/itineraries/types";
import { ROUTES } from "@/lib/routes";

function formatMoney(value: number) {
  return new Intl.NumberFormat("vi-VN").format(value);
}

function formatBudget(item: ItineraryListItem) {
  if (item.budgetMin === 0 && item.budgetMax === 0) {
    return "Miễn phí";
  }

  return `${formatMoney(item.budgetMin)} - ${formatMoney(item.budgetMax)}đ`;
}

function splitStyles(value: string) {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean)
    .slice(0, 4);
}

export function ItineraryCard({ item }: Readonly<{ item: ItineraryListItem }>) {
  const styles = splitStyles(item.travelStyle);

  return (
    <Link
      className="group overflow-hidden rounded-lg border border-[#d8e5e0] bg-white shadow-[0_16px_42px_-34px_rgb(22_44_40/0.72)] transition hover:-translate-y-0.5 hover:border-[#9fc7ba] hover:shadow-[0_24px_58px_-42px_rgb(22_44_40/0.78)]"
      href={ROUTES.itineraryDetail(item.slug)}
    >
      <div className="relative aspect-[16/9] bg-[#e8efe9]">
        {item.coverImageUrl ? (
          <img
            alt={item.title}
            className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.035]"
            decoding="async"
            loading="lazy"
            src={item.coverImageUrl}
          />
        ) : (
          <div className="flex h-full items-center justify-center text-[#6b7d77]">
            <Route className="size-9" />
          </div>
        )}
        <span className="absolute left-3 top-3 rounded-full bg-white/92 px-3 py-1 text-xs font-semibold text-[#18342f] shadow-sm">
          {item.province}
        </span>
      </div>

      <div className="space-y-4 p-4">
        <div className="space-y-2">
          <h2 className="text-lg font-semibold leading-6 text-[#142c27]">
            {item.title}
          </h2>
          <div className="grid grid-cols-2 gap-2 text-xs font-semibold">
            <span className="inline-flex items-center gap-1.5 rounded-lg bg-[#f2f7f5] px-2.5 py-2 text-[#46645f]">
              <CalendarDays className="size-3.5" />
              {item.durationDays} ngày
            </span>
            <span className="inline-flex items-center gap-1.5 rounded-lg bg-[#fff6e8] px-2.5 py-2 text-[#7a5a18]">
              <Wallet className="size-3.5" />
              {formatBudget(item)}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-2 text-sm font-semibold text-[#506865]">
          <MapPinned className="size-4 text-[#397168]" />
          {item.stopCount} điểm dừng
          {item.transportation ? ` · ${item.transportation}` : ""}
        </div>

        {styles.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {styles.map((style) => (
              <span
                className="rounded-full bg-[#eef7f3] px-2.5 py-1 text-xs font-semibold text-[#315d55]"
                key={style}
              >
                {style}
              </span>
            ))}
          </div>
        ) : null}
      </div>
    </Link>
  );
}
