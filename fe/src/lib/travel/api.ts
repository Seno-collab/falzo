import { mockLocations, mockPosts } from "@/lib/travel/mock-data";
import type {
  CreatePostPayload,
  NearbyLocationsParams,
  PaginatedPosts,
  TravelLocation,
  TravelPost,
} from "@/lib/travel/types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "/api";
const DEFAULT_LIMIT = 12;

function apiUrl(path: string, params?: URLSearchParams) {
  const normalizedBase = API_BASE_URL.replace(/\/$/, "");
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${normalizedBase}${normalizedPath}${params ? `?${params.toString()}` : ""}`;
}

async function requestJson<T>(path: string, params?: URLSearchParams): Promise<T> {
  const response = await fetch(apiUrl(path, params), {
    method: "GET",
    headers: {
      Accept: "application/json",
    },
    cache: "no-store",
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.status}`);
  }

  return (await response.json()) as T;
}

function asRecord(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }

  return value as Record<string, unknown>;
}

function getString(node: Record<string, unknown> | null, keys: string[], fallback = "") {
  if (!node) {
    return fallback;
  }

  for (const key of keys) {
    const value = node[key];
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }

  return fallback;
}

function getNumber(node: Record<string, unknown> | null, keys: string[], fallback = 0) {
  if (!node) {
    return fallback;
  }

  for (const key of keys) {
    const value = node[key];
    if (typeof value === "number") {
      return value;
    }

    if (typeof value === "string" && value.trim()) {
      const numericValue = Number(value);
      if (!Number.isNaN(numericValue)) {
        return numericValue;
      }
    }
  }

  return fallback;
}

function getStringArray(node: Record<string, unknown> | null, keys: string[]) {
  if (!node) {
    return [] as string[];
  }

  for (const key of keys) {
    const value = node[key];
    if (Array.isArray(value)) {
      return value
        .map((item) => (typeof item === "string" ? item.trim() : ""))
        .filter(Boolean);
    }

    if (typeof value === "string" && value.trim()) {
      return value
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean);
    }
  }

  return [] as string[];
}

function getDataArray(payload: unknown): unknown[] {
  if (Array.isArray(payload)) {
    return payload;
  }

  const node = asRecord(payload);
  if (!node) {
    return [];
  }

  if (Array.isArray(node.data)) {
    return node.data;
  }

  if (Array.isArray(node.items)) {
    return node.items;
  }

  if (node.data && typeof node.data === "object") {
    const dataNode = node.data as Record<string, unknown>;
    if (Array.isArray(dataNode.items)) {
      return dataNode.items;
    }
  }

  return [];
}

function normalizeLocation(raw: unknown): TravelLocation {
  const node = asRecord(raw);
  return {
    id: getString(node, ["id", "location_id", "locationId", "slug"], crypto.randomUUID()),
    name: getString(node, ["name", "title", "location_name"], "Unknown location"),
    subtitle: getString(node, ["subtitle", "address", "country", "city"], ""),
    lat: getNumber(node, ["lat", "latitude"], 0),
    lng: getNumber(node, ["lng", "lon", "longitude"], 0),
    imageUrl: getString(node, ["image", "imageUrl", "cover", "thumbnail"], ""),
    postsCount: getNumber(node, ["posts_count", "postsCount", "count"], 0),
    countryCode: getString(node, ["country_code", "countryCode"], ""),
  };
}

