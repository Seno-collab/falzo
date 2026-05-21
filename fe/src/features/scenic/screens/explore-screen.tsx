"use client";

import {
  Bookmark,
  Camera,
  Clock3,
  Flame,
  LocateFixed,
  MapIcon,
  Menu,
  Plus,
  Search,
  SlidersHorizontal,
  Sparkles,
  Tags,
  UserRound,
  X,
} from "lucide-react";
import {
  useInfiniteQuery,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import type { InfiniteData } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  useCallback,
  useDeferredValue,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { toast } from "sonner";
import MapClient from "@/components/map";
import type { Coordinates, MapPoint } from "@/components/map";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  getApiErrorMessage,
  getMeApi,
  hasAuthSession,
} from "@/features/auth/api";
import type { AuthUser } from "@/features/auth/types";
import { getAuthUserDisplayName } from "@/features/auth/user-display";
import { getCategoriesApi } from "@/features/categories/api";
import type { Category } from "@/features/categories/types";
import {
  normalizeLocationSearchQuery,
  searchLocationsWithFallbackApi,
} from "@/features/locations/search";
import { NotificationBell } from "@/features/notifications/notification-bell";
import {
  cleanNotificationIds,
  getNotificationsApi,
  markNotificationsReadApi,
  subscribeNotificationEvents,
} from "@/features/notifications/api";
import type { AppNotification } from "@/features/notifications/types";
import {
  createPostCommentApi,
  deletePostApi,
  getPostCommentEventsUrl,
  getPostCommentsApi,
  getPostDetailApi,
  getPostEventsUrl,
  getPostsPageApi,
  likePostApi,
  parsePostCommentCreatedEvent,
  parsePostCreatedEvent,
  parsePostDeletedEvent,
  parseUserAvatarUpdatedEvent,
  reportPostApi,
  savePostApi,
  unlikePostApi,
  unsavePostApi,
  updatePostCommentApi,
} from "@/features/posts/api";
import type { Post, PostComment, PostsPage } from "@/features/posts/types";
import type { PostSort } from "@/features/posts/types";
import { ExplorePostCard } from "@/features/scenic/components/explore-cards";
import { PostDetailDialog } from "@/features/scenic/components/explore-detail-dialogs";
import {
  ALL_COLLECTION,
  FOLLOWING_COLLECTION,
  getExploreCollections,
  showsCommunityFeed,
  toggleSetValue,
} from "@/features/scenic/lib/explore-utils";
import { ROUTES } from "@/lib/routes";
import { cn } from "@/lib/utils";

type PostsInfiniteData = InfiniteData<PostsPage, string | null>;

const postsPageSize = 24;
const maxNotifications = 30;
const maxNearbyRadiusMeters = 1_000_000;
const nearbyRadiusOptions = [
  5_000,
  10_000,
  25_000,
  50_000,
  100_000,
  300_000,
  500_000,
  1_000_000,
] as const;

const firstLoadTravelSuggestions = [
  {
    query: "weekend trip",
    sort: "popular" as const,
    title: "Weekend escape",
  },
  {
    query: "scenic view",
    sort: "trending" as const,
    title: "Scenic views",
  },
  {
    query: "local culture",
    sort: "newest" as const,
    title: "Local culture",
  },
] as const;

const exploreHeroHeadlineWords =
  "Discover beautiful destinations and places worth visiting.".split(" ");

function clampNearbyRadiusMeters(radiusMeters: number) {
  if (!Number.isFinite(radiusMeters) || radiusMeters <= 0) {
    return 25_000;
  }

  return Math.min(Math.round(radiusMeters), maxNearbyRadiusMeters);
}

function formatRadiusLabel(radiusMeters: number) {
  return radiusMeters >= 1000
    ? `${Math.round(radiusMeters / 1000)} km`
    : `${radiusMeters} m`;
}

