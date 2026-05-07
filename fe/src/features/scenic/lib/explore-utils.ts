import type { Category } from "@/features/categories/types";

export const ALL_COLLECTION = "All";
export const COMMUNITY_COLLECTION = "Community";

export function showsCommunityFeed(collection: string) {
  return collection === ALL_COLLECTION || collection === COMMUNITY_COLLECTION;
}

export function getExploreCollections(categories: Category[] | undefined) {
  const names = [
    ...(categories ?? []).map((category) => category.name),
  ]
    .map((name) => name.trim())
    .filter(
      (name) =>
        name.length > 0 &&
        name !== ALL_COLLECTION &&
        name !== COMMUNITY_COLLECTION,
    );

  return [ALL_COLLECTION, COMMUNITY_COLLECTION, ...Array.from(new Set(names))];
}

export function toggleSetValue<T>(current: Set<T>, value: T) {
  const next = new Set(current);

  if (next.has(value)) {
    next.delete(value);
  } else {
    next.add(value);
  }

  return next;
}
