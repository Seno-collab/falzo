"use client";

import Image from "next/image";
import Link from "next/link";
import { Heart, MapPinned } from "lucide-react";
import { TagBadge } from "@/components/common/tag-badge";
import { useSavedLocations } from "@/providers/saved-locations-provider";
import type { TravelLocation, TravelPost } from "@/lib/travel/types";
import { cn } from "@/lib/utils";

function toLocationSummary(post: TravelPost): TravelLocation {
  return {
    id: post.locationId,
    name: post.locationName,
    subtitle: post.locationSubtitle,
    lat: 0,
    lng: 0,
    imageUrl: post.imageUrl,
    postsCount: undefined,
  };
}

export function PostCard({ post }: { post: TravelPost }) {
  const { isSaved, toggleSaved } = useSavedLocations();
  const saved = isSaved(post.locationId);

  return (
    <article className="surface surface-hover masonry-item overflow-hidden">
      <div className="relative aspect-[3/4] w-full overflow-hidden bg-[#eff4fb]">
        <Image
          alt={post.caption || post.locationName}
          className="h-full w-full object-cover"
          fill
          loading="lazy"
          sizes="(max-width: 767px) 50vw, (max-width: 1279px) 33vw, 25vw"
          src={post.imageUrl}
        />
      </div>

      <div className="space-y-3 p-3.5">
        <p className="line-clamp-2 text-sm font-semibold leading-6 text-[#183555]">{post.caption}</p>

        <Link className="inline-flex items-center gap-1.5 text-xs font-semibold text-[#3f628b]" href={`/locations/${post.locationId}`}>
          <MapPinned className="size-3.5" />
          {post.locationName}
        </Link>

        {post.tags.length ? (
          <div className="flex flex-wrap gap-1.5">
            {post.tags.slice(0, 3).map((tag) => (
              <TagBadge key={`${post.id}-${tag}`} label={tag} />
            ))}
          </div>
        ) : null}

        <div className="flex items-center justify-between">
          <p className="text-xs text-[#728ba8]">
            {post.likes.toLocaleString()} likes · {post.saves.toLocaleString()} saves
          </p>

          <button
            className={cn(
              "inline-flex h-8 w-8 items-center justify-center rounded-full border transition",
              saved
                ? "border-[#f4c6ce] bg-[#fdeef2] text-[#cc3f53]"
                : "border-[#d4e0ee] bg-white text-[#6f86a2] hover:bg-[#f2f7ff]",
            )}
            onClick={() => toggleSaved(toLocationSummary(post))}
            type="button"
          >
            <Heart className={cn("size-4", saved ? "fill-current" : "")} />
          </button>
        </div>
      </div>
    </article>
  );
}
