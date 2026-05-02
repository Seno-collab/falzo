"use client";

import {
  Bell,
  Bookmark,
  Camera,
  ChevronDown,
  Heart,
  Home,
  Map,
  Menu,
  MessageCircle,
  Plus,
  Search,
  Send,
  SlidersHorizontal,
  Sparkles,
  UserRound,
} from "lucide-react";
import { motion } from "motion/react";
import { useInfiniteQuery, useMutation, useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { ScenicImage } from "@/components/scenic-image";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { getApiErrorMessage, hasAuthSession } from "@/features/auth/api";
import {
  createPostCommentApi,
  getPostDetailApi,
  getPostCommentsApi,
  getPostsApi,
  likePostApi,
  savePostApi,
} from "@/features/posts/api";
import type { Post, PostComment } from "@/features/posts/types";
import {
  exploreCollections,
  explorePins,
  type ExploreCollection,
} from "@/features/scenic/data";
import { ROUTES } from "@/lib/routes";
import { cn } from "@/lib/utils";

type ActiveCollection = ExploreCollection | "Community";
const postsPageSize = 24;

export function ExploreScreen() {
  const router = useRouter();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [activeCollection, setActiveCollection] =
    useState<ActiveCollection>("All");
  const [likedPosts, setLikedPosts] = useState<Set<number>>(new Set());
  const [savedPosts, setSavedPosts] = useState<Set<number>>(new Set());
  const [savedPins, setSavedPins] = useState<Set<string>>(new Set());
  const [openComments, setOpenComments] = useState<Set<number>>(new Set());
  const [loadingComments, setLoadingComments] = useState<Set<number>>(new Set());
  const [selectedPostId, setSelectedPostId] = useState<number | null>(null);
  const [commentsByPost, setCommentsByPost] = useState<
    Record<number, PostComment[]>
  >({});
  const [commentInputs, setCommentInputs] = useState<Record<number, string>>({});
  const loadMoreRef = useRef<HTMLDivElement | null>(null);

  const postsQuery = useInfiniteQuery({
    queryKey: ["posts", "explore"],
    queryFn: ({ pageParam }) =>
      getPostsApi({ page: pageParam, limit: postsPageSize }),
    getNextPageParam: (lastPage, _pages, lastPageParam) =>
      lastPage.length < postsPageSize ? undefined : lastPageParam + 1,
    initialPageParam: 1,
  });

  const postDetailQuery = useQuery({
    enabled: selectedPostId !== null,
    queryKey: ["posts", "detail", selectedPostId],
    queryFn: () => getPostDetailApi(selectedPostId ?? 0),
  });

  useEffect(() => {
    document.title = "Falzo Explore | Visual Inspiration";
    setIsAuthenticated(hasAuthSession());
  }, []);

  const likeMutation = useMutation({
    mutationFn: likePostApi,
    onError: (error) => {
      toast.error(getApiErrorMessage(error));
    },
    onSuccess: (_data, postId) => {
      setLikedPosts((current) => new Set(current).add(postId));
    },
  });

  const saveMutation = useMutation({
    mutationFn: savePostApi,
    onError: (error) => {
      toast.error(getApiErrorMessage(error));
    },
    onSuccess: (_data, postId) => {
      setSavedPosts((current) => new Set(current).add(postId));
    },
  });

  const commentMutation = useMutation({
    mutationFn: createPostCommentApi,
    onError: (error) => {
      toast.error(getApiErrorMessage(error));
    },
    onSuccess: (comment, variables) => {
      setCommentsByPost((current) => ({
        ...current,
        [variables.postId]: [...(current[variables.postId] ?? []), comment],
      }));
      setCommentInputs((current) => ({ ...current, [variables.postId]: "" }));
      toast.success("Comment posted.");
    },
  });

  const collections = useMemo<ActiveCollection[]>(
    () => [
      "All",
      "Community",
      ...exploreCollections.filter((item) => item !== "All"),
    ],
    [],
  );

  const visiblePins = useMemo(() => {
    if (activeCollection === "All") {
      return explorePins;
    }

    if (activeCollection === "Community") {
      return [];
    }

    return explorePins.filter((pin) => pin.collection === activeCollection);
  }, [activeCollection]);

  const visiblePosts = useMemo(() => {
    if (activeCollection !== "All" && activeCollection !== "Community") {
      return [];
    }

    return postsQuery.data?.pages.flat() ?? [];
  }, [activeCollection, postsQuery.data]);

  const selectedPost = useMemo<Post | null>(() => {
    if (selectedPostId === null) {
      return null;
    }

    return (
      postDetailQuery.data ??
      visiblePosts.find((post) => post.id === selectedPostId) ??
      null
    );
  }, [postDetailQuery.data, selectedPostId, visiblePosts]);

  const shouldLoadMorePosts =
    (activeCollection === "All" || activeCollection === "Community") &&
    postsQuery.hasNextPage &&
    !postsQuery.isFetchingNextPage;

  useEffect(() => {
    const node = loadMoreRef.current;
    if (!node || !shouldLoadMorePosts) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          void postsQuery.fetchNextPage();
        }
      },
      { rootMargin: "700px 0px" },
    );

    observer.observe(node);

    return () => {
      observer.disconnect();
    };
  }, [postsQuery.fetchNextPage, shouldLoadMorePosts]);

  function toggleSaved(id: string) {
    setSavedPins((current) => {
      const next = new Set(current);

      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }

      return next;
    });
  }

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

  function handleLikePost(postId: number) {
    if (!requireAuth() || likedPosts.has(postId) || likeMutation.isPending) {
      return;
    }

    likeMutation.mutate(postId);
  }

  function handleSavePost(postId: number) {
    if (!requireAuth() || savedPosts.has(postId) || saveMutation.isPending) {
      return;
    }

    saveMutation.mutate(postId);
  }

  async function loadComments(postId: number) {
    if (commentsByPost[postId] || loadingComments.has(postId)) {
      return;
    }

    setLoadingComments((current) => new Set(current).add(postId));

    try {
      const comments = await getPostCommentsApi(postId);
      setCommentsByPost((current) => ({ ...current, [postId]: comments }));
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
    setOpenComments((current) => {
      const next = new Set(current);
      if (next.has(postId)) {
        next.delete(postId);
      } else {
        next.add(postId);
      }
      return next;
    });

    if (willOpen) {
      void loadComments(postId);
    }
  }

  function openPostDetail(postId: number) {
    setSelectedPostId(postId);
    setOpenComments((current) => new Set(current).add(postId));
    void loadComments(postId);
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

    commentMutation.mutate({ postId, content });
  }

  const selectedPostComments =
    selectedPostId === null ? [] : (commentsByPost[selectedPostId] ?? []);
  const isSelectedPostLoadingComments =
    selectedPostId !== null && loadingComments.has(selectedPostId);
  const isSelectedPostLiked =
    selectedPostId !== null && likedPosts.has(selectedPostId);
  const isSelectedPostSaved =
    selectedPostId !== null && savedPosts.has(selectedPostId);

  return (
    <main className="min-h-screen bg-[#f7f7f5] text-[#1f1f1f]">
      <header className="sticky top-0 z-40 border-b border-black/6 bg-[#f7f7f5]/86 backdrop-blur-2xl">
        <div className="mx-auto flex w-full max-w-370 items-center gap-2 px-3 py-3 sm:px-5 lg:px-8">
          <Link
            aria-label="Home"
            className="inline-flex size-10 items-center justify-center rounded-full bg-[#111] text-white shadow-[0_14px_30px_-20px_rgb(0_0_0/0.72)] transition hover:scale-[1.03]"
            href="/"
          >
            <Camera className="size-4" />
          </Link>

          <nav className="hidden items-center gap-1 md:flex">
            <Button asChild className="rounded-full" size="sm" variant="ghost">
              <Link href="/">
                <Home className="size-4" />
                Home
              </Link>
            </Button>
            <Button
              className="rounded-full bg-[#111] text-white hover:bg-[#222]"
              size="sm"
            >
              Explore
            </Button>
            <Button asChild className="rounded-full" size="sm" variant="ghost">
              <Link href={ROUTES.locations}>
                <Map className="size-4" />
                Locations
              </Link>
            </Button>
          </nav>

          <div className="relative ml-1 flex-1">
            <Search className="-translate-y-1/2 pointer-events-none absolute left-4 top-1/2 size-4 text-[#777]" />
            <input
              className="h-11 w-full rounded-full border border-black/6 bg-white px-11 text-sm text-[#1f1f1f] shadow-[0_12px_32px_-28px_rgb(0_0_0/0.45)] outline-none transition placeholder:text-[#8a8a8a] focus:border-black/10 focus:bg-white focus:shadow-[0_18px_40px_-30px_rgb(0_0_0/0.58)]"
              placeholder="Search places, rooms, tables, textures"
              type="search"
            />
            <Button
              aria-label="Search filters"
              className="-translate-y-1/2 absolute right-1.5 top-1/2 rounded-full"
              size="icon-sm"
              type="button"
              variant="ghost"
            >
              <SlidersHorizontal className="size-4" />
            </Button>
          </div>

          <div className="hidden items-center gap-1 sm:flex">
            <Button
              aria-label="Create"
              asChild
              className="rounded-full"
              size="icon-sm"
              variant="ghost"
            >
              <Link href={ROUTES.upload}>
                <Plus className="size-4" />
              </Link>
            </Button>
            <Button
              aria-label="Notifications"
              className="rounded-full"
              size="icon-sm"
              type="button"
              variant="ghost"
            >
              <Bell className="size-4" />
            </Button>
            <Button
              aria-label="Profile"
              className="rounded-full"
              size="icon-sm"
              type="button"
              variant="outline"
            >
              <UserRound className="size-4" />
            </Button>
          </div>

          <Button
            aria-label="Menu"
            className="rounded-full sm:hidden"
            size="icon-sm"
            type="button"
            variant="ghost"
          >
            <Menu className="size-4" />
          </Button>
        </div>
      </header>

      <section className="mx-auto w-full max-w-[1480px] px-4 pb-4 pt-6 sm:px-6 lg:px-8">
        <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
          <div className="max-w-3xl">
            <div className="mb-3 inline-flex items-center gap-2 rounded-full border border-black/[0.06] bg-white px-3 py-1.5 text-xs font-semibold text-[#555] shadow-[0_12px_30px_-24px_rgb(0_0_0/0.35)]">
              <Sparkles className="size-3.5 text-[#ff385c]" />
              Curated today
            </div>
            <h1 className="max-w-2xl text-4xl font-semibold tracking-normal text-[#111] sm:text-5xl lg:text-6xl">
              Fresh visual ideas for beautiful stays and memorable travel.
            </h1>
          </div>

          <div className="flex items-center gap-2 lg:justify-end">
            <Button
              className="rounded-full border-black/[0.08] bg-white"
              type="button"
              variant="outline"
            >
              Trending
              <ChevronDown className="size-4" />
            </Button>
            <Button
              className="rounded-full bg-[#ff385c] text-white shadow-[0_18px_38px_-24px_rgb(255_56_92/0.8)] hover:bg-[#e93152]"
              type="button"
            >
              <Bookmark className="size-4" />
              Board
            </Button>
          </div>
        </div>

        <div className="mt-6 flex gap-2 overflow-x-auto pb-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          {collections.map((collection) => (
            <button
              className={cn(
                "shrink-0 rounded-full border px-4 py-2 text-sm font-semibold transition",
                activeCollection === collection
                  ? "border-[#111] bg-[#111] text-white shadow-[0_16px_32px_-24px_rgb(0_0_0/0.75)]"
                  : "border-black/[0.07] bg-white text-[#444] hover:border-black/15 hover:bg-[#fbfbfa]",
              )}
              key={collection}
              onClick={() => setActiveCollection(collection)}
              type="button"
            >
              {collection}
            </button>
          ))}
        </div>
      </section>

      <section className="mx-auto w-full max-w-370 px-4 pb-14 sm:px-6 lg:px-8">
        <div className="columns-1 gap-4 sm:columns-2 lg:columns-3 2xl:columns-4">
          {visiblePosts.map((post, index) => {
            const id = `post-${post.id}`;
            const isLiked = likedPosts.has(post.id);
            const isSaved = savedPosts.has(post.id);
            const commentsOpen = openComments.has(post.id);
            const comments = commentsByPost[post.id] ?? [];
            const isLoadingComments = loadingComments.has(post.id);

            return (
              <motion.article
                className="group mb-4 break-inside-avoid overflow-hidden rounded-[28px] border border-black/[0.05] bg-white shadow-[0_18px_48px_-38px_rgb(0_0_0/0.6)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_26px_70px_-42px_rgb(0_0_0/0.72)]"
                initial={{ opacity: 0, y: 18 }}
                key={id}
                transition={{
                  duration: 0.34,
                  delay: Math.min(index * 0.035, 0.22),
                  ease: "easeOut",
                }}
                viewport={{ amount: 0.12, once: true }}
                whileInView={{ opacity: 1, y: 0 }}
              >
                <div
                  className="relative h-96 cursor-zoom-in overflow-hidden bg-[#e9eef3]"
                  onClick={() => openPostDetail(post.id)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      openPostDetail(post.id);
                    }
                  }}
                  role="button"
                  tabIndex={0}
                >
                  <img
                    alt={post.caption || post.location_name || "Uploaded post"}
                    className="h-full w-full object-cover transition duration-500 ease-out group-hover:scale-[1.035]"
                    loading={index < 2 ? "eager" : "lazy"}
                    src={post.image_url}
                  />
                  <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0.02)_0%,rgb(0_0_0/0.02)_48%,rgb(0_0_0/0.44)_100%)] opacity-80 transition group-hover:opacity-100" />
                  <div className="absolute left-3 right-3 top-3 flex items-start justify-between gap-2 opacity-0 transition duration-200 group-hover:opacity-100">
                    <span className="rounded-full bg-white/86 px-3 py-1 text-xs font-semibold text-[#222] shadow-sm backdrop-blur-xl">
                      Community
                    </span>
                    <Button
                      aria-label={isLiked ? "Liked" : "Like"}
                      className={cn(
                        "rounded-full shadow-sm backdrop-blur-xl",
                        isLiked
                          ? "bg-[#ff385c] text-white hover:bg-[#e93152]"
                          : "bg-white/86 text-[#222] hover:bg-white",
                      )}
                      onClick={(event) => {
                        event.stopPropagation();
                        handleLikePost(post.id);
                      }}
                      size="icon-sm"
                      type="button"
                      variant="ghost"
                    >
                      <Heart
                        className={cn("size-4", isLiked ? "fill-current" : "")}
                      />
                    </Button>
                  </div>
                  <div className="absolute inset-x-4 bottom-4 text-white">
                    <p className="text-xs font-semibold uppercase tracking-[0.16em] text-white/76">
                      {post.location_name || "Uploaded"}
                    </p>
                    <h2 className="mt-1 text-2xl font-semibold tracking-normal">
                      {post.caption || "Community post"}
                    </h2>
                  </div>
                </div>

                <div className="space-y-3 p-4">
                  <div className="flex items-center justify-between gap-3">
                    <div className="min-w-0">
                      <p className="truncate text-sm font-semibold text-[#202020]">
                        User #{post.user_id}
                      </p>
                      <p className="mt-0.5 text-xs font-medium text-[#777]">
                        {new Date(post.created_at).toLocaleDateString()}
                      </p>
                    </div>
                    <div className="flex shrink-0 items-center gap-1">
                      <Button
                        aria-label={isLiked ? "Liked" : "Like post"}
                        className={cn(
                          "rounded-full",
                          isLiked
                            ? "border-[#ffb3c1] bg-[#fff1f4] text-[#cf2142]"
                            : "",
                        )}
                        onClick={() => handleLikePost(post.id)}
                        size="icon-sm"
                        type="button"
                        variant="outline"
                      >
                        <Heart
                          className={cn(
                            "size-4",
                            isLiked ? "fill-current" : "",
                          )}
                        />
                      </Button>
                      <Button
                        aria-label="View comments"
                        className={cn(
                          "rounded-full",
                          commentsOpen
                            ? "border-[#b9d6f2] bg-[#f0f7ff] text-[#2f6fb8]"
                            : "",
                        )}
                        onClick={() => toggleComments(post.id)}
                        size="icon-sm"
                        type="button"
                        variant="outline"
                      >
                        <MessageCircle className="size-4" />
                      </Button>
                      <Button
                        aria-label={isSaved ? "Saved" : "Save post"}
                        className={cn(
                          "rounded-full",
                          isSaved
                            ? "border-[#c8ddf1] bg-[#f2f7fd] text-[#2f6fb8]"
                            : "",
                        )}
                        onClick={() => handleSavePost(post.id)}
                        size="icon-sm"
                        type="button"
                        variant="outline"
                      >
                        <Bookmark
                          className={cn(
                            "size-4",
                            isSaved ? "fill-current" : "",
                          )}
                        />
                      </Button>
                    </div>
                  </div>

                  {commentsOpen ? (
                    <div className="rounded-2xl border border-black/[0.06] bg-[#f8f8f7] p-3">
                      <div className="space-y-2">
                        {isLoadingComments ? (
                          <p className="text-xs font-semibold text-[#555]">
                            Loading comments
                          </p>
                        ) : null}
                        {!isLoadingComments && comments.length === 0 ? (
                          <p className="text-xs font-semibold text-[#555]">
                            No comments yet.
                          </p>
                        ) : null}
                        {comments.map((comment) => (
                          <div
                            className="rounded-xl bg-white px-3 py-2 text-sm text-[#333]"
                            key={comment.id}
                          >
                            <p className="text-xs font-semibold text-[#777]">
                              User #{comment.user_id}
                            </p>
                            <p className="mt-1 leading-5">{comment.content}</p>
                          </div>
                        ))}
                      </div>
                      <div className="mt-3 flex items-center gap-2">
                        <input
                          className="h-9 min-w-0 flex-1 rounded-full border border-black/[0.08] bg-white px-3 text-sm outline-none placeholder:text-[#999] focus:border-[#111]/20"
                          disabled={!isAuthenticated || commentMutation.isPending}
                          onChange={(event) =>
                            setCommentInputs((current) => ({
                              ...current,
                              [post.id]: event.target.value,
                            }))
                          }
                          placeholder={
                            isAuthenticated ? "Write a comment" : "Login to comment"
                          }
                          value={commentInputs[post.id] ?? ""}
                        />
                        <Button
                          aria-label={
                            isAuthenticated ? "Submit comment" : "Login to comment"
                          }
                          className="rounded-full"
                          disabled={commentMutation.isPending}
                          onClick={() => submitComment(post.id)}
                          size="icon-sm"
                          type="button"
                          variant="outline"
                        >
                          <Send className="size-4" />
                        </Button>
                      </div>
                    </div>
                  ) : null}
                </div>
              </motion.article>
            );
          })}

          {visiblePins.map((pin, index) => {
            const isSaved = savedPins.has(pin.id);

            return (
              <motion.article
                className="group mb-4 break-inside-avoid overflow-hidden rounded-[28px] border border-black/[0.05] bg-white shadow-[0_18px_48px_-38px_rgb(0_0_0/0.6)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_26px_70px_-42px_rgb(0_0_0/0.72)]"
                initial={{ opacity: 0, y: 18 }}
                key={pin.id}
                transition={{
                  duration: 0.34,
                  delay: Math.min(index * 0.035, 0.22),
                  ease: "easeOut",
                }}
                viewport={{ amount: 0.12, once: true }}
                whileInView={{ opacity: 1, y: 0 }}
              >
                <div
                  className={cn(
                    "relative overflow-hidden",
                    pin.height,
                    pin.gradient,
                  )}
                >
                  <ScenicImage
                    alt={pin.title}
                    className="h-full w-full object-cover transition duration-500 ease-out group-hover:scale-[1.035]"
                    id={pin.imageId}
                    sizes="(max-width: 640px) 92vw, (max-width: 1024px) 46vw, (max-width: 1536px) 31vw, 23vw"
                  />
                  <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0.02)_0%,rgb(0_0_0/0.02)_48%,rgb(0_0_0/0.44)_100%)] opacity-80 transition group-hover:opacity-100" />
                  <div className="absolute left-3 right-3 top-3 flex items-start justify-between gap-2 opacity-0 transition duration-200 group-hover:opacity-100">
                    <span className="rounded-full bg-white/86 px-3 py-1 text-xs font-semibold text-[#222] shadow-sm backdrop-blur-xl">
                      {pin.collection}
                    </span>
                    <Button
                      aria-label={isSaved ? "Remove save" : "Save"}
                      className={cn(
                        "rounded-full shadow-sm backdrop-blur-xl",
                        isSaved
                          ? "bg-[#ff385c] text-white hover:bg-[#e93152]"
                          : "bg-white/86 text-[#222] hover:bg-white",
                      )}
                      onClick={() => toggleSaved(pin.id)}
                      size="icon-sm"
                      type="button"
                      variant="ghost"
                    >
                      <Heart
                        className={cn("size-4", isSaved ? "fill-current" : "")}
                      />
                    </Button>
                  </div>
                  <div className="absolute inset-x-4 bottom-4 text-white">
                    <p className="text-xs font-semibold uppercase tracking-[0.16em] text-white/76">
                      {pin.city}
                    </p>
                    <h2 className="mt-1 text-2xl font-semibold tracking-normal">
                      {pin.title}
                    </h2>
                  </div>
                </div>

                <div className="flex items-center justify-between gap-3 p-4">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-semibold text-[#202020]">
                      {pin.author}
                    </p>
                    <p className="mt-0.5 text-xs font-medium text-[#777]">
                      {pin.saves} saves
                    </p>
                  </div>
                  <Button
                    aria-label={isSaved ? "Saved" : "Save pin"}
                    className={cn(
                      "rounded-full",
                      isSaved
                        ? "border-[#ffb3c1] bg-[#fff1f4] text-[#cf2142]"
                        : "",
                    )}
                    onClick={() => toggleSaved(pin.id)}
                    size="icon-sm"
                    type="button"
                    variant="outline"
                  >
                    <Bookmark
                      className={cn("size-4", isSaved ? "fill-current" : "")}
                    />
                  </Button>
                </div>
              </motion.article>
            );
          })}
        </div>

        {activeCollection === "All" || activeCollection === "Community" ? (
          <div
            className="flex min-h-20 items-center justify-center py-4"
            ref={loadMoreRef}
          >
            {postsQuery.isFetchingNextPage ? (
              <div className="inline-flex items-center gap-2 rounded-full border border-black/[0.08] bg-white px-4 py-2 text-sm font-semibold text-[#555] shadow-[0_12px_30px_-24px_rgb(0_0_0/0.35)]">
                <span className="size-2 animate-pulse rounded-full bg-[#ff385c]" />
                Loading more posts
              </div>
            ) : null}
            {!postsQuery.hasNextPage && visiblePosts.length > 0 ? (
              <p className="text-sm font-semibold text-[#777]">
                You have reached the end.
              </p>
            ) : null}
          </div>
        ) : null}
      </section>

      <Dialog
        onOpenChange={(open) => {
          if (!open) {
            setSelectedPostId(null);
          }
        }}
        open={selectedPostId !== null}
      >
        <DialogContent className="max-h-[92vh] w-[min(96vw,76rem)] overflow-hidden border-white/16 bg-[#101010] p-0 text-white">
          <div className="grid max-h-[92vh] overflow-hidden lg:grid-cols-[minmax(0,1.25fr)_minmax(360px,0.72fr)]">
            <div className="min-h-[56vh] bg-black lg:min-h-[82vh]">
              {selectedPost ? (
                <img
                  alt={
                    selectedPost.caption ||
                    selectedPost.location_name ||
                    "Post detail"
                  }
                  className="h-full max-h-[82vh] w-full object-contain"
                  src={selectedPost.image_url}
                />
              ) : (
                <div className="flex h-full min-h-[56vh] items-center justify-center text-sm font-semibold text-white/68">
                  Loading post
                </div>
              )}
            </div>

            <aside className="flex max-h-[92vh] min-h-0 flex-col bg-[#f7f7f5] text-[#1f1f1f]">
              <div className="border-b border-black/[0.06] p-5">
                <DialogHeader>
                  <DialogTitle className="text-xl leading-7 text-[#111]">
                    {selectedPost?.caption || "Community post"}
                  </DialogTitle>
                  <DialogDescription className="text-sm text-[#777]">
                    {selectedPost?.location_name || "Uploaded"} - User #
                    {selectedPost?.user_id ?? ""}
                  </DialogDescription>
                </DialogHeader>

                <div className="mt-4 flex items-center gap-2">
                  <Button
                    aria-label={
                      isSelectedPostLiked ? "Liked post" : "Like post"
                    }
                    className={cn(
                      "rounded-full",
                      isSelectedPostLiked
                        ? "border-[#ffb3c1] bg-[#fff1f4] text-[#cf2142]"
                        : "",
                    )}
                    disabled={selectedPostId === null || likeMutation.isPending}
                    onClick={() => {
                      if (selectedPostId !== null) {
                        handleLikePost(selectedPostId);
                      }
                    }}
                    size="icon-sm"
                    type="button"
                    variant="outline"
                  >
                    <Heart
                      className={cn(
                        "size-4",
                        isSelectedPostLiked ? "fill-current" : "",
                      )}
                    />
                  </Button>
                  <Button
                    aria-label={
                      isSelectedPostSaved ? "Saved post" : "Save post"
                    }
                    className={cn(
                      "rounded-full",
                      isSelectedPostSaved
                        ? "border-[#c8ddf1] bg-[#f2f7fd] text-[#2f6fb8]"
                        : "",
                    )}
                    disabled={selectedPostId === null || saveMutation.isPending}
                    onClick={() => {
                      if (selectedPostId !== null) {
                        handleSavePost(selectedPostId);
                      }
                    }}
                    size="icon-sm"
                    type="button"
                    variant="outline"
                  >
                    <Bookmark
                      className={cn(
                        "size-4",
                        isSelectedPostSaved ? "fill-current" : "",
                      )}
                    />
                  </Button>
                  <Button
                    aria-label="Focus comment input"
                    className="rounded-full"
                    disabled={selectedPostId === null}
                    onClick={() => {
                      if (selectedPostId !== null) {
                        setOpenComments((current) =>
                          new Set(current).add(selectedPostId),
                        );
                        void loadComments(selectedPostId);
                      }
                    }}
                    size="icon-sm"
                    type="button"
                    variant="outline"
                  >
                    <MessageCircle className="size-4" />
                  </Button>
                </div>
              </div>

              <div className="min-h-0 flex-1 overflow-y-auto p-5">
                <div className="space-y-3">
                  {isSelectedPostLoadingComments ? (
                    <p className="text-sm font-semibold text-[#555]">
                      Loading comments
                    </p>
                  ) : null}
                  {!isSelectedPostLoadingComments &&
                  selectedPostComments.length === 0 ? (
                    <div className="rounded-2xl border border-black/[0.06] bg-white px-4 py-5 text-sm font-semibold text-[#555]">
                      No comments yet.
                    </div>
                  ) : null}
                  {selectedPostComments.map((comment) => (
                    <div
                      className="rounded-2xl border border-black/[0.05] bg-white px-4 py-3 text-sm text-[#333]"
                      key={comment.id}
                    >
                      <div className="flex items-center justify-between gap-3">
                        <p className="text-xs font-semibold text-[#777]">
                          User #{comment.user_id}
                        </p>
                        <p className="text-xs font-medium text-[#999]">
                          {new Date(comment.created_at).toLocaleDateString()}
                        </p>
                      </div>
                      <p className="mt-2 leading-6">{comment.content}</p>
                    </div>
                  ))}
                </div>
              </div>

              <div className="border-t border-black/[0.06] bg-white p-4">
                <div className="flex items-center gap-2">
                  <input
                    className="h-10 min-w-0 flex-1 rounded-full border border-black/[0.08] bg-white px-4 text-sm outline-none placeholder:text-[#999] focus:border-[#111]/20"
                    disabled={
                      !isAuthenticated ||
                      commentMutation.isPending ||
                      selectedPostId === null
                    }
                    onChange={(event) => {
                      if (selectedPostId === null) {
                        return;
                      }
                      setCommentInputs((current) => ({
                        ...current,
                        [selectedPostId]: event.target.value,
                      }));
                    }}
                    placeholder={
                      isAuthenticated ? "Write a comment" : "Login to comment"
                    }
                    value={
                      selectedPostId === null
                        ? ""
                        : (commentInputs[selectedPostId] ?? "")
                    }
                  />
                  <Button
                    aria-label={
                      isAuthenticated ? "Submit comment" : "Login to comment"
                    }
                    className="rounded-full"
                    disabled={commentMutation.isPending || selectedPostId === null}
                    onClick={() => {
                      if (selectedPostId !== null) {
                        submitComment(selectedPostId);
                      }
                    }}
                    size="icon-sm"
                    type="button"
                    variant="outline"
                  >
                    <Send className="size-4" />
                  </Button>
                </div>
              </div>
            </aside>
          </div>
        </DialogContent>
      </Dialog>
    </main>
  );
}
