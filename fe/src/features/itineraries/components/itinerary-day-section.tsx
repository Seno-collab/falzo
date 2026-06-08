import { CalendarDays } from "lucide-react";
import { ItineraryStopCard } from "@/features/itineraries/components/itinerary-stop-card";
import type { ItineraryDay } from "@/features/itineraries/types";

export function ItineraryDaySection({
  day,
}: Readonly<{
  day: ItineraryDay;
}>) {
  return (
    <section className="space-y-3">
      <div className="flex items-center gap-2">
        <span className="flex size-9 items-center justify-center rounded-full bg-[#e8f6ef] text-[#285f4d]">
          <CalendarDays className="size-4" />
        </span>
        <div>
          <h2 className="text-xl font-semibold text-[#17332e]">
            Ngày {day.dayNumber}
          </h2>
          <p className="text-sm text-[#667b76]">
            {day.stops.length} điểm dừng
          </p>
        </div>
      </div>
      <div className="space-y-3">
        {day.stops.map((stop) => (
          <ItineraryStopCard key={stop.id} stop={stop} />
        ))}
      </div>
    </section>
  );
}
