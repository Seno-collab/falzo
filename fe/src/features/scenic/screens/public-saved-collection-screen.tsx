"use client";

import {
  ArrowLeft,
  Bookmark,
  Camera,
  Compass,
  Loader2,
  MapPin,
  Share2,
} from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useEffect, useMemo } from "react";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import MapClient from "@/components/map";
import type { MapPoint } from "@/components/map";
import { Button } from "@/components/ui/button";
import { getPublicSavedCollectionApi } from "@/features/posts/api";
import { getApiErrorMessage } from "@/features/auth/api";
import { ROUTES } from "@/lib/routes";

export function PublicSavedCollectionScreen({
  shareSlug,
}: Readonly<{
  shareSlug: string;
}>) {
  const normalizedSlug = shareSlug.trim();
  const collectionQuery = useQuery({
    enabled: normalizedSlug.length > 0,
    queryKey: ["posts", "saved-collections", "public", normalizedSlug],
    queryFn: () => getPublicSavedCollectionApi(normalizedSlug),
    retry: false,
  });
  const collection = collectionQuery.data ?? null;
  const posts = collection?.posts ?? [];
  const mapPoints = useMemo<MapPoint[]>(
    () =>
      posts.map((post, index) => ({
        id: String(post.id),
        name: `${index + 1}. ${post.location_name || "Stop"}`,
        address: post.caption || post.user_name,
        imageUrl: post.image_url,
        latitude: post.latitude,
        longitude: post.longitude,
      })),
    [posts],
  );

  useEffect(() => {
    document.title = collection
      ? `${collection.name} | Falzo itinerary`
      : "Shared itinerary | Falzo";
  }, [collection]);

  return (
    <PageShell
      className="bg-[#f7f7f5] text-[#1f1f1f]"
      contentClassName="space-y-5 pb-14"
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
          brand="Shared itinerary"
          brandIcon={<Share2 className="size-3.5" />}
          mobileMenuTitle="Shared"
          subtitle="A public Falzo collection built from saved places."
        />
      }
    >
      {!normalizedSlug ? (
        <EmptySharedState title="Missing collection link" />
      ) : collectionQuery.isLoading ? (
        <div className="flex min-h-96 items-center justify-center rounded-2xl border border-black/6 bg-white">
          <Loader2 className="size-6 animate-spin text-[#777]" />
        </div>
      ) : collectionQuery.error ? (
        <EmptySharedState title={getApiErrorMessage(collectionQuery.error)} />
      ) : collection ? (
        <>
          <section className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_22rem]">
            <div className="space-y-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
                  Public collection
                </p>
                <h1 className="mt-2 max-w-4xl text-4xl font-semibold tracking-normal text-[#111] sm:text-5xl">
                  {collection.name}
                </h1>
                <p className="mt-3 max-w-2xl text-sm leading-6 text-[#666]">
                  {posts.length} saved stop{posts.length === 1 ? "" : "s"} in
                  a route you can scan, save, and explore.
                </p>
              </div>

              {mapPoints.length > 0 ? (
                <MapClient points={mapPoints} zoom={12} />
              ) : null}
            </div>

            <aside className="space-y-3 lg:sticky lg:top-24 lg:self-start">
              <div className="rounded-2xl border border-black/6 bg-white p-4">
                <div className="flex items-center gap-2">
                  <Compass className="size-4 text-[#2f6fb8]" />
                  <h2 className="text-base font-semibold text-[#111]">
                    Itinerary stops
                  </h2>
                </div>
                <div className="mt-4 space-y-3">
                  {posts.map((post, index) => (
                    <a
                      className="block rounded-xl border border-black/6 bg-[#f8f8f6] px-3 py-3 transition hover:border-black/16 hover:bg-white"
                      href={`#post-${post.id}`}
                      key={post.id}
                    >
                      <span className="text-xs font-semibold uppercase tracking-[0.14em] text-[#777]">
                        Stop {index + 1}
                      </span>
                      <p className="mt-1 line-clamp-2 text-sm font-semibold text-[#111]">
                        {post.location_name || post.caption || "Saved place"}
                      </p>
                    </a>
                  ))}
                </div>
              </div>
            </aside>
          </section>

          <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {posts.map((post, index) => (
              <article
                className="overflow-hidden rounded-2xl border border-black/6 bg-white shadow-[0_18px_46px_-36px_rgb(0_0_0/0.55)]"
                id={`post-${post.id}`}
                key={post.id}
              >
                <div className="relative aspect-[4/5] bg-[#ece9e2]">
                  <img
                    alt={post.caption || post.location_name || "Itinerary stop"}
                    className="h-full w-full object-cover"
                    loading="lazy"
                    src={post.image_url}
                  />
                  <span className="absolute left-3 top-3 rounded-full bg-white/90 px-3 py-1 text-xs font-semibold text-[#111] shadow-sm backdrop-blur-xl">
                    Stop {index + 1}
                  </span>
                </div>
                <div className="space-y-2 p-4">
                  <p className="flex items-center gap-1 text-xs font-semibold text-[#777]">
                    <MapPin className="size-3.5" />
                    <span className="truncate">
                      {post.location_name || "Saved place"}
                    </span>
                  </p>
                  <h2 className="line-clamp-2 text-lg font-semibold tracking-normal text-[#111]">
                    {post.caption || "Community post"}
                  </h2>
                  <p className="text-xs font-semibold text-[#777]">
                    By {post.user_name || `User #${post.user_id}`}
                  </p>
                </div>
              </article>
            ))}
          </section>
        </>
      ) : (
        <EmptySharedState title="This itinerary is unavailable." />
      )}
    </PageShell>
  );
}

function EmptySharedState({ title }: Readonly<{ title: string }>) {
  return (
    <div className="flex min-h-96 flex-col items-center justify-center rounded-2xl border border-dashed border-black/10 bg-white/76 px-6 text-center">
      <span className="flex size-12 items-center justify-center rounded-full bg-[#f4f1ec] text-[#777]">
        <Bookmark className="size-5" />
      </span>
      <h1 className="mt-4 text-2xl font-semibold tracking-normal text-[#111]">
        {title}
      </h1>
      <Button asChild className="mt-5 rounded-full">
        <Link href={ROUTES.explore}>
          <Camera className="size-4" />
          Explore Falzo
        </Link>
      </Button>
    </div>
  );
}
