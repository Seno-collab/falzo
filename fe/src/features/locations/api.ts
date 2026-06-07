import type { AxiosRequestConfig } from "axios";
import { LOCATIONS_ENDPOINT, PLACES_ENDPOINT } from "@/lib/api-config";
import { apiGet, endpointPath } from "@/lib/api-utils";
import type { Location, LocationPost, NearbyLocation, PlaceDetail } from "./types";

export const searchLocationsApi = (
  query: string,
  config?: AxiosRequestConfig,
): Promise<Location[]> =>
  apiGet<Location[]>(
    endpointPath(LOCATIONS_ENDPOINT, "search"),
    {
      ...config,
      params: { q: query },
    },
  );

export const getNearbyLocationsApi = (params: {
  latitude: number;
  longitude: number;
  radiusMeters: number;
  signal?: AbortSignal;
}): Promise<NearbyLocation[]> =>
  apiGet<NearbyLocation[]>(endpointPath(LOCATIONS_ENDPOINT, "nearby"), {
    params: {
      lat: params.latitude,
      lng: params.longitude,
      radius: params.radiusMeters,
    },
    signal: params.signal,
  });

export const getLocationPostsApi = (
  locationId: string,
  config?: AxiosRequestConfig,
): Promise<LocationPost[]> =>
  apiGet<LocationPost[]>(
    endpointPath(LOCATIONS_ENDPOINT, locationId, "posts"),
    config,
  );

export const getPlaceBySlugApi = (
  slug: string,
  config?: AxiosRequestConfig,
): Promise<PlaceDetail> =>
  apiGet<PlaceDetail>(endpointPath(PLACES_ENDPOINT, slug), config);