function getDistanceMeters(
  origin: Coordinates | null,
  target: Coordinates,
) {
  if (!origin) {
    return undefined;
  }

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

function getCommentTime(comment: PostComment) {
  const timestamp = new Date(comment.created_at).getTime();
  return Number.isFinite(timestamp) ? timestamp : 0;
}

function sortPostComments(comments: PostComment[]) {
  return [...comments].sort(
    (left, right) =>
      getCommentTime(left) - getCommentTime(right) || left.id - right.id,
  );
}

function upsertPostComment(
  commentsByPost: Record<number, PostComment[]>,
  postId: number,
  comment: PostComment,
) {
  const comments = commentsByPost[postId] ?? [];
  const existingIndex = comments.findIndex((item) => item.id === comment.id);
  const nextComments =
    existingIndex >= 0
      ? comments.map((item) => (item.id === comment.id ? comment : item))
      : [...comments, comment];

  return {
    ...commentsByPost,
    [postId]: sortPostComments(nextComments),
  };
}

function mergePostComments(
  commentsByPost: Record<number, PostComment[]>,
  postId: number,
  comments: PostComment[],
) {
  const commentsById = new Map<number, PostComment>();
  for (const comment of commentsByPost[postId] ?? []) {
    commentsById.set(comment.id, comment);
  }
  for (const comment of comments) {
    commentsById.set(comment.id, comment);
  }

  return {
    ...commentsByPost,
    [postId]: sortPostComments(Array.from(commentsById.values())),
  };
}

type PostActionOverrides = Record<number, boolean>;

function isPostLiked(post: Post | null, likedPosts: PostActionOverrides) {
  return Boolean(post && (likedPosts[post.id] ?? post.is_liked));
}

function isPostSaved(post: Post | null, savedPosts: PostActionOverrides) {
  return Boolean(post && (savedPosts[post.id] ?? post.is_saved));
}

function readAuthUserId(user: AuthUser | null | undefined) {
  const rawId = user?.id ?? user?.user_id ?? user?.userId ?? user?.subject;
  const id = Number(rawId);
  return Number.isFinite(id) && id > 0 ? id : null;
}

function normalizeSearchValue(value: string) {
  return value
    .normalize("NFD")
    .replaceAll(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .trim();
}

function matchesPostSearch(post: Post, searchValue: string) {
  const terms = normalizeSearchValue(searchValue).split(/\s+/).filter(Boolean);
  if (terms.length === 0) {
    return true;
  }

  const searchableText = normalizeSearchValue(
    [
      post.caption,
      post.location_name,
      post.user_name,
      post.category_name,
      post.category_slug,
    ].join(" "),
  );

  return terms.every((term) => searchableText.includes(term));
}

function mergeNotification(
  notifications: AppNotification[],
  notification: AppNotification,
) {
  const existingIndex = notifications.findIndex(
    (item) => item.id === notification.id,
  );
  const next =
    existingIndex >= 0
      ? notifications.map((item) =>
          item.id === notification.id ? notification : item,
        )
      : [notification, ...notifications];

  return next
    .sort(
      (left, right) =>
        new Date(right.created_at).getTime() -
        new Date(left.created_at).getTime(),
    )
    .slice(0, maxNotifications);
}

function getNotificationPostId(notification: AppNotification) {
  const id = Number(notification.post_id);
  return Number.isFinite(id) && id > 0 ? id : null;
}

export function ExploreScreen() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const loadMoreRef = useRef<HTMLDivElement | null>(null);
  const commentInputRefs = useRef<Record<number, HTMLInputElement | null>>({});

  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [activeCollection, setActiveCollection] = useState("All");
  const [feedSort, setFeedSort] = useState<PostSort>("newest");
  const [nearbyCoords, setNearbyCoords] = useState<{
    latitude: number;
    longitude: number;
  } | null>(null);
  const [nearbyPlaceName, setNearbyPlaceName] = useState<string | null>(null);
  const [nearbyRadiusMeters, setNearbyRadiusMeters] = useState(50_000);
  const [showMapPanel, setShowMapPanel] = useState(false);
  const [selectedMapPostId, setSelectedMapPostId] = useState<number | null>(
    null,
  );
  const [selectedMapClusterId, setSelectedMapClusterId] = useState<
    string | null
  >(null);
  const [searchValue, setSearchValue] = useState("");
  const [showFirstLoadSuggestions, setShowFirstLoadSuggestions] =
    useState(true);
  const [showSavedBoard, setShowSavedBoard] = useState(false);
  const [likedPosts, setLikedPosts] = useState<PostActionOverrides>({});
  const [savedPosts, setSavedPosts] = useState<PostActionOverrides>({});
  const [openComments, setOpenComments] = useState<Set<number>>(new Set());
  const [loadingComments, setLoadingComments] = useState<Set<number>>(
    new Set(),
  );
  const [selectedPostId, setSelectedPostId] = useState<number | null>(null);
  const [selectedChatPostId, setSelectedChatPostId] = useState<number | null>(
    null,
  );
  const [commentsByPost, setCommentsByPost] = useState<
    Record<number, PostComment[]>
  >({});
  const [commentInputs, setCommentInputs] = useState<Record<number, string>>(
    {},
  );
  const [replyTargets, setReplyTargets] = useState<
    Record<number, PostComment | null>
  >({});
  const [editingComments, setEditingComments] = useState<
    Record<number, PostComment | null>
  >({});
  const [notifications, setNotifications] = useState<AppNotification[]>([]);
  const [readNotificationIds, setReadNotificationIds] = useState<Set<string>>(
    () => new Set(),
  );
  const deferredSearchValue = useDeferredValue(searchValue);
  const activeSearch = deferredSearchValue.trim();
  const normalizedActiveSearch = normalizeSearchValue(activeSearch);

  const categoriesQuery = useQuery({
    queryKey: ["categories"],
    queryFn: getCategoriesApi,
  });
  const activeCategorySlug = useMemo(() => {
    if (showsCommunityFeed(activeCollection)) {
      return "";
    }

    return (
      categoriesQuery.data?.find(
        (category) => category.name === activeCollection,
      )?.slug ?? ""
    );
  }, [activeCollection, categoriesQuery.data]);
  const activeFeed =
    activeCollection === FOLLOWING_COLLECTION ? "following" : undefined;

  const postsQuery = useInfiniteQuery({
    queryKey: [
      "posts",
      "explore",
      activeSearch,
      activeCategorySlug,
      activeFeed,
      feedSort,
      nearbyCoords,
      nearbyRadiusMeters,
    ],
    queryFn: ({ pageParam }) =>
      getPostsPageApi({
        cursor: pageParam,
        limit: postsPageSize,
        search: activeSearch,
        categorySlug: activeCategorySlug,
        feed: activeFeed,
        sort: feedSort,
        latitude: nearbyCoords?.latitude,
        longitude: nearbyCoords?.longitude,
        radiusMeters: feedSort === "nearby" ? nearbyRadiusMeters : undefined,
      }),
    getNextPageParam: (lastPage) =>
      lastPage.has_more && lastPage.next_cursor
        ? lastPage.next_cursor
        : undefined,
    initialPageParam: null as string | null,
  });

  const postDetailQuery = useQuery({
    enabled: selectedPostId !== null,
    queryKey: ["posts", "detail", selectedPostId],
    queryFn: () => getPostDetailApi(selectedPostId ?? 0),
  });

  const profileQuery = useQuery({
    enabled: isAuthenticated,
    queryKey: ["auth", "me", "explore"],
    queryFn: () => getMeApi<AuthUser>(),
    refetchOnMount: "always",
    retry: false,
    staleTime: 0,
  });
  const notificationsQuery = useQuery({
    enabled: isAuthenticated,
    queryKey: ["notifications", "list"],
    queryFn: () => getNotificationsApi(maxNotifications),
    refetchOnMount: "always",
    retry: false,
    staleTime: 0,
  });
  const profileName =
    profileQuery.data && !profileQuery.isFetching
      ? getAuthUserDisplayName(profileQuery.data, "") || null
      : null;
  const currentUserId = readAuthUserId(profileQuery.data);

  const addNotification = useCallback((notification: AppNotification) => {
    setNotifications((current) => mergeNotification(current, notification));
  }, []);

  const markNotificationIdsRead = useCallback((ids: string[]) => {
    const cleanIds = cleanNotificationIds(ids);
    if (cleanIds.length === 0) {
      return;
    }

    const readAt = new Date().toISOString();
    setReadNotificationIds((current) => {
      const next = new Set(current);
      cleanIds.forEach((id) => next.add(id));
      return next;
    });
    setNotifications((current) =>
      current.map((notification) =>
        cleanIds.includes(notification.id)
          ? { ...notification, read_at: notification.read_at ?? readAt }
          : notification,
      ),
    );
    void markNotificationsReadApi(cleanIds).catch(() => undefined);
  }, []);

  const markNotificationsRead = useCallback(() => {
    markNotificationIdsRead(
      notifications
        .filter(
          (notification) =>
            !notification.read_at && !readNotificationIds.has(notification.id),
        )
        .map((notification) => notification.id),
    );
  }, [markNotificationIdsRead, notifications, readNotificationIds]);

  const unreadNotificationCount = useMemo(
    () =>
      notifications.reduce(
        (count, notification) =>
          readNotificationIds.has(notification.id) ? count : count + 1,
        0,
      ),
    [notifications, readNotificationIds],
  );

  useEffect(() => {
    document.title = "Falzo Travel | Destination Discovery";
    setIsAuthenticated(hasAuthSession());
  }, []);

  useEffect(() => {
    if (!isAuthenticated) {
      return;
    }

    return subscribeNotificationEvents({
      onNotification: addNotification,
      onError: () => undefined,
    });
  }, [addNotification, isAuthenticated]);

  useEffect(() => {
    if (!notificationsQuery.data) {
      return;
    }

    setNotifications((current) =>
      notificationsQuery.data.reduce(
        (merged, notification) => mergeNotification(merged, notification),
        current,
      ),
    );
    setReadNotificationIds((current) => {
      const next = new Set(current);
      notificationsQuery.data
        .filter((notification) => Boolean(notification.read_at))
        .forEach((notification) => next.add(notification.id));
      return next;
    });
  }, [notificationsQuery.data]);

  useEffect(() => {
    const source = new EventSource(getPostEventsUrl());
    const handlePostCreated = (event: Event) => {
      const post = parsePostCreatedEvent(event as MessageEvent<string>);
      if (!post) {
        return;
      }
      if (activeSearch && !matchesPostSearch(post, activeSearch)) {
        return;
      }
      if (activeCategorySlug && post.category_slug !== activeCategorySlug) {
        return;
      }
      if (activeFeed) {
        return;
      }

      queryClient.setQueryData<PostsInfiniteData>(
        [
          "posts",
          "explore",
          activeSearch,
          activeCategorySlug,
          activeFeed,
          feedSort,
          nearbyCoords,
          nearbyRadiusMeters,
        ],
        (current) => {
          if (!current) {
            return {
              pageParams: [null],
              pages: [{ items: [post], has_more: false }],
            };
          }

          if (
            current.pages.some((page) =>
              page.items.some((item) => item.id === post.id),
            )
          ) {
            return current;
          }

          const [firstPage, ...restPages] = current.pages;
          if (!firstPage) {
            return {
              ...current,
              pages: [{ items: [post], has_more: false }],
            };
          }

          return {
            ...current,
            pages: [
              { ...firstPage, items: [post, ...firstPage.items] },
              ...restPages,
            ],
          };
        },
      );
    };
    const handlePostDeleted = (event: Event) => {
      const deleted = parsePostDeletedEvent(event as MessageEvent<string>);
      if (!deleted) {
        return;
      }

      queryClient.setQueriesData<PostsInfiniteData>(
        { queryKey: ["posts", "explore"] },
        (current) => {
          if (!current) {
            return current;
          }

          return {
            ...current,
            pages: current.pages.map((page) => ({
              ...page,
              items: page.items.filter((item) => item.id !== deleted.id),
            })),
          };
        },
      );
      queryClient.removeQueries({ queryKey: ["posts", "detail", deleted.id] });
      void queryClient.invalidateQueries({ queryKey: ["posts"] });
      setSelectedPostId((current) =>
        current === deleted.id ? null : current,
      );
      setSelectedChatPostId((current) =>
        current === deleted.id ? null : current,
      );
      setCommentsByPost((current) => {
        if (!(deleted.id in current)) {
          return current;
        }
        const next = { ...current };
        delete next[deleted.id];
        return next;
      });
    };
    const handleUserAvatarUpdated = (event: Event) => {
      const updated = parseUserAvatarUpdatedEvent(
        event as MessageEvent<string>,
      );
      if (!updated) {
        return;
      }

      const applyAvatar = (post: Post): Post =>
        post.user_id === updated.user_id
          ? {
              ...post,
              avatar_url: updated.avatar_url,
              user_avatar_url: updated.avatar_url,
            }
          : post;

      queryClient.setQueriesData<PostsInfiniteData>(
        { queryKey: ["posts", "explore"] },
        (current) => {
          if (!current) {
            return current;
          }

          return {
            ...current,
            pages: current.pages.map((page) => ({
              ...page,
              items: page.items.map(applyAvatar),
            })),
          };
        },
      );
      queryClient.setQueriesData<Post>(
        { queryKey: ["posts", "detail"] },
        (current) => (current ? applyAvatar(current) : current),
      );
    };

    source.addEventListener("post.created", handlePostCreated);
    source.addEventListener("post.deleted", handlePostDeleted);
    source.addEventListener("user.avatar_updated", handleUserAvatarUpdated);
    return () => {
      source.removeEventListener("post.created", handlePostCreated);
      source.removeEventListener("post.deleted", handlePostDeleted);
      source.removeEventListener("user.avatar_updated", handleUserAvatarUpdated);
      source.close();
    };
  }, [
    activeCategorySlug,
    activeFeed,
    activeSearch,
    feedSort,
    nearbyCoords,
    nearbyRadiusMeters,
    queryClient,
  ]);

  const likeMutation = useMutation({
    mutationFn: async ({
      isLiked,
      postId,
    }: {
      isLiked: boolean;
      postId: number;
    }) => {
      if (isLiked) {
        await unlikePostApi(postId);
        return false;
      }

      await likePostApi(postId);
      return true;
    },
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: (nextIsLiked, variables) => {
      setLikedPosts((current) => ({
        ...current,
        [variables.postId]: nextIsLiked,
      }));
    },
  });

  const saveMutation = useMutation({
    mutationFn: async ({
      isSaved,
      postId,
    }: {
      isSaved: boolean;
      postId: number;
    }) => {
      if (isSaved) {
        await unsavePostApi(postId);
        return false;
      }

      await savePostApi(postId);
      return true;
    },
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: (nextIsSaved, variables) => {
      setSavedPosts((current) => ({
        ...current,
        [variables.postId]: nextIsSaved,
      }));
    },
  });

  const deletePostMutation = useMutation({
    mutationFn: deletePostApi,
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["posts"] });
      setSelectedPostId(null);
      toast.success("Travel post deleted.");
    },
  });

  const reportPostMutation = useMutation({
    mutationFn: ({ postId, reason }: { postId: number; reason: string }) =>
      reportPostApi(postId, { reason }),
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: () => toast.success("Report sent."),
  });

  const commentMutation = useMutation({
    mutationFn: createPostCommentApi,
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: (comment, variables) => {
      setCommentsByPost((current) =>
        upsertPostComment(current, variables.postId, comment),
      );
      setCommentInputs((current) => ({ ...current, [variables.postId]: "" }));
      setReplyTargets((current) => ({ ...current, [variables.postId]: null }));
      toast.success("Comment posted.");
    },
  });

  const updateCommentMutation = useMutation({
    mutationFn: updatePostCommentApi,
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: (comment, variables) => {
      setCommentsByPost((current) =>
        upsertPostComment(current, variables.postId, comment),
      );
      setCommentInputs((current) => ({ ...current, [variables.postId]: "" }));
      setEditingComments((current) => ({
        ...current,
        [variables.postId]: null,
      }));
      toast.success("Comment updated.");
    },
  });

  const collections = useMemo(
    () => getExploreCollections(categoriesQuery.data),
    [categoriesQuery.data],
  );
  const categories = useMemo(
    () => categoriesQuery.data ?? [],
    [categoriesQuery.data],
  );

  const loadedPosts = useMemo(
    () => postsQuery.data?.pages.flatMap((page) => page.items) ?? [],
    [postsQuery.data],
  );

  const savedBoardPosts = useMemo(
    () => loadedPosts.filter((post) => isPostSaved(post, savedPosts)),
    [loadedPosts, savedPosts],
  );

  const visiblePosts = useMemo(() => {
    const basePosts = (() => {
      if (showSavedBoard) {
        return savedBoardPosts;
      }

      if (!showsCommunityFeed(activeCollection)) {
        return activeCategorySlug ? loadedPosts : [];
      }

      return loadedPosts;
    })();

    const normalizedSearch = normalizedActiveSearch;
    if (!normalizedSearch) {
      return basePosts;
    }

    return basePosts.filter((post) =>
      matchesPostSearch(post, normalizedSearch),
    );
  }, [
    activeCollection,
    activeCategorySlug,
    loadedPosts,
    savedBoardPosts,
    normalizedActiveSearch,
    showSavedBoard,
  ]);

  const hasSearch = normalizedActiveSearch.length > 0;
  const searchResultsLabel = hasSearch
    ? `${visiblePosts.length} matching results`
    : null;
  const exploreMapPoints = useMemo<MapPoint[]>(
    () => {
      const clusters = new Map<
        string,
        {
          posts: Post[];
          latitude: number;
          longitude: number;
        }
      >();

      for (const post of visiblePosts) {
        if (
          !Number.isFinite(post.latitude) ||
          !Number.isFinite(post.longitude)
        ) {
          continue;
        }

        const key = `${post.latitude.toFixed(3)},${post.longitude.toFixed(3)}`;
        const cluster = clusters.get(key);
        if (cluster) {
          cluster.posts.push(post);
          continue;
        }

        clusters.set(key, {
          posts: [post],
          latitude: post.latitude,
          longitude: post.longitude,
        });
      }

      return Array.from(clusters.values()).map((cluster) => {
        const [firstPost] = cluster.posts;
        const postCount = cluster.posts.length;
        const locationName =
          firstPost.location_name || firstPost.caption || "Falzo destination";

        return {
          id: `location:${cluster.latitude.toFixed(3)},${cluster.longitude.toFixed(3)}`,
          name:
            postCount > 1
              ? `${postCount} travel posts near ${locationName}`
              : locationName,
          address: firstPost.caption || firstPost.user_name,
          count: postCount,
          imageUrl: firstPost.image_url,
          latitude: cluster.latitude,
          longitude: cluster.longitude,
          distanceMeters: getDistanceMeters(nearbyCoords, {
            latitude: cluster.latitude,
            longitude: cluster.longitude,
          }),
          postIds: cluster.posts.map((post) => post.id),
        };
      });
    },
    [nearbyCoords, visiblePosts],
  );
  const selectedMapPointId =
    selectedMapClusterId ??
    (selectedMapPostId === null
      ? null
      : (exploreMapPoints.find((point) =>
          point.postIds?.includes(selectedMapPostId),
        )?.id ?? null));
  const lastExploreMapPointsRef = useRef<MapPoint[]>([]);
  useEffect(() => {
    if (exploreMapPoints.length > 0) {
      lastExploreMapPointsRef.current = exploreMapPoints;
    }
  }, [exploreMapPoints]);
  const displayedExploreMapPoints =
    exploreMapPoints.length > 0 || !postsQuery.isFetching
      ? exploreMapPoints
      : lastExploreMapPointsRef.current;
  const selectedMapClusterPosts = useMemo(() => {
    if (!selectedMapClusterId) {
      return [];
    }

    const point = displayedExploreMapPoints.find(
      (item) => item.id === selectedMapClusterId,
    );
    if (!point?.postIds || point.postIds.length <= 1) {
      return [];
    }

    const selectedIds = new Set(point.postIds);
    return visiblePosts.filter((post) => selectedIds.has(post.id));
  }, [displayedExploreMapPoints, selectedMapClusterId, visiblePosts]);

  const selectedPost = useMemo(() => {
    if (selectedPostId === null) {
      return null;
    }

    return (
      postDetailQuery.data ??
      loadedPosts.find((post) => post.id === selectedPostId) ??
      null
    );
  }, [loadedPosts, postDetailQuery.data, selectedPostId]);

  const shouldLoadMorePosts =
    (showSavedBoard ||
      showsCommunityFeed(activeCollection) ||
      Boolean(activeCategorySlug)) &&
    postsQuery.hasNextPage &&
    !postsQuery.isFetchingNextPage;
  const fetchNextPostsPage = postsQuery.fetchNextPage;

  useEffect(() => {
    const node = loadMoreRef.current;
    if (!node || !shouldLoadMorePosts) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          void fetchNextPostsPage();
        }
      },
      { rootMargin: "700px 0px" },
    );

    observer.observe(node);
    return () => observer.disconnect();
  }, [fetchNextPostsPage, shouldLoadMorePosts]);

  const commentEventPostIds = useMemo(() => {
    const postIds = new Set(openComments);
    if (selectedChatPostId !== null) {
      postIds.add(selectedChatPostId);
    }

    return Array.from(postIds).sort((left, right) => left - right);
  }, [openComments, selectedChatPostId]);
  const commentEventPostIdsKey = commentEventPostIds.join(",");

  useEffect(() => {
    if (!commentEventPostIdsKey) {
      return;
    }

    const eventSources = commentEventPostIds.map((postId) => {
      const source = new EventSource(getPostCommentEventsUrl(postId));
      const handleCommentCreated = (event: Event) => {
        const comment = parsePostCommentCreatedEvent(
          event as MessageEvent<string>,
        );
        if (!comment || comment.post_id !== postId) {
          return;
        }

        setCommentsByPost((current) =>
          upsertPostComment(current, postId, comment),
        );
      };

      source.addEventListener("comment.created", handleCommentCreated);
      source.addEventListener("comment.updated", handleCommentCreated);
      return () => {
        source.removeEventListener("comment.created", handleCommentCreated);
        source.removeEventListener("comment.updated", handleCommentCreated);
        source.close();
      };
    });

    return () => {
      eventSources.forEach((closeEventSource) => closeEventSource());
    };
  }, [commentEventPostIds, commentEventPostIdsKey]);

  function requireAuth() {
    if (hasAuthSession()) {
      setIsAuthenticated(true);
      return true;
    }

    setIsAuthenticated(false);
    toast.error("Login is required for this action.");
    router.push(ROUTES.login);
    return false;
  }

  function updateComment(postId: number, value: string) {
    setCommentInputs((current) => ({ ...current, [postId]: value }));
  }

  function replyToComment(postId: number, comment: PostComment) {
    setReplyTargets((current) => ({ ...current, [postId]: comment }));
    setEditingComments((current) => ({ ...current, [postId]: null }));
    setOpenComments((current) => new Set(current).add(postId));
    if (selectedPostId === postId) {
      setSelectedChatPostId(postId);
    }
    globalThis.setTimeout(() => {
      commentInputRefs.current[postId]?.focus();
    }, 0);
  }

  function cancelReply(postId: number) {
    setReplyTargets((current) => ({ ...current, [postId]: null }));
  }

  function editComment(postId: number, comment: PostComment) {
    if (!requireAuth()) {
      return;
    }

    setEditingComments((current) => ({ ...current, [postId]: comment }));
    setReplyTargets((current) => ({ ...current, [postId]: null }));
    setCommentInputs((current) => ({ ...current, [postId]: comment.content }));
    setOpenComments((current) => new Set(current).add(postId));
    if (selectedPostId === postId) {
      setSelectedChatPostId(postId);
    }
    globalThis.setTimeout(() => {
      commentInputRefs.current[postId]?.focus();
    }, 0);
  }

  function cancelEdit(postId: number) {
    setEditingComments((current) => ({ ...current, [postId]: null }));
    setCommentInputs((current) => ({ ...current, [postId]: "" }));
  }

  async function loadComments(postId: number) {
    if (commentsByPost[postId] || loadingComments.has(postId)) {
      return;
    }

    setLoadingComments((current) => new Set(current).add(postId));

    try {
      const comments = await getPostCommentsApi(postId);
      setCommentsByPost((current) =>
        mergePostComments(current, postId, comments),
      );
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    } finally {
      setLoadingComments((current) => {
        const next = new Set(current);
        next.delete(postId);
        return next;
      });
    }
  }

  function toggleComments(postId: number) {
    const willOpen = !openComments.has(postId);

    setOpenComments((current) => toggleSetValue(current, postId));

    if (willOpen) {
      void loadComments(postId);
    }
  }

  function openPostDetail(postId: number) {
    setSelectedPostId(postId);
    setSelectedChatPostId(postId);
    void loadComments(postId);
  }

  function openNotificationTarget(notification: AppNotification) {
    markNotificationIdsRead([notification.id]);
    const postId = getNotificationPostId(notification);
    if (postId !== null) {
      openPostDetail(postId);
      return;
    }

    const actorUserId = Number(notification.actor_user_id);
    if (Number.isFinite(actorUserId) && actorUserId > 0) {
      router.push(ROUTES.userProfile(actorUserId));
    }
  }

  function openSelectedPostChat() {
    if (selectedPostId === null) {
      return;
    }

    setSelectedChatPostId(selectedPostId);
    void loadComments(selectedPostId);
  }

  function handleLikePost(postId: number) {
    const post =
      selectedPost?.id === postId
        ? selectedPost
        : (visiblePosts.find((item) => item.id === postId) ?? null);
    if (!requireAuth() || !post || likeMutation.isPending) {
      return;
    }

    likeMutation.mutate({ isLiked: isPostLiked(post, likedPosts), postId });
  }

  function handleSavePost(postId: number) {
    const post =
      selectedPost?.id === postId
        ? selectedPost
        : (loadedPosts.find((item) => item.id === postId) ?? null);
    if (!requireAuth() || !post || saveMutation.isPending) {
      return;
    }

    saveMutation.mutate({ isSaved: isPostSaved(post, savedPosts), postId });
  }

  function handleDeletePost(postId: number) {
    if (!requireAuth() || deletePostMutation.isPending) {
      return;
    }
    if (!globalThis.window?.confirm("Delete this travel post?")) {
      return;
    }

    deletePostMutation.mutate(postId);
  }

  function handleReportPost(postId: number) {
    if (!requireAuth() || reportPostMutation.isPending) {
      return;
    }

    const reason = globalThis.window
      ?.prompt("Why are you reporting this travel post?")
      ?.trim();
    if (!reason) {
      return;
    }

    reportPostMutation.mutate({ postId, reason });
  }

  function toggleSavedBoard() {
    if (!requireAuth()) {
      return;
    }

    router.push(ROUTES.saved);
  }

  function changeCollection(collection: string) {
    if (collection === FOLLOWING_COLLECTION && !requireAuth()) {
      return;
    }

    setShowSavedBoard(false);
    setActiveCollection(collection);
  }

  function locateNearbyFeed() {
    if (!globalThis.navigator?.geolocation) {
      toast.error("Location is not available in this browser.");
      return;
    }

    globalThis.navigator.geolocation.getCurrentPosition(
      (position) => {
        setNearbyCoords({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
        });
        setNearbyPlaceName("Your location");
        setFeedSort("nearby");
      },
      () => toast.error("Location permission is required for nearby feed."),
      { enableHighAccuracy: true, maximumAge: 60_000, timeout: 10_000 },
    );
  }

  function changeFeedSort(nextSort: PostSort) {
    if (nextSort !== "nearby") {
      setFeedSort(nextSort);
      return;
    }

    setShowMapPanel(true);
    locateNearbyFeed();
  }

  function applyTravelPrompt(query: string, sort: PostSort) {
    setShowSavedBoard(false);
    setShowFirstLoadSuggestions(false);
    setActiveCollection(ALL_COLLECTION);
    setSearchValue(query);
    changeFeedSort(sort);
  }

  function submitComment(postId: number) {
    if (!requireAuth()) {
      return;
    }

    const content = commentInputs[postId]?.trim() ?? "";
    if (!content) {
      toast.error("Comment content is required.");
      return;
    }

    const editingComment = editingComments[postId];
    if (editingComment) {
      updateCommentMutation.mutate({
        postId,
        commentId: editingComment.id,
        content,
      });
      return;
    }

    commentMutation.mutate({
      postId,
      content,
      replyToCommentId: replyTargets[postId]?.id,
    });
  }

  const selectedPostComments =
    selectedPostId === null ? [] : (commentsByPost[selectedPostId] ?? []);
  const selectedPostCommentValue =
    selectedPostId === null ? "" : (commentInputs[selectedPostId] ?? "");
  const selectedPostReplyTarget =
    selectedPostId === null ? null : (replyTargets[selectedPostId] ?? null);
  const selectedPostEditingComment =
    selectedPostId === null ? null : (editingComments[selectedPostId] ?? null);
  const isSelectedPostChatOpen =
    selectedPostId !== null && selectedChatPostId === selectedPostId;
  const isSelectedPostLoadingComments =
    isSelectedPostChatOpen &&
    selectedPostId !== null &&
    loadingComments.has(selectedPostId);
  const isSelectedPostLiked = isPostLiked(selectedPost, likedPosts);
  const isSelectedPostSaved = isPostSaved(selectedPost, savedPosts);
  const selectedMapClusterIndex =
    selectedPostId === null
      ? -1
      : selectedMapClusterPosts.findIndex((post) => post.id === selectedPostId);
  const hasSelectedMapClusterCarousel =
    selectedMapClusterPosts.length > 1 && selectedMapClusterIndex >= 0;
  const selectedMapClusterCarouselLabel = hasSelectedMapClusterCarousel
    ? `${selectedMapClusterIndex + 1} / ${selectedMapClusterPosts.length}`
    : undefined;
  const canUsePreviousMapClusterPost =
    hasSelectedMapClusterCarousel &&
    selectedMapClusterPosts.length >= 2 &&
    selectedMapClusterIndex > 0;
  const canUseNextMapClusterPost =
    hasSelectedMapClusterCarousel &&
    selectedMapClusterIndex < selectedMapClusterPosts.length - 1;

  function openMapClusterPostAt(index: number) {
    if (!hasSelectedMapClusterCarousel) {
      return;
    }

    const normalizedIndex =
      Math.max(0, Math.min(index, selectedMapClusterPosts.length - 1));
    const nextPost = selectedMapClusterPosts[normalizedIndex];
    setSelectedMapPostId(nextPost.id);
    setSelectedPostId(nextPost.id);
    setSelectedChatPostId(null);
  }

  useEffect(() => {
    if (!hasSelectedMapClusterCarousel) {
      return;
    }

    const preloadIndexes = [
      selectedMapClusterIndex - 1,
      selectedMapClusterIndex + 1,
    ];
    for (const index of preloadIndexes) {
      const imageURL = selectedMapClusterPosts[index]?.image_url;
      if (!imageURL) {
        continue;
      }

      const image = new Image();
      image.src = imageURL;
    }
  }, [
    hasSelectedMapClusterCarousel,
    selectedMapClusterIndex,
    selectedMapClusterPosts,
  ]);

  return (
    <main className="min-h-screen bg-[#f7f7f5] pb-[8.5rem] text-[#1f1f1f] sm:pb-0">
      <ExploreTopbar
        isAuthenticated={isAuthenticated}
        onClearSearch={() => setSearchValue("")}
        onProfileClick={() => {
          router.push(hasAuthSession() ? ROUTES.profile : ROUTES.login);
        }}
        notifications={notifications}
        onNotificationsOpen={markNotificationsRead}
        onNotificationSelect={openNotificationTarget}
        onSearchChange={setSearchValue}
        profileName={profileName}
        searchValue={searchValue}
        unreadNotificationCount={unreadNotificationCount}
      />

      <ExploreHero
        activeCollection={activeCollection}
        collections={collections}
        featuredPosts={visiblePosts.slice(0, 3)}
        feedSort={feedSort}
        isAuthenticated={isAuthenticated}
        onCollectionChange={changeCollection}
        onOpenPost={openPostDetail}
        onSortChange={changeFeedSort}
        onToggleSavedBoard={toggleSavedBoard}
        savedBoardCount={savedBoardPosts.length}
        showSavedBoard={showSavedBoard}
      />

      {showFirstLoadSuggestions ? (
        <FirstLoadTravelSuggestions
          onDismiss={() => setShowFirstLoadSuggestions(false)}
          onSelectSuggestion={applyTravelPrompt}
        />
      ) : null}

      {activeCollection === ALL_COLLECTION && !showSavedBoard ? (
        <ExploreCategoryBar
          categories={categories}
          onSelectCategory={(category) => {
            setShowSavedBoard(false);
            setActiveCollection(category.name);
          }}
        />
      ) : null}

      <section className="mx-auto w-full max-w-370 px-3 pb-10 sm:px-6 sm:pb-14 lg:px-8">
        {showMapPanel ? (
          <ExploreMapPanel
            currentPosition={nearbyCoords}
            feedSort={feedSort}
            onClearNearby={() => {
              setNearbyCoords(null);
              setNearbyPlaceName(null);
              if (feedSort === "nearby") {
                setFeedSort("newest");
              }
            }}
            onClose={() => setShowMapPanel(false)}
            onLocate={locateNearbyFeed}
            onMapCenterChange={(coordinates) => {
              setNearbyCoords(coordinates);
              setNearbyPlaceName("Selected map area");
              setFeedSort("nearby");
            }}
            onPlaceSearch={(coordinates, placeName) => {
              setNearbyCoords(coordinates);
              setNearbyPlaceName(placeName);
              setFeedSort("nearby");
            }}
            onPointSelect={(point) => {
              setSelectedMapClusterId(point.id);
              const postIds = point.postIds ?? [];
              if (postIds.length === 1) {
                setSelectedMapPostId(postIds[0]);
                openPostDetail(postIds[0]);
                return;
              }
              if (postIds.length > 1) {
                setSelectedMapPostId(postIds[0]);
                openPostDetail(postIds[0]);
              }
            }}
            onRadiusChange={setNearbyRadiusMeters}
            points={displayedExploreMapPoints}
            placeName={nearbyPlaceName}
            radiusMeters={nearbyRadiusMeters}
            selectedPointId={selectedMapPointId}
            totalPosts={visiblePosts.length}
          />
        ) : (
          <MapAccessPanel
            onLocate={() => {
              setShowMapPanel(true);
              locateNearbyFeed();
            }}
            onOpen={() => setShowMapPanel(true)}
            totalPosts={visiblePosts.length}
          />
        )}

        <MapClusterPostRail
          onOpenPost={(postId) => {
            setSelectedMapPostId(postId);
            openPostDetail(postId);
          }}
          posts={selectedMapClusterPosts}
        />

        {showSavedBoard || searchResultsLabel ? (
          <div className="mb-4 flex flex-col items-stretch justify-between gap-3 rounded-3xl border border-black/6 bg-white px-4 py-3 shadow-[0_14px_36px_-30px_rgb(0_0_0/0.58)] sm:flex-row sm:items-center">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
                {showSavedBoard ? "Trip board" : "Search"}
              </p>
              <h2 className="text-lg font-semibold tracking-normal text-[#111]">
                {showSavedBoard ? "Saved destinations" : searchResultsLabel}
              </h2>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              {searchResultsLabel ? (
                <Button
                  className="rounded-full"
                  onClick={() => setSearchValue("")}
                  type="button"
                  variant="outline"
                >
                  Clear search
                </Button>
              ) : null}
              {showSavedBoard ? (
                <Button
                  className="rounded-full"
                  onClick={toggleSavedBoard}
                  type="button"
                  variant="outline"
                >
                  Show all
                </Button>
              ) : null}
            </div>
          </div>
        ) : null}
        <div className="columns-1 gap-4 sm:columns-2 lg:columns-3 2xl:columns-4">
          {visiblePosts.map((post, index) => (
            <ExplorePostCard
              commentValue={commentInputs[post.id] ?? ""}
              comments={commentsByPost[post.id] ?? []}
              commentsOpen={openComments.has(post.id)}
              currentUserId={currentUserId}
              editingComment={editingComments[post.id] ?? null}
              index={index}
              isAuthenticated={isAuthenticated}
              isLiked={isPostLiked(post, likedPosts)}
              isLoadingComments={loadingComments.has(post.id)}
              isSaved={isPostSaved(post, savedPosts)}
              isSubmittingComment={
                commentMutation.isPending || updateCommentMutation.isPending
              }
              key={`post-${post.id}`}
              onCancelEdit={cancelEdit}
              onCancelReply={cancelReply}
              onCommentChange={updateComment}
              onEditComment={editComment}
              onDelete={handleDeletePost}
              onLike={handleLikePost}
              onOpen={openPostDetail}
              onReport={handleReportPost}
              onRegisterCommentInput={(postId, node) => {
                commentInputRefs.current[postId] = node;
              }}
              onReplyComment={replyToComment}
              onSave={handleSavePost}
              onSubmitComment={submitComment}
              onToggleComments={toggleComments}
              post={post}
              replyTarget={replyTargets[post.id] ?? null}
            />
          ))}
        </div>

        {(showSavedBoard || hasSearch) &&
        visiblePosts.length === 0 &&
        !postsQuery.isLoading ? (
          <div className="flex min-h-64 flex-col items-center justify-center rounded-3xl border border-dashed border-black/10 bg-white/72 px-6 text-center">
            {showSavedBoard ? (
              <Bookmark className="size-8 text-[#777]" />
            ) : (
              <Search className="size-8 text-[#777]" />
            )}
            <h2 className="mt-3 text-xl font-semibold tracking-normal text-[#111]">
              {showSavedBoard
                ? "No saved destinations yet"
                : "No destinations match your search"}
            </h2>
            <p className="mt-2 max-w-sm text-sm leading-6 text-[#666]">
              {showSavedBoard
                ? "Save places you want to visit and build a shortlist for your next trip."
                : "Try a city, landmark, travel theme, or creator name."}
            </p>
          </div>
        ) : null}

        {showSavedBoard ||
        showsCommunityFeed(activeCollection) ||
        activeCategorySlug ? (
          <LoadMorePosts
            hasNextPage={Boolean(postsQuery.hasNextPage)}
            isLoading={postsQuery.isFetchingNextPage}
            refNode={loadMoreRef}
            totalPosts={visiblePosts.length}
          />
        ) : null}
      </section>
      <PostDetailDialog
        carouselLabel={selectedMapClusterCarouselLabel}
        commentValue={selectedPostCommentValue}
        comments={selectedPostComments}
        currentUserId={currentUserId}
        editingComment={selectedPostEditingComment}
        isAuthenticated={isAuthenticated}
        isChatOpen={isSelectedPostChatOpen}
        isCommentPending={
          commentMutation.isPending || updateCommentMutation.isPending
        }
        isLiked={isSelectedPostLiked}
        isLoadingComments={isSelectedPostLoadingComments}
        isSaved={isSelectedPostSaved}
        replyTarget={selectedPostReplyTarget}
        onNextPost={
          canUseNextMapClusterPost
            ? () => openMapClusterPostAt(selectedMapClusterIndex + 1)
            : undefined
        }
        onPreviousPost={
          canUsePreviousMapClusterPost
            ? () => openMapClusterPostAt(selectedMapClusterIndex - 1)
            : undefined
        }
        onClose={() => {
          setSelectedPostId(null);
        }}
        onCancelReply={() => {
          if (selectedPostId !== null) {
            cancelReply(selectedPostId);
          }
        }}
        onCancelEdit={() => {
          if (selectedPostId !== null) {
            cancelEdit(selectedPostId);
          }
        }}
        onCommentChange={(value) => {
          if (selectedPostId !== null) {
            updateComment(selectedPostId, value);
          }
        }}
        onRegisterCommentInput={(node) => {
          if (selectedPostId !== null) {
            commentInputRefs.current[selectedPostId] = node;
          }
        }}
        onLike={() => {
          if (selectedPostId !== null) {
            handleLikePost(selectedPostId);
          }
        }}
        onLoadComments={openSelectedPostChat}
        onEditComment={(comment) => {
          if (selectedPostId !== null) {
            editComment(selectedPostId, comment);
          }
        }}
        onReplyComment={(comment) => {
          if (selectedPostId !== null) {
            replyToComment(selectedPostId, comment);
          }
        }}
        onSave={() => {
          if (selectedPostId !== null) {
            handleSavePost(selectedPostId);
          }
        }}
        onSubmitComment={() => {
          if (selectedPostId !== null) {
            submitComment(selectedPostId);
          }
        }}
        open={selectedPostId !== null}
        post={selectedPost}
      />
    </main>
  );
}

