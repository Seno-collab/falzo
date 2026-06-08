import {
  ArrowRight,
  CalendarDays,
  Heart,
  MapPinned,
  MessageCircle,
  Route,
  Wallet,
} from "lucide-react";
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
      className="group flex h-full flex-col overflow-hidden rounded-lg border border-[#d8e5e0] bg-white shadow-[0_18px_46px_-34px_rgb(22_44_40/0.72)] transition hover:-translate-y-0.5 hover:border-[#9fc7ba] hover:shadow-[0_24px_60px_-42px_rgb(22_44_40/0.78)]"
      href={ROUTES.itineraryDetail(item.slug)}
    >
      <div className="relative aspect-[4/5] max-h-[420px] bg-[#e8efe9]">
        {item.coverImageUrl ? (
          <img
            alt={item.title}
            className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.035]"
            decoding="async"
            loading="lazy"
            src={item.coverImageUrl}
          />
        ) : (
          <div className="flex h-full items-center justify-center bg-[linear-gradient(145deg,#e8f4f1,#f8efe0)] text-[#6b7d77]">
            <Route className="size-10" />
          </div>
        )}
        <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(10,23,20,0.1)_30%,rgba(10,23,20,0.78)_100%)]" />
        <div className="absolute left-3 right-3 top-3 flex items-start justify-between gap-3">
          <span className="rounded-full bg-white/94 px-3 py-1 text-xs font-semibold text-[#18342f] shadow-sm backdrop-blur">
            {item.province}
          </span>
          <span className="inline-flex items-center gap-1 rounded-full bg-[#111816]/78 px-2.5 py-1 text-xs font-semibold text-white shadow-sm backdrop-blur">
            <MapPinned className="size-3.5" />
            {item.stopCount}
          </span>
        </div>
        <div className="absolute bottom-3 left-3 right-3 space-y-2 text-white">
          <h2 className="line-clamp-2 text-xl font-semibold leading-6">
            {item.title}
          </h2>
          <div className="flex flex-wrap gap-2 text-xs font-semibold">
            <span className="inline-flex items-center gap-1.5 rounded-full bg-white/16 px-2.5 py-1 backdrop-blur">
              <CalendarDays className="size-3.5" />
              {item.durationDays} ngày
            </span>
            <span className="inline-flex items-center gap-1.5 rounded-full bg-white/16 px-2.5 py-1 backdrop-blur">
              <Wallet className="size-3.5" />
              {formatBudget(item)}
            </span>
          </div>
        </div>
      </div>

      <div className="flex flex-1 flex-col gap-4 p-4">
        <div className="flex items-center justify-between gap-3 border-b border-[#edf3f0] pb-3">
          <div className="flex min-w-0 items-center gap-2.5">
            <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-[linear-gradient(145deg,#18342f,#2f796d)] text-white shadow-sm">
              <Route className="size-4" />
            </span>
            <div className="min-w-0">
              <p className="truncate text-sm font-semibold text-[#142c27]">
                Falzo itinerary
              </p>
              <p className="truncate text-xs font-medium text-[#71817d]">
                {item.transportation || "Route gợi ý"} · {item.province}
              </p>
            </div>
          </div>
          <div className="flex shrink-0 items-center gap-1 text-[#526762]">
            <span className="flex size-8 items-center justify-center rounded-full bg-[#f4f8f6]">
              <Heart className="size-4" />
            </span>
            <span className="flex size-8 items-center justify-center rounded-full bg-[#f4f8f6]">
              <MessageCircle className="size-4" />
            </span>
          </div>
        </div>

        {styles.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {styles.map((style) => (
              <span
                className="rounded-full bg-[#eef7f3] px-2.5 py-1 text-xs font-semibold text-[#315d55]"
                key={style}
              >
                #{style}
              </span>
            ))}
          </div>
        ) : null}

        <div className="mt-auto flex items-center justify-between gap-3 pt-1">
          <span className="text-sm font-semibold text-[#506865]">
            {item.stopCount} điểm dừng có thể copy
          </span>
          <span className="inline-flex items-center gap-1.5 text-sm font-semibold text-[#245d53]">
            Xem
            <ArrowRight className="size-4 transition group-hover:translate-x-0.5" />
          </span>
        </div>
      </div>
    </Link>
  );
}
