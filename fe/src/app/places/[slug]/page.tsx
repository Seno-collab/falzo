import { PlaceDetailScreen } from "@/features/locations/screens/place-detail-screen";

export function generateStaticParams() {
  return [{ slug: "hon-yen-phu-yen" }];
}

export default function PlacePage() {
  return <PlaceDetailScreen />;
}