function normalizePost(raw: unknown): TravelPost {
  const node = asRecord(raw);
  const locationNode = asRecord(node?.location);

  const locationId =
    getString(node, ["location_id", "locationId"], "") ||
    getString(locationNode, ["id", "location_id", "locationId"], "unknown-location");

  const locationName =
    getString(node, ["location_name", "locationName"], "") ||
    getString(locationNode, ["name", "title"], "Unknown location");

  return {
    id: getString(node, ["id", "post_id", "postId"], crypto.randomUUID()),
    imageUrl:
      getString(node, ["image", "image_url", "imageUrl", "thumbnail", "cover"], "") ||
      "https://picsum.photos/seed/travel-discovery/1000/1200",
    caption: getString(node, ["caption", "description", "title"], ""),
    locationId,
    locationName,
    locationSubtitle:
      getString(node, ["location_subtitle", "locationSubtitle"], "") ||
      getString(locationNode, ["subtitle", "address", "country", "city"], ""),
    tags: getStringArray(node, ["tags", "tag_list", "tagList"]),
    likes: getNumber(node, ["likes", "likes_count", "likesCount"], 0),
    saves: getNumber(node, ["saves", "saved_count", "savedCount"], 0),
    createdAt: getString(node, ["created_at", "createdAt"], ""),
  };
}

function resolvePagination(payload: unknown, requestedPage: number, itemsCount: number) {
  const node = asRecord(payload);
  const explicitNext = getNumber(node, ["next_page", "nextPage"], Number.NaN);

  if (!Number.isNaN(explicitNext) && explicitNext > 0) {
    return explicitNext;
  }

  const hasNext =
    (node && typeof node.hasNext === "boolean" && node.hasNext) ||
    (node && typeof node.has_more === "boolean" && node.has_more);

  if (hasNext) {
    return requestedPage + 1;
  }

  return itemsCount >= DEFAULT_LIMIT ? requestedPage + 1 : null;
}

function paginatePosts(posts: TravelPost[], page: number, limit: number): PaginatedPosts {
  const start = (page - 1) * limit;
  const data = posts.slice(start, start + limit);

  return {
    data,
    nextPage: start + limit < posts.length ? page + 1 : null,
  };
}

export async function getPosts(page = 1, limit = DEFAULT_LIMIT): Promise<PaginatedPosts> {
  const params = new URLSearchParams({
    page: String(page),
    limit: String(limit),
  });

  try {
    const payload = await requestJson<unknown>("/posts", params);
    const items = getDataArray(payload).map(normalizePost);

    return {
      data: items,
      nextPage: resolvePagination(payload, page, items.length),
    };
  } catch {
    return paginatePosts(mockPosts, page, limit);
  }
}

export async function searchLocations(query: string): Promise<TravelLocation[]> {
  const params = new URLSearchParams();
  if (query.trim()) {
    params.set("q", query.trim());
  }

  try {
    const payload = await requestJson<unknown>("/locations/search", params);
    const items = getDataArray(payload).map(normalizeLocation);
    return items;
  } catch {
    return mockLocations.filter((location) => {
      const keyword = query.trim().toLowerCase();
      if (!keyword) {
        return true;
      }

      return `${location.name} ${location.subtitle ?? ""}`.toLowerCase().includes(keyword);
    });
  }
}

export async function getNearbyLocations({
  lat,
  lng,
  radiusKm = 20,
}: NearbyLocationsParams): Promise<TravelLocation[]> {
  const params = new URLSearchParams({
    lat: String(lat),
    lng: String(lng),
    radius_km: String(radiusKm),
  });

  try {
    const payload = await requestJson<unknown>("/locations/nearby", params);
    return getDataArray(payload).map(normalizeLocation);
  } catch {
    return mockLocations;
  }
}

export async function getLocationPosts(locationId: string): Promise<TravelPost[]> {
  try {
    const payload = await requestJson<unknown>(`/locations/${locationId}/posts`);
    return getDataArray(payload).map(normalizePost);
  } catch {
    return mockPosts.filter((post) => post.locationId === locationId);
  }
}

export async function createPost(payload: CreatePostPayload): Promise<void> {
  const formData = new FormData();
  formData.set("caption", payload.caption);
  formData.set("locationId", payload.locationId);
  formData.set("image", payload.imageFile);
  formData.set("tags", payload.tags.join(","));

  const response = await fetch(apiUrl("/posts"), {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    throw new Error(`Create post failed: ${response.status}`);
  }
}

export function getLocationById(locationId: string): TravelLocation | null {
  return mockLocations.find((item) => item.id === locationId) ?? null;
}
