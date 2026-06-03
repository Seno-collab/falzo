"use client";

import { useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  Compass,
  Crosshair,
  ImageIcon,
  Loader2,
  Map as MapIcon,
  MapPin,
  Navigation,
  Search,
  Upload,
} from "lucide-react";
import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import MapClient, { type MapPoint } from "@/components/map";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { getApiErrorMessage, hasAuthSession } from "@/features/auth/api";
import {
  getLocationPostsApi,
  getNearbyLocationsApi,
} from "@/features/locations/api";
import {
  defaultLocationSearch,
  getPostBackedLocationPostsApi,
  isGeocodedLocation,
  isPostBackedLocation,
  normalizeLocationSearchQuery,
  searchLocationsWithFallbackApi,
} from "@/features/locations/search";
import type { Location, NearbyLocation } from "@/features/locations/types";
import { useI18n } from "@/i18n/locale-provider";
import { ROUTES } from "@/lib/routes";
import { cn } from "@/lib/utils";

type Coordinates = {
  latitude: number;
  longitude: number;
};

const defaultRadiusMeters = 5000;

function formatDistance(meters: number) {
  if (meters >= 1000) {
    return `${(meters / 1000).toFixed(1)} km`;
  }

  return `${Math.round(meters)} m`;
}

function formatCountLabel(count: number, singular: string, plural: string) {
  return `${count} ${count === 1 ? singular : plural}`;
}