function ExploreTopbar({
  isAuthenticated,
  notifications,
  onClearSearch,
  onNotificationsOpen,
  onNotificationSelect,
  onProfileClick,
  onSearchChange,
  profileName,
  searchValue,
  unreadNotificationCount,
}: Readonly<{
  isAuthenticated: boolean;
  notifications: AppNotification[];
  onClearSearch: () => void;
  onNotificationsOpen: () => void;
  onNotificationSelect: (notification: AppNotification) => void;
  onProfileClick: () => void;
  onSearchChange: (value: string) => void;
  profileName: string | null;
  searchValue: string;
  unreadNotificationCount: number;
}>) {
  const profileLabel =
    isAuthenticated && profileName ? `Profile: ${profileName}` : "Profile";
  const authLabel = isAuthenticated ? profileLabel : "Login";

  return (
    <>
    <header className="sticky top-0 z-40 hidden border-b border-black/6 bg-[#f7f7f5]/86 backdrop-blur-2xl sm:block">
      <div className="mx-auto flex w-full max-w-370 items-center gap-2 px-2.5 py-2.5 sm:px-5 sm:py-3 lg:px-8">
        <Link
          aria-label="Explore"
          className="inline-flex size-9 shrink-0 items-center justify-center rounded-full bg-[#111] text-white shadow-[0_14px_30px_-20px_rgb(0_0_0/0.72)] transition hover:scale-[1.03] sm:size-10"
          href={ROUTES.explore}
        >
          <Camera className="size-4" />
        </Link>

        <nav className="hidden items-center gap-1 md:flex">
          <Button
            className="rounded-full bg-[#111] text-white hover:bg-[#222]"
            size="sm"
          >
            Explore
          </Button>
          <Button asChild className="rounded-full" size="sm" variant="ghost">
            <Link href={ROUTES.locations}>
              <MapIcon className="size-4" />
              Destinations
            </Link>
          </Button>
          {isAuthenticated ? (
            <Button asChild className="rounded-full" size="sm" variant="ghost">
              <Link href={ROUTES.saved}>
                <Bookmark className="size-4" />
                Saved
              </Link>
            </Button>
          ) : null}
        </nav>

        <div className="relative ml-0 min-w-0 flex-1 sm:ml-1">
          <Search className="-translate-y-1/2 pointer-events-none absolute left-3 top-1/2 size-4 text-[#777] sm:left-4" />
          <input
            className="h-10 w-full rounded-full border border-black/6 bg-white px-9 text-sm text-[#1f1f1f] shadow-[0_12px_32px_-28px_rgb(0_0_0/0.45)] outline-none transition placeholder:text-[#8a8a8a] focus:border-black/10 focus:bg-white focus:shadow-[0_18px_40px_-30px_rgb(0_0_0/0.58)] sm:h-11 sm:px-11"
            onChange={(event) => onSearchChange(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Escape" && searchValue) {
                onClearSearch();
              }
            }}
            placeholder="Search destinations, cities, beaches, mountains, culture"
            type="search"
            value={searchValue}
          />
          <Button
            aria-label={searchValue ? "Clear search" : "Search filters"}
            className="-translate-y-1/2 absolute right-1.5 top-1/2 rounded-full"
            onClick={searchValue ? onClearSearch : undefined}
            size="icon-sm"
            type="button"
            variant="ghost"
          >
            {searchValue ? (
              <X className="size-4" />
            ) : (
              <SlidersHorizontal className="size-4" />
            )}
          </Button>
        </div>

        <div className="hidden items-center gap-1 sm:flex">
          {isAuthenticated ? (
            <>
              <Button
                aria-label="Create travel post"
                asChild
                className="rounded-full"
                size="icon-sm"
                variant="ghost"
              >
                <Link href={ROUTES.upload}>
                  <Plus className="size-4" />
                </Link>
              </Button>
              <NotificationBell
                notifications={notifications}
                onOpen={onNotificationsOpen}
                onSelectNotification={onNotificationSelect}
                unreadCount={unreadNotificationCount}
              />
            </>
          ) : null}
          <Button
            aria-label={profileLabel}
            className={cn(
              "rounded-full",
              isAuthenticated && profileName ? "max-w-44 px-3" : "",
            )}
            onClick={onProfileClick}
            size={isAuthenticated && !profileName ? "icon-sm" : "sm"}
            type="button"
            variant="outline"
          >
            <UserRound className="size-4" />
            {isAuthenticated && profileName ? (
              <span className="max-w-28 truncate">{profileName}</span>
            ) : !isAuthenticated ? (
              <span>Login</span>
            ) : null}
          </Button>
        </div>

        <Sheet>
          <SheetTrigger asChild>
            <Button
              aria-label="Menu"
              className="rounded-full sm:hidden"
              size="icon-sm"
              type="button"
              variant="ghost"
            >
              <Menu className="size-4" />
            </Button>
          </SheetTrigger>
          <SheetContent side="right">
            <SheetHeader>
              <SheetTitle>Explore menu</SheetTitle>
              <SheetDescription>
                Find destinations, open saved places, and manage your account.
              </SheetDescription>
            </SheetHeader>
            <div className="space-y-2 px-5">
              <SheetClose asChild>
                <Link
                  className={cn(
                    buttonVariants({ size: "default", variant: "outline" }),
                    "w-full justify-start rounded-full",
                  )}
                  href={ROUTES.locations}
                >
                  <MapIcon className="size-4" />
                  Destinations
                </Link>
              </SheetClose>
              {isAuthenticated ? (
                <>
                  <SheetClose asChild>
                    <Link
                      className={cn(
                        buttonVariants({ size: "default", variant: "outline" }),
                        "w-full justify-start rounded-full",
                      )}
                      href={ROUTES.saved}
                    >
                      <Bookmark className="size-4" />
                      Saved
                    </Link>
                  </SheetClose>
                  <SheetClose asChild>
                    <Link
                      className={cn(
                        buttonVariants({ size: "default", variant: "outline" }),
                        "w-full justify-start rounded-full",
                      )}
                      href={ROUTES.upload}
                    >
                      <Plus className="size-4" />
                      Upload
                    </Link>
                  </SheetClose>
                  <div className="flex items-center justify-between rounded-full border border-black/6 px-4 py-2">
                    <span className="text-sm font-medium text-[#1f1f1f]">
                      Notifications
                    </span>
                    <NotificationBell
                      notifications={notifications}
                      onOpen={onNotificationsOpen}
                      onSelectNotification={onNotificationSelect}
                      unreadCount={unreadNotificationCount}
                    />
                  </div>
                </>
              ) : null}
              <SheetClose asChild>
                <button
                  className={cn(
                    buttonVariants({
                      size: "default",
                      variant: isAuthenticated ? "outline" : "default",
                    }),
                    "w-full justify-start rounded-full",
                  )}
                  onClick={onProfileClick}
                  type="button"
                >
                  <UserRound className="size-4" />
                  {authLabel}
                </button>
              </SheetClose>
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </header>

    <nav className="fixed inset-x-0 bottom-0 z-50 border-t border-black/8 bg-[#f7f7f5]/94 px-3 pb-[calc(env(safe-area-inset-bottom)+0.65rem)] pt-2 shadow-[0_-18px_48px_-34px_rgb(0_0_0/0.72)] backdrop-blur-2xl sm:hidden">
      <div className="mx-auto max-w-md">
        <div className="relative mb-2">
          <Search className="-translate-y-1/2 pointer-events-none absolute left-3 top-1/2 size-4 text-[#777]" />
          <input
            className="h-10 w-full rounded-full border border-black/8 bg-white px-9 text-sm text-[#1f1f1f] shadow-[0_10px_28px_-24px_rgb(0_0_0/0.5)] outline-none placeholder:text-[#8a8a8a] focus:border-black/14"
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="Search destinations"
            type="search"
            value={searchValue}
          />
          <button
            aria-label={searchValue ? "Clear search" : "Search"}
            className="-translate-y-1/2 absolute right-1.5 top-1/2 flex size-8 items-center justify-center rounded-full text-[#555] transition hover:bg-black/5"
            onClick={searchValue ? onClearSearch : undefined}
            type="button"
          >
            {searchValue ? <X className="size-4" /> : <SlidersHorizontal className="size-4" />}
          </button>
        </div>

        <div className={cn("grid items-center gap-1", isAuthenticated ? "grid-cols-5" : "grid-cols-3")}>
          <Link
            aria-label="Explore"
            className="flex flex-col items-center gap-1 rounded-2xl bg-[#111] px-2 py-2 text-[11px] font-semibold text-white"
            href={ROUTES.explore}
          >
            <Camera className="size-4" />
            Explore
          </Link>
          <Link
            aria-label="Destinations"
            className="flex flex-col items-center gap-1 rounded-2xl px-2 py-2 text-[11px] font-semibold text-[#555] transition hover:bg-white"
            href={ROUTES.locations}
          >
            <MapIcon className="size-4" />
            Places
          </Link>
          {isAuthenticated ? (
            <>
              <Link
                aria-label="Upload"
                className="mx-auto flex size-12 items-center justify-center rounded-full bg-[#ff385c] text-white shadow-[0_16px_34px_-22px_rgb(255_56_92/0.85)]"
                href={ROUTES.upload}
              >
                <Plus className="size-5" />
              </Link>
              <Link
                aria-label="Saved"
                className="flex flex-col items-center gap-1 rounded-2xl px-2 py-2 text-[11px] font-semibold text-[#555] transition hover:bg-white"
                href={ROUTES.saved}
              >
                <Bookmark className="size-4" />
                Saved
              </Link>
            </>
          ) : null}
          <button
            aria-label={profileLabel}
            className="flex flex-col items-center gap-1 rounded-2xl px-2 py-2 text-[11px] font-semibold text-[#555] transition hover:bg-white"
            onClick={onProfileClick}
            type="button"
          >
            <UserRound className="size-4" />
            {isAuthenticated ? "Account" : "Login"}
          </button>
        </div>
      </div>
    </nav>
    </>
  );
}

