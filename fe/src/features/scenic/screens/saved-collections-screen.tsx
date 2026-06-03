"use client";

import {
  ArrowLeft,
  Bookmark,
  Camera,
  Copy,
  Folder,
  FolderPlus,
  Globe2,
  Loader2,
  MapPin,
  Plus,
  Share2,
  Trash2,
  X,
} from "lucide-react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  addPostToSavedCollectionApi,
  createSavedCollectionApi,
  deleteSavedCollectionApi,
  getSavedCollectionsApi,
  getSavedPostsApi,
  removePostFromSavedCollectionApi,
  unsavePostApi,
  updateSavedCollectionApi,
} from "@/features/posts/api";
import type { Post, SavedCollection } from "@/features/posts/types";
import { getApiErrorMessage, hasAuthSession } from "@/features/auth/api";
import { useI18n } from "@/i18n/locale-provider";
import { ROUTES } from "@/lib/routes";
import { cn } from "@/lib/utils";

const allSavedKey = "all";

function getCollectionPostIds(collection: SavedCollection | null) {
  return new Set((collection?.posts ?? []).map((post) => post.id));
}

function getShareUrl(collection: SavedCollection) {
  if (globalThis.window === undefined) {
    return ROUTES.savedCollection(collection.share_slug);
  }

  return new URL(
    ROUTES.savedCollection(collection.share_slug),
    globalThis.window.location.origin,
  ).toString();
}

function PostTile({
  collections,
  isMutating,
  mode,
  onAddToCollection,
  onOpenCollection,
  onRemoveFromCollection,
  onUnsave,
  post,
}: Readonly<{
  collections: SavedCollection[];
  isMutating: boolean;
  mode: "all" | "collection";
  onAddToCollection: (collectionId: number, postId: number) => void;
  onOpenCollection: (collectionId: number) => void;
  onRemoveFromCollection?: (postId: number) => void;
  onUnsave: (postId: number) => void;
  post: Post;
}>) {
  const linkedCollectionIds = new Set(
    collections
      .filter((collection) =>
        collection.posts.some((item) => item.id === post.id),
      )
      .map((collection) => collection.id),
  );
  const availableCollections = collections.filter(
    (collection) => !linkedCollectionIds.has(collection.id),
  );

  return (
    <article className="overflow-hidden rounded-2xl border border-black/6 bg-white shadow-[0_18px_46px_-36px_rgb(0_0_0/0.55)]">
      <div className="relative aspect-4/5 bg-[#ece9e2]">
        <img
          alt={post.caption || post.location_name || "Saved post"}
          className="h-full w-full object-cover"
          decoding="async"
          fetchPriority="low"
          loading="lazy"
          sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 22rem"
          src={post.image_url}
        />
        <div className="absolute inset-x-3 top-3 flex items-center justify-between gap-2">
          <span className="max-w-[70%] truncate rounded-full bg-white/88 px-3 py-1 text-xs font-semibold text-[#222] shadow-sm backdrop-blur-xl">
            {post.category_name || "Community"}
          </span>
          <Button
            aria-label="Unsave post"
            className="rounded-full bg-white/88 text-[#b4233f] shadow-sm backdrop-blur-xl hover:bg-white"
            disabled={isMutating}
            onClick={() => onUnsave(post.id)}
            size="icon-sm"
            type="button"
            variant="ghost"
          >
            <X className="size-4" />
          </Button>
        </div>
      </div>

      <div className="space-y-3 p-4">
        <div className="min-w-0">
          <h2 className="line-clamp-2 text-base font-semibold tracking-normal text-[#171717]">
            {post.caption || "Community post"}
          </h2>
          <p className="mt-1 flex items-center gap-1 text-xs font-medium text-[#6c6c6c]">
            <MapPin className="size-3.5" />
            <span className="truncate">{post.location_name || "Saved"}</span>
          </p>
        </div>

        {mode === "collection" && onRemoveFromCollection ? (
          <Button
            className="w-full rounded-full"
            disabled={isMutating}
            onClick={() => onRemoveFromCollection(post.id)}
            type="button"
            variant="outline"
          >
            <X className="size-4" />
            Remove from collection
          </Button>
        ) : null}

        {mode === "all" && collections.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            {linkedCollectionIds.size > 0 ? (
              collections
                .filter((collection) => linkedCollectionIds.has(collection.id))
                .map((collection) => (
                  <button
                    className="rounded-full bg-[#eef5f0] px-3 py-1 text-xs font-semibold text-[#2f6847] transition hover:bg-[#e3eee6]"
                    key={collection.id}
                    onClick={() => onOpenCollection(collection.id)}
                    type="button"
                  >
                    {collection.name}
                  </button>
                ))
            ) : (
              <span className="rounded-full bg-[#f4f1ec] px-3 py-1 text-xs font-semibold text-[#776b5f]">
                Unorganized
              </span>
            )}
            {availableCollections.slice(0, 3).map((collection) => (
              <Button
                className="h-7 rounded-full px-3 text-xs"
                disabled={isMutating}
                key={collection.id}
                onClick={() => onAddToCollection(collection.id, post.id)}
                type="button"
                variant="outline"
              >
                <Plus className="size-3.5" />
                {collection.name}
              </Button>
            ))}
          </div>
        ) : null}
      </div>
    </article>
  );
}

