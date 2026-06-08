import { Suspense } from "react";
import { ItineraryListScreen } from "@/features/itineraries/screens/itinerary-list-screen";

export default function ItinerariesPage() {
  return (
    <Suspense fallback={null}>
      <ItineraryListScreen />
    </Suspense>
  );
}