function MapClusterPostRail({
  onOpenPost,
  posts,
}: Readonly<{
  onOpenPost: (postId: number) => void;
  posts: Post[];
}>) {
  if (posts.length <= 1) {
    return null;
  }

  return (
    <section className="mb-5 rounded-3xl border border-black/6 bg-white px-4 py-4 shadow-[0_14px_36px_-30px_rgb(0_0_0/0.58)]">
      <div className="mb-3 flex items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
            Same destination
          </p>
          <h2 className="truncate text-lg font-semibold tracking-normal text-[#111]">
            {posts.length} travel posts here
          </h2>
        </div>
      </div>

      <div className="flex snap-x gap-3 overflow-x-auto pb-1 scrollbar-none [&::-webkit-scrollbar]:hidden">
        {posts.map((post) => (
          <button
            className="group relative h-48 w-64 shrink-0 snap-start overflow-hidden rounded-2xl border border-black/6 bg-[#e9eef3] text-left shadow-[0_16px_38px_-30px_rgb(0_0_0/0.65)] transition hover:-translate-y-0.5"
            key={post.id}
            onClick={() => onOpenPost(post.id)}
            type="button"
          >
            <img
              alt={post.caption || post.location_name || "Destination photo"}
              className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.035]"
              loading="lazy"
              src={post.image_url}
            />
            <div className="absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0.02)_0%,rgb(0_0_0/0.08)_42%,rgb(0_0_0/0.68)_100%)]" />
            <div className="absolute inset-x-3 bottom-3 text-white">
              <p className="line-clamp-1 text-xs font-semibold uppercase tracking-[0.14em] text-white/76">
                {post.location_name || "Destination"}
              </p>
              <p className="mt-1 line-clamp-2 text-base font-semibold leading-tight">
                {post.caption || post.user_name || "Travel story"}
              </p>
            </div>
          </button>
        ))}
      </div>
    </section>
  );
}