export function SavedCollectionsScreen() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { locale } = useI18n();
  const [collectionName, setCollectionName] = useState("");
  const [activeCollectionId, setActiveCollectionId] = useState<
    number | typeof allSavedKey
  >(allSavedKey);

  useEffect(() => {
    document.title = "Saved Collections | Falzo";
    if (!hasAuthSession()) {
      router.replace(ROUTES.login);
    }
  }, [router]);

  const savedPostsQuery = useQuery({
    enabled: hasAuthSession(),
    queryKey: ["posts", "saved", locale],
    queryFn: ({ signal }) => getSavedPostsApi({ signal }),
    retry: false,
    staleTime: 45_000,
  });

  const collectionsQuery = useQuery({
    enabled: hasAuthSession(),
    queryKey: ["posts", "saved-collections", locale],
    queryFn: ({ signal }) => getSavedCollectionsApi({ signal }),
    retry: false,
    staleTime: 45_000,
  });

  const collections = collectionsQuery.data ?? [];
  const savedPosts = savedPostsQuery.data ?? [];
  const activeCollection =
    activeCollectionId === allSavedKey
      ? null
      : (collections.find(
          (collection) => collection.id === activeCollectionId,
        ) ?? null);
  const activePostIds = useMemo(
    () => getCollectionPostIds(activeCollection),
    [activeCollection],
  );
  const visiblePosts =
    activeCollectionId === allSavedKey
      ? savedPosts
      : (activeCollection?.posts ?? []);
  const collectionCandidates =
    activeCollectionId === allSavedKey
      ? []
      : savedPosts.filter((post) => !activePostIds.has(post.id));

  const invalidateSavedData = () => {
    void queryClient.invalidateQueries({ queryKey: ["posts", "saved"] });
    void queryClient.invalidateQueries({
      queryKey: ["posts", "saved-collections"],
    });
  };

  const createCollectionMutation = useMutation({
    mutationFn: createSavedCollectionApi,
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: (collection) => {
      setCollectionName("");
      setActiveCollectionId(collection.id);
      invalidateSavedData();
      toast.success("Collection created.");
    },
  });

  const addPostMutation = useMutation({
    mutationFn: ({
      collectionId,
      postId,
    }: {
      collectionId: number;
      postId: number;
    }) => addPostToSavedCollectionApi(collectionId, postId),
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: () => {
      invalidateSavedData();
      toast.success("Post added to collection.");
    },
  });

  const removePostMutation = useMutation({
    mutationFn: ({
      collectionId,
      postId,
    }: {
      collectionId: number;
      postId: number;
    }) => removePostFromSavedCollectionApi(collectionId, postId),
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: () => {
      invalidateSavedData();
      toast.success("Post removed from collection.");
    },
  });

  const deleteCollectionMutation = useMutation({
    mutationFn: deleteSavedCollectionApi,
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: () => {
      setActiveCollectionId(allSavedKey);
      invalidateSavedData();
      toast.success("Collection deleted.");
    },
  });

  const updateCollectionMutation = useMutation({
    mutationFn: updateSavedCollectionApi,
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: (collection) => {
      invalidateSavedData();
      toast.success(
        collection.is_public
          ? "Collection is public."
          : "Collection is private.",
      );
    },
  });

  const unsaveMutation = useMutation({
    mutationFn: unsavePostApi,
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: () => {
      invalidateSavedData();
      toast.success("Post removed from saved.");
    },
  });

  const isMutating =
    createCollectionMutation.isPending ||
    addPostMutation.isPending ||
    removePostMutation.isPending ||
    deleteCollectionMutation.isPending ||
    updateCollectionMutation.isPending ||
    unsaveMutation.isPending;
  const isLoading = savedPostsQuery.isLoading || collectionsQuery.isLoading;

  function submitCollection(name: string) {
    const nextName = name.trim();
    if (!nextName) {
      toast.error("Collection name is required.");
      return;
    }

    createCollectionMutation.mutate({ name: nextName });
  }

  function copyCollectionLink(collection: SavedCollection) {
    if (!collection.is_public) {
      toast.error("Publish this collection before sharing it.");
      return;
    }

    const shareUrl = getShareUrl(collection);
    void globalThis.navigator?.clipboard
      ?.writeText(shareUrl)
      .then(() => toast.success("Share link copied."))
      .catch(() => toast.error("Unable to copy share link."));
  }

  return (
    <PageShell
      className="bg-[#f7f7f5] text-[#1f1f1f]"
      contentClassName="pb-14"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "explore",
              icon: <ArrowLeft className="size-4" />,
              label: "Explore",
              to: ROUTES.explore,
              variant: "outline",
            },
          ]}
          brand="Saved"
          brandIcon={<Bookmark className="size-3.5" />}
          mobileMenuTitle="Saved"
          subtitle="Collections for trips, ideas, and places worth returning to."
        />
      }
    >
      <div className="grid gap-5 lg:grid-cols-[18rem_minmax(0,1fr)]">
        <aside className="space-y-4 lg:sticky lg:top-24 lg:self-start">
          <section className="rounded-2xl border border-black/6 bg-white p-4 shadow-[0_16px_42px_-34px_rgb(0_0_0/0.5)]">
            <div className="flex items-center gap-2">
              <FolderPlus className="size-4 text-[#a15c2d]" />
              <h1 className="text-base font-semibold tracking-normal text-[#111]">
                New collection
              </h1>
            </div>
            <form
              className="mt-3 space-y-2"
              onSubmit={(event) => {
                event.preventDefault();
                submitCollection(collectionName);
              }}
            >
              <Input
                maxLength={120}
                onChange={(event) => setCollectionName(event.target.value)}
                placeholder="Collection name"
                value={collectionName}
              />
              <Button
                className="w-full rounded-full"
                disabled={createCollectionMutation.isPending}
                type="submit"
              >
                {createCollectionMutation.isPending ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Plus className="size-4" />
                )}
                Create
              </Button>
            </form>
          </section>

          <section className="rounded-2xl border border-black/6 bg-white p-2 shadow-[0_16px_42px_-34px_rgb(0_0_0/0.5)]">
            <button
              className={cn(
                "flex w-full items-center justify-between rounded-xl px-3 py-2 text-left text-sm font-semibold transition",
                activeCollectionId === allSavedKey
                  ? "bg-[#111] text-white"
                  : "text-[#333] hover:bg-[#f4f4f1]",
              )}
              onClick={() => setActiveCollectionId(allSavedKey)}
              type="button"
            >
              <span className="inline-flex min-w-0 items-center gap-2">
                <Bookmark className="size-4" />
                All saved
              </span>
              <span>{savedPosts.length}</span>
            </button>
            <div className="mt-2 space-y-1">
              {collections.map((collection) => (
                <button
                  className={cn(
                    "flex w-full items-center justify-between gap-2 rounded-xl px-3 py-2 text-left text-sm font-semibold transition",
                    activeCollectionId === collection.id
                      ? "bg-[#eef5f0] text-[#24523a]"
                      : "text-[#333] hover:bg-[#f4f4f1]",
                  )}
                  key={collection.id}
                  onClick={() => setActiveCollectionId(collection.id)}
                  type="button"
                >
                  <span className="inline-flex min-w-0 items-center gap-2">
                    <Folder className="size-4 shrink-0" />
                    <span className="truncate">{collection.name}</span>
                  </span>
                  <span>{collection.post_count}</span>
                </button>
              ))}
            </div>
          </section>
        </aside>

        <section className="min-w-0 space-y-5">
          <div className="flex flex-wrap items-end justify-between gap-3">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
                {activeCollection ? "Collection" : "Saved library"}
              </p>
              <h2 className="mt-1 text-3xl font-semibold tracking-normal text-[#111]">
                {activeCollection?.name ?? "All saved posts"}
              </h2>
            </div>
            {activeCollection ? (
              <div className="flex flex-wrap gap-2">
                <Button
                  className="rounded-full"
                  disabled={updateCollectionMutation.isPending}
                  onClick={() =>
                    updateCollectionMutation.mutate({
                      collectionId: activeCollection.id,
                      isPublic: !activeCollection.is_public,
                    })
                  }
                  type="button"
                  variant={activeCollection.is_public ? "default" : "outline"}
                >
                  <Globe2 className="size-4" />
                  {activeCollection.is_public ? "Public" : "Publish"}
                </Button>
                <Button
                  className="rounded-full"
                  disabled={!activeCollection.is_public}
                  onClick={() => copyCollectionLink(activeCollection)}
                  type="button"
                  variant="outline"
                >
                  <Copy className="size-4" />
                  Copy link
                </Button>
                {activeCollection.is_public ? (
                  <Button
                    asChild
                    className="rounded-full"
                    type="button"
                    variant="outline"
                  >
                    <Link
                      href={ROUTES.savedCollection(activeCollection.share_slug)}
                    >
                      <Share2 className="size-4" />
                      View itinerary
                    </Link>
                  </Button>
                ) : null}
                <Button
                  className="rounded-full"
                  disabled={deleteCollectionMutation.isPending}
                  onClick={() =>
                    deleteCollectionMutation.mutate(activeCollection.id)
                  }
                  type="button"
                  variant="outline"
                >
                  <Trash2 className="size-4" />
                  Delete
                </Button>
              </div>
            ) : null}
          </div>

          {activeCollection ? (
            <div className="rounded-2xl border border-black/6 bg-white px-4 py-3 text-sm text-[#555]">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <span className="font-semibold text-[#111]">
                  {activeCollection.is_public
                    ? "Anyone with the link can view this itinerary."
                    : "This collection is private."}
                </span>
                {activeCollection.is_public ? (
                  <span className="truncate text-xs font-semibold text-[#777]">
                    {getShareUrl(activeCollection)}
                  </span>
                ) : null}
              </div>
            </div>
          ) : null}

          {activeCollection && collectionCandidates.length > 0 ? (
            <section className="rounded-2xl border border-black/6 bg-white p-4">
              <div className="flex items-center justify-between gap-3">
                <h3 className="text-sm font-semibold text-[#111]">
                  Add saved posts
                </h3>
                <span className="text-xs font-semibold text-[#777]">
                  {collectionCandidates.length} available
                </span>
              </div>
              <div className="mt-3 flex gap-3 overflow-x-auto pb-1">
                {collectionCandidates.slice(0, 12).map((post) => (
                  <button
                    className="group w-32 shrink-0 overflow-hidden rounded-xl border border-black/6 bg-[#f8f8f6] text-left transition hover:border-black/16"
                    disabled={addPostMutation.isPending}
                    key={post.id}
                    onClick={() =>
                      addPostMutation.mutate({
                        collectionId: activeCollection.id,
                        postId: post.id,
                      })
                    }
                    type="button"
                  >
                    <span className="relative block aspect-square bg-[#ece9e2]">
                      <img
                        alt={post.caption || post.location_name || "Saved post"}
                        className="h-full w-full object-cover transition group-hover:scale-[1.035]"
                        decoding="async"
                        fetchPriority="low"
                        loading="lazy"
                        sizes="8rem"
                        src={post.image_url}
                      />
                    </span>
                    <span className="block truncate px-2 py-2 text-xs font-semibold text-[#333]">
                      {post.location_name || post.caption || "Saved post"}
                    </span>
                  </button>
                ))}
              </div>
            </section>
          ) : null}

          {isLoading ? (
            <div className="flex min-h-80 items-center justify-center rounded-2xl border border-black/6 bg-white">
              <Loader2 className="size-6 animate-spin text-[#777]" />
            </div>
          ) : visiblePosts.length === 0 ? (
            <div className="flex min-h-80 flex-col items-center justify-center rounded-2xl border border-dashed border-black/10 bg-white/76 px-6 text-center">
              <Camera className="size-9 text-[#777]" />
              <h2 className="mt-3 text-xl font-semibold tracking-normal text-[#111]">
                No saved posts here
              </h2>
              <Button asChild className="mt-4 rounded-full">
                <Link href={ROUTES.explore}>Explore posts</Link>
              </Button>
            </div>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
              {visiblePosts.map((post) => (
                <PostTile
                  collections={collections}
                  isMutating={isMutating}
                  key={post.id}
                  mode={activeCollection ? "collection" : "all"}
                  onAddToCollection={(collectionId, postId) =>
                    addPostMutation.mutate({ collectionId, postId })
                  }
                  onOpenCollection={setActiveCollectionId}
                  onRemoveFromCollection={
                    activeCollection
                      ? (postId) =>
                          removePostMutation.mutate({
                            collectionId: activeCollection.id,
                            postId,
                          })
                      : undefined
                  }
                  onUnsave={(postId) => unsaveMutation.mutate(postId)}
                  post={post}
                />
              ))}
            </div>
          )}
        </section>
      </div>
    </PageShell>
  );
}
