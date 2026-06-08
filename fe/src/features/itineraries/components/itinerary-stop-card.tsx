import { Clock3, MapPinned, Navigation, Wallet } from "lucide-react";
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
  isLast = false,
  stop,
}: Readonly<{
  isLast?: boolean;
  stop: ItineraryStop;
}>) {
  return (
    <article className="relative pl-9">
      <span className="absolute left-0 top-4 z-1 flex size-8 items-center justify-center rounded-full border-4 border-white bg-[#17332e] text-xs font-bold text-white shadow-sm">
        {stop.stopOrder}
      </span>
      {!isLast ? (
        <span className="absolute left-3.5 top-11 h-[calc(100%_-_1.25rem)] w-px bg-[#d6e5df]" />
      ) : null}
      <div className="rounded-lg border border-[#dbe6e2] bg-white p-4 shadow-[0_14px_34px_-30px_rgb(22_44_40/0.58)] transition hover:border-[#b6d3c9]">
      <div className="flex gap-3">
        <div className="min-w-0 flex-1 space-y-3">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h3 className="text-base font-semibold leading-6 text-[#142c27]">
                {stop.locationName}
              </h3>
              <p className="mt-1 inline-flex items-center gap-1.5 text-xs font-semibold text-[#667b76]">
                <MapPinned className="size-3.5" />
                {stop.latitude.toFixed(4)}, {stop.longitude.toFixed(4)}
              </p>
            </div>
            <span className="inline-flex w-fit items-center gap-1.5 rounded-full bg-[#edf6ff] px-2.5 py-1 text-xs font-semibold text-[#315f91]">
              <Navigation className="size-3.5" />
              Điểm {stop.stopOrder}
            </span>
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
            <p className="rounded-lg border border-[#edf3f0] bg-[#f8fbf9] px-3 py-2 text-sm leading-6 text-[#506865]">
              {stop.note}
            </p>
          ) : null}
        </div>
      </div>
      </div>
    </article>
  );
}
