import { searchLocationsApi } from "@/features/locations/api";
import type { Location, LocationPost } from "@/features/locations/types";
import { getPostsApi } from "@/features/posts/api";
import type { Post } from "@/features/posts/types";

const defaultLocationSearch = "Ho Chi Minh";
const nominatimSearchUrl = "https://nominatim.openstreetmap.org/search";
const postBackedLocationIdPrefix = "post-location:";
const assignedLocationCityRadiusMeters = 50_000;
const selectedLocationPostRadiusMeters = 500;

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

function trimString(value: unknown) {
  return typeof value === "string" ? value.trim() : "";
}

function normalizeComparableText(value: unknown) {
  return trimString(value)
    .normalize("NFD")
    .replaceAll(/[\u0300-\u036f]/g, "")
    .replaceAll(/[.\-_]/g, " ")
    .replaceAll(/\s+/g, " ")
    .toLowerCase();
}

export function normalizeLocationSearchQuery(query: unknown) {
  const trimmed = trimString(query);
  const normalized = normalizeComparableText(trimmed);

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
    const name = trimString(location.name);
    if (!name) {
      continue;
    }

    const key = [
      name.toLowerCase(),
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

function getDistanceMeters(
  origin: Pick<Location, "latitude" | "longitude">,
  target: Pick<Location, "latitude" | "longitude">,
) {
  const earthRadiusMeters = 6_371_000;
  const toRadians = (value: number) => (value * Math.PI) / 180;
  const deltaLatitude = toRadians(target.latitude - origin.latitude);
  const deltaLongitude = toRadians(target.longitude - origin.longitude);
  const originLatitude = toRadians(origin.latitude);
  const targetLatitude = toRadians(target.latitude);
  const haversine =
    Math.sin(deltaLatitude / 2) ** 2 +
    Math.cos(originLatitude) *
      Math.cos(targetLatitude) *
      Math.sin(deltaLongitude / 2) ** 2;

  return 2 * earthRadiusMeters * Math.asin(Math.sqrt(haversine));
}

function postToLocationPost(post: Post): LocationPost {
  return {
    id: String(post.id),
    user_id: post.user_id,
    image_url: post.image_url,
    caption: post.caption,
    location_name: post.location_name,
    latitude: post.latitude,
    longitude: post.longitude,
  };
}

function postHasCoordinates(post: Post) {
  return Number.isFinite(post.latitude) && Number.isFinite(post.longitude);
}

function postsToAssignedLocations(posts: Post[]) {
  const locationsByKey = new Map<
    string,
    Location & { post_ids: number[]; post_count: number }
  >();

  for (const post of posts) {
    const name = trimString(post.location_name);
    if (!name || !postHasCoordinates(post)) {
      continue;
    }

    const key = [
      normalizeComparableText(name),
      post.latitude.toFixed(4),
      post.longitude.toFixed(4),
    ].join(":");
    const existing = locationsByKey.get(key);

    if (existing) {
      existing.post_ids.push(post.id);
      existing.post_count += 1;
      existing.address = `Previously assigned from ${existing.post_count} posts`;
      continue;
    }

    locationsByKey.set(key, {
      id: `${postBackedLocationIdPrefix}${encodeURIComponent(key)}`,
      name,
      address: "Previously assigned from 1 post",
      latitude: post.latitude,
      longitude: post.longitude,
      post_ids: [post.id],
      post_count: 1,
    });
  }

  return Array.from(locationsByKey.values());
}

async function searchAssignedLocationsApi(
  query: string,
  center?: Pick<Location, "latitude" | "longitude">,
) {
  const requests = [
    getPostsApi({
      limit: 50,
      search: query,
    }),
  ];

  if (center) {
    requests.push(
      getPostsApi({
        latitude: center.latitude,
        limit: 50,
        longitude: center.longitude,
        radiusMeters: assignedLocationCityRadiusMeters,
        sort: "nearby",
      }),
    );
  }

  const settled = await Promise.allSettled(requests);
  const posts = settled.flatMap((result) =>
    result.status === "fulfilled" ? result.value : [],
  );

  return postsToAssignedLocations(posts);
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
  let backendLocations: Location[] = [];

  try {
    backendLocations = await searchLocationsApi(normalizedQuery);
  } catch (error) {
    backendError = error;
  }

  try {
    const assignedLocations = await searchAssignedLocationsApi(
      normalizedQuery,
      backendLocations[0],
    );
    const localMatches = dedupeLocations([
      ...backendLocations,
      ...assignedLocations,
    ]);

    if (localMatches.length > 0) {
      return localMatches;
    }
  } catch {
    // Search should still fall back to geocoding if post search is unavailable.
  }

  try {
    const externalLocations = await searchExternalLocationsApi(normalizedQuery);
    const assignedLocations = await searchAssignedLocationsApi(
      normalizedQuery,
      externalLocations[0],
    ).catch(() => []);
    const matches = dedupeLocations([
      ...backendLocations,
      ...assignedLocations,
      ...externalLocations,
    ]);

    if (matches.length > 0) {
      return matches;
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

export function isPostBackedLocation(location: Location | null | undefined) {
  return location?.id.startsWith(postBackedLocationIdPrefix) ?? false;
}

export async function getPostBackedLocationPostsApi(location: Location) {
  const selectedPostIds = new Set(location.post_ids ?? []);
  const settled = await Promise.allSettled([
    getPostsApi({
      limit: 50,
      search: location.name,
    }),
    getPostsApi({
      latitude: location.latitude,
      limit: 50,
      longitude: location.longitude,
      radiusMeters: selectedLocationPostRadiusMeters,
      sort: "nearby",
    }),
  ]);
  const postsById = new Map<number, Post>();

  for (const result of settled) {
    if (result.status !== "fulfilled") {
      continue;
    }

    for (const post of result.value) {
      postsById.set(post.id, post);
    }
  }

  return Array.from(postsById.values())
    .filter((post) => {
      if (selectedPostIds.size > 0) {
        return selectedPostIds.has(post.id);
      }

      return (
        normalizeComparableText(post.location_name) ===
          normalizeComparableText(location.name) &&
        postHasCoordinates(post) &&
        getDistanceMeters(location, post) <= selectedLocationPostRadiusMeters
      );
    })
    .map(postToLocationPost);
}

export { defaultLocationSearch };
