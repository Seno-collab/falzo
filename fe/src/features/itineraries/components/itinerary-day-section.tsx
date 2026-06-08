import { CalendarDays, MapPinned } from "lucide-react";
import { ItineraryStopCard } from "@/features/itineraries/components/itinerary-stop-card";
import type { ItineraryDay } from "@/features/itineraries/types";

export function ItineraryDaySection({
  day,
}: Readonly<{
  day: ItineraryDay;
}>) {
  return (
    <section className="rounded-lg border border-[#d8e5e0] bg-white/92 p-4 shadow-[0_16px_42px_-36px_rgb(22_44_40/0.62)]">
      <div className="flex items-center justify-between gap-3 border-b border-[#e9f1ee] pb-4">
        <div className="flex items-center gap-3">
          <span className="flex size-10 items-center justify-center rounded-full bg-[#17332e] text-white shadow-sm">
            <CalendarDays className="size-4" />
          </span>
          <div>
            <h2 className="text-xl font-semibold text-[#17332e]">
              Ngày {day.dayNumber}
            </h2>
            <p className="text-sm text-[#667b76]">
              {day.stops.length} điểm dừng đã sắp xếp
            </p>
          </div>
        </div>
        <span className="hidden items-center gap-1.5 rounded-full bg-[#eef7f3] px-3 py-1 text-xs font-semibold text-[#315d55] sm:inline-flex">
          <MapPinned className="size-3.5" />
          Route ngày {day.dayNumber}
        </span>
      </div>
      <div className="relative mt-4 space-y-3 before:absolute before:bottom-6 before:left-4 before:top-6 before:w-px before:bg-[#d6e5df]">
        {day.stops.map((stop, index) => (
          <ItineraryStopCard
            isLast={index === day.stops.length - 1}
            key={stop.id}
            stop={stop}
          />
        ))}
      </div>
    </section>
  );
}