function MapAccessPanel({
  onLocate,
  onOpen,
  totalPosts,
}: Readonly<{
  onLocate: () => void;
  onOpen: () => void;
  totalPosts: number;
}>) {
  return (
    <section className="mb-5 rounded-2xl border border-black/6 bg-white px-4 py-3 shadow-[0_14px_36px_-30px_rgb(0_0_0/0.58)]">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 items-center gap-3">
          <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-[#111] text-white">
            <MapIcon className="size-5" />
          </span>
          <div className="min-w-0">
            <p className="text-sm font-semibold text-[#111]">
              Map view is optional
            </p>
            <p className="text-xs leading-5 text-[#777]">
              Browse the feed first, or open the map for {totalPosts} visible
              destination{totalPosts === 1 ? "" : "s"}.
            </p>
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button className="rounded-full" onClick={onOpen} type="button">
            <MapIcon className="size-4" />
            Open map
          </Button>
          <Button
            className="rounded-full"
            onClick={onLocate}
            type="button"
            variant="outline"
          >
            <LocateFixed className="size-4" />
            Near me
          </Button>
        </div>
      </div>
    </section>
  );
}

function ExploreMapPanel({
  currentPosition,
  feedSort,
  onClearNearby,
  onClose,
  onLocate,
  onMapCenterChange,
  onPlaceSearch,
  onPointSelect,
  onRadiusChange,
  placeName,
  points,
  radiusMeters,
  selectedPointId,
  totalPosts,
}: Readonly<{
  currentPosition: Coordinates | null;
  feedSort: PostSort;
  onClearNearby: () => void;
  onClose: () => void;
  onLocate: () => void;
  onMapCenterChange: (coordinates: Coordinates) => void;
  onPlaceSearch: (coordinates: Coordinates, placeName: string) => void;
  onPointSelect: (point: MapPoint) => void;
  onRadiusChange: (radiusMeters: number) => void;
  placeName: string | null;
  points: MapPoint[];
  radiusMeters: number;
  selectedPointId: string | null;
  totalPosts: number;
}>) {
  const hasNearbyCenter = currentPosition !== null;
  const [placeQuery, setPlaceQuery] = useState("");
  const [isSearchingPlace, setIsSearchingPlace] = useState(false);
  const [customRadiusKm, setCustomRadiusKm] = useState(
    String(Math.round(radiusMeters / 1000)),
  );

  useEffect(() => {
    setCustomRadiusKm(String(Math.round(radiusMeters / 1000)));
  }, [radiusMeters]);

  function applyCustomRadius(value: string) {
    const radiusKm = Number(value);
    const nextRadiusMeters = clampNearbyRadiusMeters(radiusKm * 1000);
    setCustomRadiusKm(String(Math.round(nextRadiusMeters / 1000)));
    onRadiusChange(nextRadiusMeters);
  }

  async function searchPlace() {
    const query = normalizeLocationSearchQuery(placeQuery);
    if (!query) {
      return;
    }

    setIsSearchingPlace(true);
    try {
      const [location] = await searchLocationsWithFallbackApi(query);
      if (!location) {
        toast.error("No city or province found.");
        return;
      }

      onPlaceSearch(
        {
          latitude: location.latitude,
          longitude: location.longitude,
        },
        location.name,
      );
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    } finally {
      setIsSearchingPlace(false);
    }
  }

  return (
    <section className="mb-5 overflow-hidden rounded-2xl border border-black/6 bg-white shadow-[0_18px_46px_-36px_rgb(0_0_0/0.55)] sm:rounded-3xl">
      <div className="grid min-h-[620px] gap-0 xl:grid-cols-[minmax(0,1fr)_20rem]">
        <div className="min-h-[420px] xl:min-h-[620px]">
          <MapClient
            className="h-full rounded-none border-0 shadow-none"
            currentPosition={currentPosition}
            currentPositionLabel={placeName ?? "Selected area"}
            height="large"
            onSelectCoordinates={onMapCenterChange}
            onSelectPoint={onPointSelect}
            points={points}
            selectedPointId={selectedPointId}
            zoom={feedSort === "nearby" ? 12 : 5}
          />
        </div>

        <aside className="flex flex-col justify-between gap-4 border-black/6 border-t p-4 sm:gap-5 sm:p-5 xl:border-l xl:border-t-0">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
              Travel map
            </p>
            <div className="mt-1 flex items-start justify-between gap-3">
              <h2 className="text-xl font-semibold leading-tight tracking-normal text-[#111] sm:text-2xl">
                Explore places worth visiting
              </h2>
              <Button
                aria-label="Close map"
                className="shrink-0 rounded-full"
                onClick={onClose}
                size="icon-sm"
                type="button"
                variant="ghost"
              >
                <X className="size-4" />
              </Button>
            </div>
            <p className="mt-2 text-sm leading-6 text-[#666]">
              Browse photos, locations, and real travel inspiration to choose
              the right place for your next itinerary.
            </p>
          </div>

          <div className="grid gap-3">
            <form
              className="rounded-2xl border border-black/6 bg-[#f7f7f5] p-3"
              onSubmit={(event) => {
                event.preventDefault();
                void searchPlace();
              }}
            >
              <label className="text-xs font-semibold uppercase tracking-[0.14em] text-[#777]">
                City / province
              </label>
              <div className="mt-2 grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
                <input
                  className="h-10 min-w-0 rounded-full border border-black/8 bg-white px-3 text-sm font-semibold text-[#333] outline-none transition placeholder:text-[#999] focus:border-black/20"
                  onChange={(event) => setPlaceQuery(event.target.value)}
                  placeholder="Da Nang, Ha Noi, Bangkok"
                  type="search"
                  value={placeQuery}
                />
                <Button
                  className="w-full rounded-full sm:w-auto"
                  disabled={isSearchingPlace}
                  size="sm"
                  type="submit"
                  variant="outline"
                >
                  <Search className="size-4" />
                  Search
                </Button>
              </div>
            </form>

            <div className="rounded-2xl border border-black/6 bg-[#f7f7f5] px-4 py-3">
              <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#777]">
                Visible
              </p>
              <p className="mt-1 text-lg font-semibold text-[#111]">
                {totalPosts} destination{totalPosts === 1 ? "" : "s"}
              </p>
            </div>

            {hasNearbyCenter ? (
              <div className="rounded-2xl border border-[#c8ddf1] bg-[#f2f7fd] px-4 py-3 text-sm text-[#385c80]">
                <p className="font-semibold">
                  {placeName ?? "Nearby center"}
                </p>
                <p className="mt-1">
                  {currentPosition.latitude.toFixed(5)},{" "}
                  {currentPosition.longitude.toFixed(5)}
                </p>
              </div>
            ) : null}

            <div>
              <p className="mb-2 text-xs font-semibold uppercase tracking-[0.14em] text-[#777]">
                Radius
              </p>
              <div className="flex flex-wrap gap-2">
                {nearbyRadiusOptions.map((option) => (
                  <button
                    className={cn(
                      "rounded-full border px-3 py-1.5 text-xs font-semibold transition",
                      radiusMeters === option
                        ? "border-[#111] bg-[#111] text-white"
                        : "border-black/8 bg-white text-[#444] hover:border-black/16",
                    )}
                    key={option}
                    onClick={() => onRadiusChange(option)}
                    type="button"
                  >
                    {formatRadiusLabel(option)}
                  </button>
                ))}
              </div>
              <form
                className="mt-3 grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto]"
                onSubmit={(event) => {
                  event.preventDefault();
                  applyCustomRadius(customRadiusKm);
                }}
              >
                <label className="min-w-0">
                  <span className="sr-only">Custom radius in kilometers</span>
                  <input
                    className="h-9 w-full rounded-full border border-black/8 bg-white px-3 text-sm font-semibold text-[#333] outline-none transition placeholder:text-[#999] focus:border-black/20"
                    inputMode="numeric"
                    max={1000}
                    min={1}
                    onBlur={(event) => applyCustomRadius(event.target.value)}
                    onChange={(event) => setCustomRadiusKm(event.target.value)}
                    placeholder="Custom km"
                    type="number"
                    value={customRadiusKm}
                  />
                </label>
                <Button
                  className="w-full rounded-full sm:w-auto"
                  size="sm"
                  type="submit"
                  variant="outline"
                >
                  Apply
                </Button>
              </form>
              <p className="mt-2 text-xs font-medium text-[#777]">
                Custom radius supports up to 1000 km.
              </p>
            </div>
          </div>

          <div className="flex flex-wrap gap-2">
            <Button className="rounded-full" onClick={onLocate} type="button">
              <LocateFixed className="size-4" />
              Near me
            </Button>
            {hasNearbyCenter ? (
              <Button
                className="rounded-full"
                onClick={onClearNearby}
                type="button"
                variant="outline"
              >
                <X className="size-4" />
                Clear
              </Button>
            ) : null}
          </div>
        </aside>
      </div>
    </section>
  );
}

function ExploreHero({
  activeCollection,
  collections,
  featuredPosts,
  feedSort,
  isAuthenticated,
  onCollectionChange,
  onOpenPost,
  onSortChange,
  onToggleSavedBoard,
  savedBoardCount,
  showSavedBoard,
}: Readonly<{
  activeCollection: string;
  collections: string[];
  featuredPosts: Post[];
  feedSort: PostSort;
  isAuthenticated: boolean;
  onCollectionChange: (collection: string) => void;
  onOpenPost: (postId: number) => void;
  onSortChange: (sort: PostSort) => void;
  onToggleSavedBoard: () => void;
  savedBoardCount: number;
  showSavedBoard: boolean;
}>) {
  const visibleCollections = isAuthenticated
    ? collections
    : collections.filter((collection) => collection !== FOLLOWING_COLLECTION);

  return (
    <section className="mx-auto w-full max-w-370 px-3 pb-4 pt-5 sm:px-6 sm:pt-6 lg:px-8">
      <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_minmax(21rem,31rem)] lg:items-end">
        <div className="max-w-3xl">
          <div className="mb-3 inline-flex items-center gap-2 rounded-full border border-black/6 bg-white px-3 py-1.5 text-xs font-semibold text-[#555] shadow-[0_12px_30px_-24px_rgb(0_0_0/0.35)]">
            <Sparkles className="size-3.5 text-[#ff385c]" />
            Curated for travelers
          </div>
          <h1 className="max-w-2xl text-3xl font-semibold leading-[1.12] tracking-normal text-[#111] sm:text-5xl lg:text-6xl">
            <span className="sr-only">
              Discover beautiful destinations and places worth visiting.
            </span>
            <span aria-hidden="true" className="inline-flex flex-wrap gap-x-2 sm:gap-x-3">
              {exploreHeroHeadlineWords.map((word, index) => (
                <span
                  className="inline-block animate-[hero-word-rise_0.72s_cubic-bezier(0.22,1,0.36,1)_both]"
                  key={`${word}-${index}`}
                  style={{ animationDelay: `${index * 78}ms` }}
                >
                  {word}
                </span>
              ))}
            </span>
          </h1>
          <p className="mt-4 max-w-xl text-base leading-7 text-[#5f5f5f] sm:text-lg">
            Falzo brings together photos, locations, and real travel stories so
            you can quickly choose the right place for a vacation, short escape,
            or longer journey.
          </p>
        </div>

        <HeroDestinationPreview
          onOpenPost={onOpenPost}
          posts={featuredPosts}
        />
      </div>

      <div className="mt-5 grid grid-cols-2 gap-2 sm:flex sm:flex-wrap sm:items-center sm:justify-between lg:mt-6">
        <div className="contents sm:flex sm:flex-wrap sm:items-center sm:gap-2">
          {[
            { icon: <Clock3 className="size-4" />, label: "Newest", value: "newest" as const },
            { icon: <Flame className="size-4" />, label: "Popular", value: "popular" as const },
            { icon: <Flame className="size-4" />, label: "Trending", value: "trending" as const },
            { icon: <LocateFixed className="size-4" />, label: "Nearby", value: "nearby" as const },
          ].map((item) => (
            <Button
              className={cn(
                "rounded-full border-black/8 px-3",
                feedSort === item.value
                  ? "bg-[#111] text-white hover:bg-[#222]"
                  : "bg-white",
              )}
              key={item.value}
              onClick={() => onSortChange(item.value)}
              type="button"
              variant={feedSort === item.value ? "default" : "outline"}
            >
              {item.icon}
              {item.label}
            </Button>
          ))}
        </div>
        {isAuthenticated ? (
          <Button
            aria-pressed={showSavedBoard}
            className={cn(
              "col-span-2 rounded-full shadow-[0_18px_38px_-24px_rgb(255_56_92/0.8)] sm:col-span-1",
              showSavedBoard
                ? "bg-[#111] text-white hover:bg-[#222]"
                : "bg-[#ff385c] text-white hover:bg-[#e93152]",
            )}
            onClick={onToggleSavedBoard}
            type="button"
          >
            <Bookmark className="size-4" />
            Trip board
            {savedBoardCount > 0 ? (
              <span className="rounded-full bg-white/18 px-2 py-0.5 text-xs">
                {savedBoardCount}
              </span>
            ) : null}
          </Button>
        ) : null}
      </div>

      <div className="mt-5 flex gap-2 overflow-x-auto pb-2 scrollbar-none sm:mt-6 [&::-webkit-scrollbar]:hidden">
        {visibleCollections.map((collection) => (
          <button
            className={cn(
              "shrink-0 rounded-full border px-4 py-2 text-sm font-semibold transition",
              activeCollection === collection
                ? "border-[#111] bg-[#111] text-white shadow-[0_16px_32px_-24px_rgb(0_0_0/0.75)]"
                : "border-black/7 bg-white text-[#444] hover:border-black/15 hover:bg-[#fbfbfa]",
            )}
            key={collection}
            onClick={() => onCollectionChange(collection)}
            type="button"
          >
            {collection}
          </button>
        ))}
      </div>
    </section>
  );
}

