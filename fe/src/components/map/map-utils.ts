import type { LatLngTuple } from "leaflet";
import type { Coordinates, MapPoint } from "./types";

export const defaultMapCenter: LatLngTuple = [10.7769, 106.7009];

export function coordinatesToLatLng(coordinates: Coordinates): LatLngTuple {
  return [coordinates.latitude, coordinates.longitude];
}

export function formatDistance(meters?: number) {
  if (meters === undefined) {
    return null;
  }

  return meters >= 1000
    ? `${(meters / 1000).toFixed(1)} km away`
    : `${Math.round(meters)} m away`;
}

export function getInitialCenter({
  currentPosition,
  points,
  selectedPoint,
}: {
  currentPosition?: Coordinates | null;
  points: MapPoint[];
  selectedPoint?: MapPoint;
}) {
  if (selectedPoint) {
    return coordinatesToLatLng(selectedPoint);
  }

  if (currentPosition) {
    return coordinatesToLatLng(currentPosition);
  }

  if (points[0]) {
    return coordinatesToLatLng(points[0]);
  }

  return defaultMapCenter;
}
