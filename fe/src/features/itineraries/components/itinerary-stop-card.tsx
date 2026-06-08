import { Clock3, MapPinned, Wallet } from "lucide-react";
import type { ItineraryStop } from "@/features/itineraries/types";

function formatMoney(value: number) {
  if (value === 0) {
    return "0đ";
  }

  return `${new Intl.NumberFormat("vi-VN").format(value)}đ`;
}

function formatTimeRange(stop: ItineraryStop) {
  if (stop.startTime && stop.endTime) {
    return `${stop.startTime} - ${stop.endTime}`;
  }
  if (stop.startTime) {
    return stop.startTime;
  }
  return "Linh hoạt";
}

export function ItineraryStopCard({
  stop,
}: Readonly<{
  stop: ItineraryStop;
}>) {
  return (
    <article className="rounded-lg border border-[#dbe6e2] bg-white p-4 shadow-[0_14px_34px_-30px_rgb(22_44_40/0.58)]">
      <div className="flex gap-3">
        <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-[#17332e] text-sm font-bold text-white">
          {stop.stopOrder}
        </span>
        <div className="min-w-0 flex-1 space-y-3">
          <div>
            <h3 className="text-base font-semibold leading-6 text-[#142c27]">
              {stop.locationName}
            </h3>
            <p className="mt-1 inline-flex items-center gap-1.5 text-xs font-semibold text-[#667b76]">
              <MapPinned className="size-3.5" />
              {stop.latitude.toFixed(4)}, {stop.longitude.toFixed(4)}
            </p>
          </div>

          <div className="grid gap-2 text-xs font-semibold sm:grid-cols-2">
            <span className="inline-flex items-center gap-1.5 rounded-lg bg-[#f2f7f5] px-2.5 py-2 text-[#46645f]">
              <Clock3 className="size-3.5" />
              {formatTimeRange(stop)}
            </span>
            <span className="inline-flex items-center gap-1.5 rounded-lg bg-[#fff6e8] px-2.5 py-2 text-[#7a5a18]">
              <Wallet className="size-3.5" />
              {formatMoney(stop.estimatedCost)}
            </span>
          </div>

          {stop.note ? (
            <p className="rounded-lg bg-[#f8fbf9] px-3 py-2 text-sm leading-6 text-[#506865]">
              {stop.note}
            </p>
          ) : null}
        </div>
      </div>
    </article>
  );
}
