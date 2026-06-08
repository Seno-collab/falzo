import { ItineraryDetailScreen } from "@/features/itineraries/screens/itinerary-detail-screen";

export function generateStaticParams() {
  return [{ slug: "phu-yen-2-ngay-1-dem-duoi-1tr5" }];
}

export default function ItineraryDetailPage() {
  return <ItineraryDetailScreen />;
}
