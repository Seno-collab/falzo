import { apiGet, endpointPath, envEndpoint } from "@/lib/api-utils";
import type { Location, LocationPost, NearbyLocation } from "./types";

const LOCATIONS_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_LOCATIONS_ENDPOINT,
  process.env.VITE_LOCATIONS_ENDPOINT,
  "/locations",
);

export const searchLocationsApi = (query: string): Promise<Location[]> =>
  apiGet<Location[]>(
    endpointPath(LOCATIONS_ENDPOINT, "search"),
    {
      params: { q: query },
    },
  );

export const getNearbyLocationsApi = (params: {
  latitude: number;
  longitude: number;
  radiusMeters: number;
}): Promise<NearbyLocation[]> =>
  apiGet<NearbyLocation[]>(endpointPath(LOCATIONS_ENDPOINT, "nearby"), {
    params: {
      lat: params.latitude,
      lng: params.longitude,
      radius: params.radiusMeters,
    },
  });

export const getLocationPostsApi = (
  locationId: string,
): Promise<LocationPost[]> =>
  apiGet<LocationPost[]>(
    endpointPath(LOCATIONS_ENDPOINT, locationId, "posts"),
  );
