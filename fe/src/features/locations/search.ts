import { searchLocationsApi } from "@/features/locations/api";
import type { Location } from "@/features/locations/types";

const defaultLocationSearch = "Ho Chi Minh";
const nominatimSearchUrl = "https://nominatim.openstreetmap.org/search";

type NominatimPlace = {
  address?: {
    city?: string;
    town?: string;
    village?: string;
    municipality?: string;
    state?: string;
    province?: string;
    county?: string;
    country?: string;
  };
  display_name?: string;
  lat?: string;
  lon?: string;
  name?: string;
  osm_id?: number;
  place_id?: number;
};

export function normalizeLocationSearchQuery(query: string) {
  const trimmed = query.trim();
  const normalized = trimmed
    .normalize("NFD")
    .replaceAll(/[\u0300-\u036f]/g, "")
    .replaceAll(/[.\-_]/g, " ")
    .replaceAll(/\s+/g, " ")
    .toLowerCase();

  if (
    normalized === "tphcm" ||
    normalized === "tp hcm" ||
    normalized === "hcm" ||
    normalized === "ho chi minh city" ||
    normalized === "sai gon" ||
    normalized === "saigon"
  ) {
    return defaultLocationSearch;
  }

  return trimmed;
}

function getPlaceName(place: NominatimPlace) {
  return (
    place.address?.city ??
    place.address?.town ??
    place.address?.village ??
    place.address?.municipality ??
    place.address?.state ??
    place.address?.province ??
    place.address?.county ??
    place.name ??
    place.display_name?.split(",")[0]?.trim() ??
    "Location"
  );
}

function toExternalLocation(place: NominatimPlace): Location | null {
  const latitude = Number(place.lat);
  const longitude = Number(place.lon);

  if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) {
    return null;
  }

  const id = place.place_id ?? place.osm_id ?? `${latitude},${longitude}`;

  return {
    id: `geocode:${id}`,
    name: getPlaceName(place),
    address:
      place.display_name ?? `${latitude.toFixed(5)}, ${longitude.toFixed(5)}`,
    latitude,
    longitude,
  };
}

function dedupeLocations(locations: Location[]) {
  const seen = new Set<string>();
  const result: Location[] = [];

  for (const location of locations) {
    const key = [
      location.name.trim().toLowerCase(),
      location.latitude.toFixed(4),
      location.longitude.toFixed(4),
    ].join(":");

    if (seen.has(key)) {
      continue;
    }

    seen.add(key);
    result.push(location);
  }

  return result;
}

async function searchExternalLocationsApi(query: string): Promise<Location[]> {
  const params = new URLSearchParams({
    addressdetails: "1",
    "accept-language": "vi,en",
    format: "jsonv2",
    limit: "8",
    q: query,
  });

  const response = await fetch(`${nominatimSearchUrl}?${params.toString()}`, {
    headers: {
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error("Unable to search external locations.");
  }

  const places = (await response.json()) as NominatimPlace[];
  return dedupeLocations(
    places
      .map(toExternalLocation)
      .filter((location): location is Location => location !== null),
  );
}

export async function searchLocationsWithFallbackApi(query: string) {
  const normalizedQuery = normalizeLocationSearchQuery(query);

  if (!normalizedQuery) {
    return [];
  }

  let backendError: unknown = null;

  try {
    const backendLocations = await searchLocationsApi(normalizedQuery);
    if (backendLocations.length > 0) {
      return backendLocations;
    }
  } catch (error) {
    backendError = error;
  }

  try {
    const externalLocations = await searchExternalLocationsApi(normalizedQuery);
    if (externalLocations.length > 0) {
      return externalLocations;
    }
  } catch {
    if (backendError) {
      throw backendError;
    }
  }

  if (backendError) {
    throw backendError;
  }

  return [];
}

export function isGeocodedLocation(location: Location | null | undefined) {
  return location?.id.startsWith("geocode:") ?? false;
}

export { defaultLocationSearch };
