"use client";

import Image from "next/image";
import Link from "next/link";
import { Heart, MapPinned } from "lucide-react";
import { Button } from "@/components/common/button";
import { useSavedLocations } from "@/providers/saved-locations-provider";
import type { TravelLocation } from "@/lib/travel/types";

export function LocationCard({
  location,
  compact = false,
}: {
  location: TravelLocation;
  compact?: boolean;
}) {
  const { isSaved, toggleSaved } = useSavedLocations();
  const saved = isSaved(location.id);

  return (
    <article className="surface surface-hover overflow-hidden">
      <div className={compact ? "flex gap-3 p-3" : "space-y-3 p-3"}>
        <div className={compact ? "relative h-20 w-24 shrink-0 overflow-hidden rounded-lg" : "relative aspect-[16/9] w-full overflow-hidden rounded-xl"}>
          <Image
            alt={location.name}
            className="h-full w-full object-cover"
            fill
            loading="lazy"
            sizes={compact ? "96px" : "(max-width: 767px) 100vw, 33vw"}
            src={
              location.imageUrl ||
              "https://images.unsplash.com/photo-1469474968028-56623f02e42e?auto=format&fit=crop&w=1000&q=80"
            }
          />
        </div>

        <div className="min-w-0 flex-1 space-y-2">
          <div className="space-y-1">
            <Link className="block truncate text-sm font-semibold text-[#153353]" href={`/locations/${location.id}`}>
              {location.name}
            </Link>
            <p className="inline-flex items-center gap-1 text-xs text-[#617b9a]">
              <MapPinned className="size-3.5" />
              {location.subtitle || "Unknown region"}
            </p>
          </div>

          <div className="flex items-center justify-between">
            <p className="text-xs text-[#6e86a4]">{location.postsCount ?? 0} posts</p>
            <Button
              onClick={() => toggleSaved(location)}
              size="sm"
              type="button"
              variant={saved ? "primary" : "secondary"}
            >
              <Heart className={saved ? "size-4 fill-current" : "size-4"} />
              {saved ? "Saved" : "Save"}
            </Button>
          </div>
        </div>
      </div>
    </article>
  );
}