function LocationRow({
  location,
  distanceMeters,
  selected,
  onSelect,
}: Readonly<{
  location: Location;
  distanceMeters?: number;
  selected: boolean;
  onSelect: () => void;
}>) {
  const { messages } = useI18n();
  const copy = messages.locationsPage;
  const detailBadge =
    distanceMeters !== undefined
      ? formatDistance(distanceMeters)
      : location.post_count
        ? formatCountLabel(
            location.post_count,
            copy.photoSingular,
            copy.photoPlural,
          )
        : null;

  return (
    <button
      className={cn(
        "w-full rounded-2xl border px-4 py-3 text-left transition",
        selected
          ? "border-[#111] bg-[#f5f7f4] shadow-[0_14px_30px_-24px_rgb(0_0_0/0.52)]"
          : "border-black/6 bg-white hover:border-black/14 hover:bg-[#fbfbfa]",
      )}
      onClick={onSelect}
      type="button"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-[#777]">
            {copy.travelStop}
          </p>
          <p className="mt-1 truncate text-sm font-semibold text-[#111]">
            {location.name}
          </p>
          <p className="mt-1 line-clamp-2 text-xs leading-5 text-[#666]">
            {location.address}
          </p>
        </div>
        {detailBadge ? (
          <span className="shrink-0 rounded-full bg-[#f2f7fd] px-2.5 py-1 text-xs font-semibold text-[#356792]">
            {detailBadge}
          </span>
        ) : null}
      </div>
      <p className="mt-2 text-xs font-medium text-[#888]">
        {location.latitude.toFixed(5)}, {location.longitude.toFixed(5)}
      </p>
    </button>
  );
}

export function LocationsScreen() {
  const { locale, messages } = useI18n();
  const copy = messages.locationsPage;
  const [searchInput, setSearchInput] = useState("");
  const [submittedSearch, setSubmittedSearch] = useState(defaultLocationSearch);
  const [coords, setCoords] = useState<Coordinates | null>(null);
  const [radiusMeters, setRadiusMeters] = useState(defaultRadiusMeters);
  const [selectedLocation, setSelectedLocation] = useState<Location | null>(
    null,
  );
  const [isLocating, setIsLocating] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    document.title = copy.documentTitle;
    setIsAuthenticated(hasAuthSession());
  }, [copy.documentTitle]);

  const searchQuery = useQuery({
    enabled: submittedSearch.trim().length > 0,
    queryKey: ["locations", "search", submittedSearch, locale],
    queryFn: ({ signal }) =>
      searchLocationsWithFallbackApi(submittedSearch, signal, locale),
  });

  const nearbyQuery = useQuery({
    enabled: coords !== null,
    queryKey: ["locations", "nearby", coords, radiusMeters],
    queryFn: ({ signal }) =>
      getNearbyLocationsApi({
        latitude: coords?.latitude ?? 0,
        longitude: coords?.longitude ?? 0,
        radiusMeters,
        signal,
      }),
    staleTime: 60_000,
  });

  const postsQuery = useQuery({
    enabled:
      selectedLocation !== null &&
      (!isGeocodedLocation(selectedLocation) ||
        isPostBackedLocation(selectedLocation)),
    queryKey: [
      "locations",
      selectedLocation?.id,
      selectedLocation?.post_ids,
      "posts",
    ],
    queryFn: ({ signal }) => {
      if (!selectedLocation) {
        return [];
      }

      if (isPostBackedLocation(selectedLocation)) {
        return getPostBackedLocationPostsApi(selectedLocation, signal);
      }

      return getLocationPostsApi(selectedLocation.id, { signal });
    },
    staleTime: 45_000,
  });

  const nearbyLocations = useMemo<NearbyLocation[]>(
    () => nearbyQuery.data ?? [],
    [nearbyQuery.data],
  );
  const selectedLocationPosts = useMemo(
    () => postsQuery.data ?? [],
    [postsQuery.data],
  );
  const selectedTravelPhotoCount =
    selectedLocationPosts.length || selectedLocation?.post_count || 0;
  const mapPoints = useMemo<MapPoint[]>(() => {
    const points = new Map<string, MapPoint>();
    const toPostIds = () =>
      selectedLocationPosts
        .map((post) => Number(post.id))
        .filter((postId) => Number.isFinite(postId));
    const toMapPoint = (
      location: Location,
      distanceMeters?: number,
    ): MapPoint => {
      const isSelected = selectedLocation?.id === location.id;
      const selectedPostIds = isSelected ? toPostIds() : [];
      const point: MapPoint = {
        id: location.id,
        name: location.name,
        address: location.address,
        latitude: location.latitude,
        longitude: location.longitude,
      };

      if (distanceMeters !== undefined) {
        point.distanceMeters = distanceMeters;
      }

      if (isSelected) {
        if (selectedTravelPhotoCount > 0) {
          point.count = selectedTravelPhotoCount;
        }
        if (selectedLocationPosts[0]?.image_url) {
          point.imageUrl = selectedLocationPosts[0].image_url;
        }
        if (selectedPostIds.length > 0) {
          point.postIds = selectedPostIds;
        }
      } else {
        if (location.post_count) {
          point.count = location.post_count;
        }
        if (location.post_ids?.length) {
          point.postIds = location.post_ids;
        }
      }

      return point;
    };

    for (const location of searchQuery.data ?? []) {
      points.set(location.id, toMapPoint(location));
    }

    for (const item of nearbyLocations) {
      points.set(
        item.location.id,
        toMapPoint(item.location, item.distance_meters),
      );
    }

    if (selectedLocation && !points.has(selectedLocation.id)) {
      points.set(selectedLocation.id, toMapPoint(selectedLocation));
    }

    return Array.from(points.values());
  }, [
    nearbyLocations,
    searchQuery.data,
    selectedLocation,
    selectedLocationPosts,
    selectedTravelPhotoCount,
  ]);
  const selectedTravelPhotoLabel =
    selectedTravelPhotoCount > 0
      ? `${formatCountLabel(
          selectedTravelPhotoCount,
          copy.travelPhotoSingular,
          copy.travelPhotoPlural,
        )} - `
      : "";
  const selectedLocationSubtitle = selectedLocation
    ? `${selectedTravelPhotoLabel}${selectedLocation.latitude.toFixed(4)}, ${selectedLocation.longitude.toFixed(4)}`
    : copy.chooseMarkerOrSearchResult;

  function useCurrentPosition() {
    if (!navigator.geolocation) {
      toast.error(copy.noLocationAccess);
      return;
    }

    setIsLocating(true);
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setCoords({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
        });
        setIsLocating(false);
      },
      (error) => {
        toast.error(error.message || copy.unableToReadLocation);
        setIsLocating(false);
      },
      {
        enableHighAccuracy: true,
        maximumAge: 30_000,
        timeout: 12_000,
      },
    );
  }

  useEffect(() => {
    if (searchQuery.error) {
      toast.error(getApiErrorMessage(searchQuery.error));
    }
  }, [searchQuery.error]);

  useEffect(() => {
    if (nearbyQuery.error) {
      toast.error(getApiErrorMessage(nearbyQuery.error));
    }
  }, [nearbyQuery.error]);

  useEffect(() => {
    if (postsQuery.error) {
      toast.error(getApiErrorMessage(postsQuery.error));
    }
  }, [postsQuery.error]);

  return (
    <PageShell
      contentClassName="space-y-5 pb-10"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "explore",
              icon: <Compass className="size-4" />,
              label: messages.common.explore,
              to: ROUTES.explore,
              variant: "outline",
            },
            {
              id: "upload",
              icon: <Upload className="size-4" />,
              label: messages.common.upload,
              to: ROUTES.upload,
              variant: "default",
            },
            {
              id: "back",
              icon: <ArrowLeft className="size-4" />,
              label: messages.common.explore,
              to: ROUTES.explore,
              variant: "outline",
            },
          ]}
          brand={copy.brand}
          brandIcon={<MapIcon className="size-3.5" />}
          mobileMenuTitle={copy.mobileMenuTitle}
          subtitle={copy.topbarSubtitle}
        />
      }
    >
      <section className="rounded-4xl border border-black/6 bg-white px-4 py-5 shadow-[0_18px_50px_-42px_rgb(0_0_0/0.62)] sm:px-6">
        <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
              {copy.travelMap}
            </p>
            <h1 className="mt-1 max-w-3xl text-3xl font-semibold leading-tight tracking-normal text-[#111] sm:text-4xl">
              {copy.heroTitle}
            </h1>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-[#666] sm:text-base">
              {copy.heroDescription}
            </p>
          </div>
          <div className="grid grid-cols-2 gap-2 sm:flex sm:items-center">
            <div className="rounded-2xl border border-black/6 bg-[#f8f8f6] px-4 py-3">
              <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#777]">
                {copy.visible}
              </p>
              <p className="mt-1 text-lg font-semibold text-[#111]">
                {formatCountLabel(
                  mapPoints.length,
                  copy.placeSingular,
                  copy.placePlural,
                )}
              </p>
            </div>
            <div className="rounded-2xl border border-black/6 bg-[#f8f8f6] px-4 py-3">
              <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#777]">
                {copy.selected}
              </p>
              <p className="mt-1 max-w-44 truncate text-lg font-semibold text-[#111]">
                {selectedLocation?.name ?? copy.none}
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="grid gap-5 lg:grid-cols-[minmax(0,0.82fr)_minmax(360px,0.58fr)]">
        <div className="space-y-5">
          <form
            className="space-y-4 rounded-2xl border border-black/6 bg-white p-5 shadow-[0_14px_38px_-32px_rgb(0_0_0/0.58)] sm:p-6"
            onSubmit={(event) => {
              event.preventDefault();
              setSubmittedSearch(normalizeLocationSearchQuery(searchInput));
            }}
          >
            <div className="space-y-1">
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
                {copy.searchLabel}
              </p>
              <h2 className="text-2xl font-semibold tracking-normal text-[#111]">
                {copy.searchTitle}
              </h2>
              <p className="text-sm leading-6 text-[#666]">
                {copy.searchDescription}
              </p>
            </div>
            <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto]">
              <div className="space-y-2">
                <Label htmlFor="location-search">{copy.destinationLabel}</Label>
                <Input
                  id="location-search"
                  onChange={(event) => setSearchInput(event.target.value)}
                  placeholder={copy.destinationPlaceholder}
                  value={searchInput}
                />
              </div>
              <Button
                className="self-end"
                disabled={searchQuery.isFetching}
                type="submit"
                variant="gradient"
              >
                {searchQuery.isFetching ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Search className="size-4" />
                )}
                {copy.searchButton}
              </Button>
            </div>
          </form>

          <section className="space-y-4 overflow-hidden rounded-2xl border border-black/6 bg-white shadow-[0_14px_38px_-32px_rgb(0_0_0/0.58)]">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="space-y-1 px-5 pt-5 sm:px-6">
                <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
                  {copy.mapDiscovery}
                </p>
                <h2 className="text-xl font-semibold tracking-normal text-[#111]">
                  {copy.mapTitle}
                </h2>
                <p className="text-sm leading-6 text-[#666]">
                  {selectedLocationSubtitle}
                </p>
              </div>
              {selectedLocation ? (
                <span className="mr-5 mt-5 rounded-full border border-black/8 bg-[#f8f8f6] px-3 py-1 text-xs font-semibold text-[#333] sm:mr-6">
                  {selectedLocation.name}
                </span>
              ) : null}
            </div>
            <MapClient
              className="rounded-none border-0 shadow-none"
              currentPosition={coords}
              onSelectPoint={(point) => {
                const nextLocation =
                  searchQuery.data?.find(
                    (location) => location.id === point.id,
                  ) ??
                  nearbyLocations.find((item) => item.location.id === point.id)
                    ?.location ??
                  null;

                if (nextLocation) {
                  setSelectedLocation(nextLocation);
                }
              }}
              points={mapPoints}
              selectedPointId={selectedLocation?.id}
            />
          </section>

          <section className="space-y-4 rounded-2xl border border-black/6 bg-white p-5 shadow-[0_14px_38px_-32px_rgb(0_0_0/0.58)] sm:p-6">
            <div className="flex flex-wrap items-end justify-between gap-3">
              <div className="space-y-1">
                <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
                  {copy.nearbyLabel}
                </p>
                <h2 className="text-xl font-semibold tracking-normal text-[#111]">
                  {copy.nearbyTitle}
                </h2>
                <p className="text-sm leading-6 text-[#666]">
                  {copy.nearbyDescription}
                </p>
              </div>
              <div className="flex items-end gap-2">
                <div className="w-32 space-y-2">
                  <Label htmlFor="radius">{copy.radiusLabel}</Label>
                  <Input
                    id="radius"
                    min={100}
                    onChange={(event) =>
                      setRadiusMeters(Number(event.target.value))
                    }
                    step={100}
                    type="number"
                    value={radiusMeters}
                  />
                </div>
                <Button
                  disabled={isLocating || nearbyQuery.isFetching}
                  onClick={useCurrentPosition}
                  type="button"
                  variant="outline"
                >
                  {isLocating || nearbyQuery.isFetching ? (
                    <Loader2 className="size-4 animate-spin" />
                  ) : (
                    <Crosshair className="size-4" />
                  )}
                  {copy.locateButton}
                </Button>
              </div>
            </div>

            {coords ? (
              <div className="flex items-center gap-3 rounded-xl border border-[#d7e5f4] bg-[#f7fbff] px-4 py-3 text-sm text-[#385c80]">
                <Navigation className="size-4 shrink-0 text-[#2f6fb8]" />
                <span>
                  {coords.latitude.toFixed(5)}, {coords.longitude.toFixed(5)}
                </span>
              </div>
            ) : null}

            <div className="grid gap-3">
              {nearbyLocations.length > 0 ? (
                nearbyLocations.map((item) => (
                  <LocationRow
                    distanceMeters={item.distance_meters}
                    key={item.location.id}
                    location={item.location}
                    onSelect={() => setSelectedLocation(item.location)}
                    selected={selectedLocation?.id === item.location.id}
                  />
                ))
              ) : (
                <div className="app-empty-state">
                  <span className="app-empty-icon">
                    <MapPin className="size-5" />
                  </span>
                  <p className="text-sm font-semibold text-[#315578]">
                    {copy.nearbyEmpty}
                  </p>
                </div>
              )}
            </div>
          </section>
        </div>

        <aside className="space-y-5">
          <section className="space-y-4 rounded-2xl border border-black/6 bg-white p-5 shadow-[0_14px_38px_-32px_rgb(0_0_0/0.58)] sm:p-6">
            <div className="space-y-1">
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
                {copy.resultsLabel}
              </p>
              <h2 className="text-xl font-semibold tracking-normal text-[#111]">
                {copy.resultsTitle}
              </h2>
              <p className="text-sm leading-6 text-[#666]">
                {copy.resultsDescription}
              </p>
            </div>

            <div className="grid gap-3">
              {searchQuery.data?.map((location) => (
                <LocationRow
                  key={location.id}
                  location={location}
                  onSelect={() => setSelectedLocation(location)}
                  selected={selectedLocation?.id === location.id}
                />
              ))}

              {searchQuery.isFetching ? (
                <div className="flex items-center gap-2 rounded-xl border border-[#d7e5f4] bg-white/90 px-4 py-3 text-sm text-[#6682a1]">
                  <Loader2 className="size-4 animate-spin" />
                  {copy.searching}
                </div>
              ) : null}

              {!searchQuery.isFetching && searchQuery.data?.length === 0 ? (
                <div className="app-empty-state">
                  <span className="app-empty-icon">
                    <Search className="size-5" />
                  </span>
                  <p className="text-sm font-semibold text-[#315578]">
                    {copy.noLocations}
                  </p>
                </div>
              ) : null}
            </div>
          </section>

          <section className="space-y-4 rounded-2xl border border-black/6 bg-white p-5 shadow-[0_14px_38px_-32px_rgb(0_0_0/0.58)] sm:p-6">
            <div className="space-y-1">
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
                {copy.travelPhotosLabel}
              </p>
              <h2 className="text-xl font-semibold tracking-normal text-[#111]">
                {selectedLocation ? selectedLocation.name : copy.selectLocation}
              </h2>
              <p className="text-sm leading-6 text-[#666]">
                {copy.travelPhotosDescription}
              </p>
            </div>

            <div className="grid gap-3">
              {postsQuery.isFetching ? (
                <div className="flex items-center gap-2 rounded-xl border border-[#d7e5f4] bg-white/90 px-4 py-3 text-sm text-[#6682a1]">
                  <Loader2 className="size-4 animate-spin" />
                  {copy.loadingPosts}
                </div>
              ) : null}

              {postsQuery.data?.map((post) => (
                <article
                  className="overflow-hidden rounded-2xl border border-black/6 bg-white shadow-[0_14px_30px_-24px_rgb(0_0_0/0.48)]"
                  key={post.id}
                >
                  <img
                    alt={post.caption || post.location_name || copy.locationPost}
                    className="aspect-4/3 w-full object-cover"
                    decoding="async"
                    fetchPriority="low"
                    loading="lazy"
                    sizes="(max-width: 1024px) 100vw, 24rem"
                    src={post.image_url}
                  />
                  <div className="space-y-1 px-4 py-3">
                    <p className="line-clamp-2 text-sm font-semibold text-[#111]">
                      {post.caption || copy.locationPost}
                    </p>
                    <p className="text-xs font-medium text-[#777]">
                      {post.location_name || selectedLocation?.name}
                    </p>
                  </div>
                </article>
              ))}

              {selectedLocation && isGeocodedLocation(selectedLocation) ? (
                <div className="app-empty-state">
                  <span className="app-empty-icon">
                    <ImageIcon className="size-5" />
                  </span>
                  <p className="text-sm font-semibold text-[#315578]">
                    {copy.noPosts}
                  </p>
                  {isAuthenticated ? (
                    <Button asChild size="sm" variant="outline">
                      <Link href={ROUTES.upload}>
                        <Upload className="size-4" />
                        {copy.upload}
                      </Link>
                    </Button>
                  ) : null}
                </div>
              ) : null}

              {selectedLocation &&
              !isGeocodedLocation(selectedLocation) &&
              !postsQuery.isFetching &&
              postsQuery.data?.length === 0 ? (
                <div className="app-empty-state">
                  <span className="app-empty-icon">
                    <ImageIcon className="size-5" />
                  </span>
                  <p className="text-sm font-semibold text-[#315578]">
                    {copy.noPosts}
                  </p>
                  {isAuthenticated ? (
                    <Button asChild size="sm" variant="outline">
                      <Link href={ROUTES.upload}>
                        <Upload className="size-4" />
                        {copy.upload}
                      </Link>
                    </Button>
                  ) : null}
                </div>
              ) : null}
            </div>
          </section>
        </aside>
      </section>
    </PageShell>
  );
}