function HeroDestinationPreview({
  onOpenPost,
  posts,
}: Readonly<{
  onOpenPost: (postId: number) => void;
  posts: Post[];
}>) {
  const [spotlight, ...supportingPosts] = posts;

  if (!spotlight) {
    return (
      <div className="min-h-80 overflow-hidden rounded-[2rem] border border-black/6 bg-[#111] p-4 text-white shadow-[0_24px_70px_-52px_rgb(0_0_0/0.85)]">
        <div className="flex h-full min-h-72 flex-col justify-between rounded-[1.5rem] border border-white/12 bg-[linear-gradient(135deg,#1a1a1a_0%,#243c34_48%,#17445b_100%)] p-5">
          <div className="flex size-11 items-center justify-center rounded-full bg-white text-[#111]">
            <Camera className="size-5" />
          </div>
          <div>
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-white/60">
              Destination preview
            </p>
            <h2 className="mt-2 text-2xl font-semibold leading-tight tracking-normal">
              Travel photos will appear here as the feed loads.
            </h2>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="grid min-h-80 gap-3 rounded-[2rem] border border-black/6 bg-white p-3 shadow-[0_26px_70px_-50px_rgb(0_0_0/0.72)] sm:grid-cols-[minmax(0,1fr)_8.5rem]">
      <button
        className="group relative min-h-72 overflow-hidden rounded-[1.5rem] bg-[#e9eef3] text-left"
        onClick={() => onOpenPost(spotlight.id)}
        type="button"
      >
        <img
          alt={spotlight.caption || spotlight.location_name || "Featured destination"}
          className="h-full w-full object-cover transition duration-700 group-hover:scale-[1.045]"
          loading="eager"
          src={spotlight.image_url}
        />
        <div className="absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0.02)_0%,rgb(0_0_0/0.18)_48%,rgb(0_0_0/0.78)_100%)]" />
        <div className="absolute left-4 top-4 inline-flex items-center gap-2 rounded-full bg-white/86 px-3 py-1.5 text-xs font-bold text-[#111] shadow-sm backdrop-blur-xl">
          <Sparkles className="size-3.5 text-[#ff385c]" />
          Spotlight
        </div>
        <div className="absolute inset-x-4 bottom-4 text-white">
          <p className="line-clamp-1 text-xs font-bold uppercase tracking-[0.16em] text-white/70">
            {spotlight.location_name || spotlight.category_name || "Destination"}
          </p>
          <h2 className="mt-1 line-clamp-2 text-2xl font-semibold leading-tight tracking-normal">
            {spotlight.caption || "Open this destination story"}
          </h2>
          <span className="mt-3 inline-flex items-center gap-2 rounded-full bg-white px-3 py-1.5 text-xs font-bold text-[#111]">
            <Camera className="size-3.5" />
            Open travel lens
          </span>
        </div>
      </button>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-1">
        {(supportingPosts.length > 0 ? supportingPosts : [spotlight]).map(
          (post, index) => (
            <button
              className="group relative min-h-34 overflow-hidden rounded-[1.25rem] bg-[#e9eef3] text-left"
              key={`${post.id}-${index}`}
              onClick={() => onOpenPost(post.id)}
              type="button"
            >
              <img
                alt={post.caption || post.location_name || "Destination"}
                className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.06]"
                loading="lazy"
                src={post.image_url}
              />
              <div className="absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0.02)_0%,rgb(0_0_0/0.16)_42%,rgb(0_0_0/0.68)_100%)]" />
              <p className="absolute inset-x-3 bottom-3 line-clamp-2 text-sm font-semibold leading-tight text-white">
                {post.location_name || post.caption || "Explore place"}
              </p>
            </button>
          ),
        )}
      </div>
    </div>
  );
}

