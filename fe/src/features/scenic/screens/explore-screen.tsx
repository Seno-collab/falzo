"use client";

import {
  Bell,
  Bookmark,
  Camera,
  ChevronDown,
  MapIcon,
  Menu,
  Plus,
  Search,
  SlidersHorizontal,
  Sparkles,
  UserRound,
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
import { useEffect, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  getApiErrorMessage,
  getMeApi,
  hasAuthSession,
} from "@/features/auth/api";
import type { AuthUser } from "@/features/auth/types";
import { getAuthUserDisplayName } from "@/features/auth/user-display";
import { getCategoriesApi } from "@/features/categories/api";
import {
  createPostCommentApi,
  getPostCommentEventsUrl,
  getPostCommentsApi,
  getPostDetailApi,
  getPostEventsUrl,
  getPostsApi,
  likePostApi,
  parsePostCommentCreatedEvent,
  parsePostCreatedEvent,
  savePostApi,
  unlikePostApi,
  unsavePostApi,
  updatePostCommentApi,
} from "@/features/posts/api";
import type { Post, PostComment } from "@/features/posts/types";
import { ExplorePostCard } from "@/features/scenic/components/explore-cards";
import { PostDetailDialog } from "@/features/scenic/components/explore-detail-dialogs";
import {
  getExploreCollections,
  showsCommunityFeed,
  toggleSetValue,
} from "@/features/scenic/lib/explore-utils";
import { ROUTES } from "@/lib/routes";
import { cn } from "@/lib/utils";

type PostsInfiniteData = InfiniteData<Post[], number>;

const postsPageSize = 24;

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

export function ExploreScreen() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const loadMoreRef = useRef<HTMLDivElement | null>(null);
  const commentInputRefs = useRef<Record<number, HTMLInputElement | null>>({});

  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [activeCollection, setActiveCollection] = useState("All");
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

  const postsQuery = useInfiniteQuery({
    queryKey: ["posts", "explore"],
    queryFn: ({ pageParam }) =>
      getPostsApi({ page: pageParam, limit: postsPageSize }),
    getNextPageParam: (lastPage, _pages, lastPageParam) =>
      lastPage.length < postsPageSize ? undefined : lastPageParam + 1,
    initialPageParam: 1,
  });

  const categoriesQuery = useQuery({
    queryKey: ["categories"],
    queryFn: getCategoriesApi,
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

  useEffect(() => {
    document.title = "Falzo Explore | Visual Inspiration";
    setIsAuthenticated(hasAuthSession());
  }, []);

  useEffect(() => {
    const source = new EventSource(getPostEventsUrl());
    const handlePostCreated = (event: Event) => {
      const post = parsePostCreatedEvent(event as MessageEvent<string>);
      if (!post) {
        return;
      }

      queryClient.setQueryData<PostsInfiniteData>(
        ["posts", "explore"],
        (current) => {
          if (!current) {
            return {
              pageParams: [1],
              pages: [[post]],
            };
          }

          if (current.pages.some((page) => page.some((item) => item.id === post.id))) {
            return current;
          }

          const [firstPage = [], ...restPages] = current.pages;
          return {
            ...current,
            pages: [[post, ...firstPage], ...restPages],
          };
        },
      );
    };

    source.addEventListener("post.created", handlePostCreated);
    return () => {
      source.removeEventListener("post.created", handlePostCreated);
      source.close();
    };
  }, [queryClient]);

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

  const visiblePosts = useMemo(() => {
    if (!showsCommunityFeed(activeCollection)) {
      return [];
    }

    return postsQuery.data?.pages.flat() ?? [];
  }, [activeCollection, postsQuery.data]);

  const selectedPost = useMemo(() => {
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
    showsCommunityFeed(activeCollection) &&
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
    window.setTimeout(() => {
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
    window.setTimeout(() => {
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
        : visiblePosts.find((item) => item.id === postId) ?? null;
    if (!requireAuth() || !post || likeMutation.isPending) {
      return;
    }

    likeMutation.mutate({ isLiked: isPostLiked(post, likedPosts), postId });
  }

  function handleSavePost(postId: number) {
    const post =
      selectedPost?.id === postId
        ? selectedPost
        : visiblePosts.find((item) => item.id === postId) ?? null;
    if (!requireAuth() || !post || saveMutation.isPending) {
      return;
    }

    saveMutation.mutate({ isSaved: isPostSaved(post, savedPosts), postId });
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
  const profileName =
    profileQuery.data && !profileQuery.isFetching
      ? getAuthUserDisplayName(profileQuery.data, "") || null
      : null;
  const currentUserId = readAuthUserId(profileQuery.data);
  return (
    <main className="min-h-screen bg-[#f7f7f5] text-[#1f1f1f]">
      <ExploreTopbar
        isAuthenticated={isAuthenticated}
        onProfileClick={() => {
          router.push(hasAuthSession() ? ROUTES.profile : ROUTES.login);
        }}
        profileName={profileName}
      />

      <ExploreHero
        activeCollection={activeCollection}
        collections={collections}
        onCollectionChange={setActiveCollection}
      />

      <section className="mx-auto w-full max-w-370 px-4 pb-14 sm:px-6 lg:px-8">
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
              onLike={handleLikePost}
              onOpen={openPostDetail}
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

        {showsCommunityFeed(activeCollection) ? (
          <LoadMorePosts
            hasNextPage={Boolean(postsQuery.hasNextPage)}
            isLoading={postsQuery.isFetchingNextPage}
            refNode={loadMoreRef}
            totalPosts={visiblePosts.length}
          />
        ) : null}
      </section>
      <PostDetailDialog
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
  onProfileClick,
  profileName,
}: Readonly<{
  isAuthenticated: boolean;
  onProfileClick: () => void;
  profileName: string | null;
}>) {
  const profileLabel =
    isAuthenticated && profileName ? `Profile: ${profileName}` : "Profile";

  return (
    <header className="sticky top-0 z-40 border-b border-black/6 bg-[#f7f7f5]/86 backdrop-blur-2xl">
      <div className="mx-auto flex w-full max-w-370 items-center gap-2 px-3 py-3 sm:px-5 lg:px-8">
        <Link
          aria-label="Explore"
          className="inline-flex size-10 items-center justify-center rounded-full bg-[#111] text-white shadow-[0_14px_30px_-20px_rgb(0_0_0/0.72)] transition hover:scale-[1.03]"
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
            aria-label={profileLabel}
            className={cn(
              "rounded-full",
              isAuthenticated && profileName ? "max-w-44 px-3" : "",
            )}
            onClick={onProfileClick}
            size={isAuthenticated && profileName ? "sm" : "icon-sm"}
            type="button"
            variant="outline"
          >
            <UserRound className="size-4" />
            {isAuthenticated && profileName ? (
              <span className="max-w-28 truncate">{profileName}</span>
            ) : null}
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
  );
}

function ExploreHero({
  activeCollection,
  collections,
  onCollectionChange,
}: Readonly<{
  activeCollection: string;
  collections: string[];
  onCollectionChange: (collection: string) => void;
}>) {
  return (
    <section className="mx-auto w-full max-w-370 px-4 pb-4 pt-6 sm:px-6 lg:px-8">
      <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
        <div className="max-w-3xl">
          <div className="mb-3 inline-flex items-center gap-2 rounded-full border border-black/6 bg-white px-3 py-1.5 text-xs font-semibold text-[#555] shadow-[0_12px_30px_-24px_rgb(0_0_0/0.35)]">
            <Sparkles className="size-3.5 text-[#ff385c]" />
            Curated today
          </div>
          <h1 className="max-w-2xl text-4xl font-semibold tracking-normal text-[#111] sm:text-5xl lg:text-6xl">
            Fresh visual ideas for beautiful stays and memorable travel.
          </h1>
        </div>

        <div className="flex items-center gap-2 lg:justify-end">
          <Button
            className="rounded-full border-black/8 bg-white"
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
          Loading more posts
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
