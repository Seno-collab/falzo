"use client";

import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  createPost,
  getLocationPosts,
  getNearbyLocations,
  getPosts,
  searchLocations,
} from "@/lib/travel/api";
import type { CreatePostPayload } from "@/lib/travel/types";

export function useInfinitePosts() {
  return useInfiniteQuery({
    queryKey: ["travel-posts"],
    initialPageParam: 1,
    queryFn: async ({ pageParam }) => getPosts(pageParam),
    getNextPageParam: (lastPage) => lastPage.nextPage ?? undefined,
  });
}

export function useLocationSearch(query: string) {
  return useQuery({
    queryKey: ["travel-location-search", query],
    queryFn: () => searchLocations(query),
    enabled: query.trim().length > 0,
  });
}

export function useNearbyLocations(lat: number, lng: number) {
  return useQuery({
    queryKey: ["travel-nearby-locations", lat, lng],
    queryFn: () => getNearbyLocations({ lat, lng }),
    enabled: Number.isFinite(lat) && Number.isFinite(lng),
  });
}

export function useLocationPosts(locationId: string) {
  return useQuery({
    queryKey: ["travel-location-posts", locationId],
    queryFn: () => getLocationPosts(locationId),
    enabled: Boolean(locationId),
  });
}

export function useCreatePost() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreatePostPayload) => createPost(payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["travel-posts"] });
    },
  });
}