function FirstLoadTravelSuggestions({
  onDismiss,
  onSelectSuggestion,
}: Readonly<{
  onDismiss: () => void;
  onSelectSuggestion: (query: string, sort: PostSort) => void;
}>) {
  return (
    <section className="mx-auto w-full max-w-370 px-3 pb-5 sm:px-6 lg:px-8">
      <div className="relative rounded-2xl border border-black/6 bg-white px-3 py-3 shadow-[0_14px_38px_-32px_rgb(0_0_0/0.58)] sm:px-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
          <div className="flex min-w-0 items-center gap-2 pr-8 sm:shrink-0">
            <span className="flex size-8 items-center justify-center rounded-full bg-[#f2f7fd] text-[#315f8f]">
              <Sparkles className="size-4" />
            </span>
            <div className="min-w-0">
              <p className="text-sm font-semibold text-[#111]">Quick start</p>
              <p className="text-xs text-[#777]">Choose one suggestion.</p>
            </div>
          </div>

          <div className="flex min-w-0 flex-1 gap-2 overflow-x-auto pb-1 scrollbar-none sm:pb-0 [&::-webkit-scrollbar]:hidden">
            {firstLoadTravelSuggestions.map((item) => (
              <button
                className="shrink-0 rounded-full border border-black/8 bg-[#f8f8f6] px-3.5 py-2 text-sm font-semibold text-[#333] transition hover:border-black/16 hover:bg-white"
                key={item.title}
                onClick={() => onSelectSuggestion(item.query, item.sort)}
                type="button"
              >
                {item.title}
              </button>
            ))}
          </div>
        </div>

        <button
          aria-label="Hide travel suggestions"
          className="absolute right-3 top-3 flex size-8 items-center justify-center rounded-full text-[#666] transition hover:bg-black/5 hover:text-[#111]"
          onClick={onDismiss}
          type="button"
        >
          <X className="size-4" />
        </button>
      </div>
    </section>
  );
}

