import { http } from "@/lib/http";
import type { ApiEnvelope } from "@/types/api/response";
import type { Location, LocationPost, NearbyLocation } from "./types";

const LOCATIONS_ENDPOINT =
  process.env.NEXT_PUBLIC_LOCATIONS_ENDPOINT ??
  process.env.VITE_LOCATIONS_ENDPOINT ??
  "/locations";

function unwrapResponseData<T>(value: ApiEnvelope<T> | T): T {
  if (
    value &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    "success" in value &&
    "data" in value
  ) {
    return (value as ApiEnvelope<T>).data;
  }

  return value as T;
}

export async function searchLocationsApi(query: string): Promise<Location[]> {
  const response = await http.get<ApiEnvelope<Location[]> | Location[]>(
    `${LOCATIONS_ENDPOINT}/search`,
    {
      params: { q: query },
    },
  );

  return unwrapResponseData(response.data);
}

export async function getNearbyLocationsApi(params: {
  latitude: number;
  longitude: number;
  radiusMeters: number;
}): Promise<NearbyLocation[]> {
  const response = await http.get<
    ApiEnvelope<NearbyLocation[]> | NearbyLocation[]
  >(`${LOCATIONS_ENDPOINT}/nearby`, {
    params: {
      lat: params.latitude,
      lng: params.longitude,
      radius: params.radiusMeters,
    },
  });

  return unwrapResponseData(response.data);
}

export async function getLocationPostsApi(
  locationId: string,
): Promise<LocationPost[]> {
  const response = await http.get<ApiEnvelope<LocationPost[]> | LocationPost[]>(
    `${LOCATIONS_ENDPOINT}/${locationId}/posts`,
  );

  return unwrapResponseData(response.data);
}