function ExploreCategoryBar({
  categories,
  onSelectCategory,
}: Readonly<{
  categories: Category[];
  onSelectCategory: (category: Category) => void;
}>) {
  if (categories.length === 0) {
    return null;
  }

  return (
    <section className="mx-auto w-full max-w-370 px-3 pb-5 sm:px-6 lg:px-8">
      <div className="flex flex-col gap-3 border-black/6 border-y bg-white/74 py-4 sm:flex-row sm:items-center">
        <div className="flex min-w-0 shrink-0 items-center gap-2 text-[#333] sm:min-w-42">
          <span className="flex size-8 items-center justify-center rounded-full bg-[#111] text-white">
            <Tags className="size-4" />
          </span>
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
              Themes
            </p>
            <h2 className="text-base font-semibold tracking-normal text-[#111]">
              All
            </h2>
          </div>
        </div>

        <div className="flex gap-2 overflow-x-auto pb-1 scrollbar-none sm:pb-0 [&::-webkit-scrollbar]:hidden">
          {categories.map((category) => (
            <button
              className="shrink-0 rounded-full border border-black/7 bg-[#f7f7f5] px-3.5 py-2 text-sm font-semibold text-[#444] transition hover:border-black/15 hover:bg-white"
              key={category.id}
              onClick={() => onSelectCategory(category)}
              type="button"
            >
              {category.name}
            </button>
          ))}
        </div>
      </div>
    </section>
  );
}

function LoadMorePosts({
  hasNextPage,
  isLoading,
  refNode,
  totalPosts,
}: Readonly<{
  hasNextPage: boolean;
  isLoading: boolean;
  refNode: React.RefObject<HTMLDivElement | null>;
  totalPosts: number;
}>) {
  return (
    <div
      className="flex min-h-20 items-center justify-center py-4"
      ref={refNode}
    >
      {isLoading ? (
        <div className="inline-flex items-center gap-2 rounded-full border border-black/8 bg-white px-4 py-2 text-sm font-semibold text-[#555] shadow-[0_12px_30px_-24px_rgb(0_0_0/0.35)]">
          <span className="size-2 animate-pulse rounded-full bg-[#ff385c]" />
          Loading more destinations
        </div>
      ) : null}
      {!hasNextPage && totalPosts > 0 ? (
        <p className="text-sm font-semibold text-[#777]">
          You have reached the end.
        </p>
      ) : null}
    </div>
  );
}
