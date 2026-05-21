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
  const detailBadge =
    distanceMeters !== undefined
      ? formatDistance(distanceMeters)
      : location.post_count
        ? `${location.post_count} photo${location.post_count === 1 ? "" : "s"}`
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
            Travel stop
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
    document.title = "Locations | Falzo";
    setIsAuthenticated(hasAuthSession());
  }, []);

  const searchQuery = useQuery({
    enabled: submittedSearch.trim().length > 0,
    queryKey: ["locations", "search", submittedSearch],
    queryFn: () => searchLocationsWithFallbackApi(submittedSearch),
  });

  const nearbyQuery = useQuery({
    enabled: coords !== null,
    queryKey: ["locations", "nearby", coords, radiusMeters],
    queryFn: () =>
      getNearbyLocationsApi({
        latitude: coords?.latitude ?? 0,
        longitude: coords?.longitude ?? 0,
        radiusMeters,
      }),
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
    queryFn: () => {
      if (!selectedLocation) {
        return [];
      }

      if (isPostBackedLocation(selectedLocation)) {
        return getPostBackedLocationPostsApi(selectedLocation);
      }

      return getLocationPostsApi(selectedLocation.id);
    },
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
      ? `${selectedTravelPhotoCount} travel photo${
          selectedTravelPhotoCount === 1 ? "" : "s"
        } - `
      : "";
  const selectedLocationSubtitle = selectedLocation
    ? `${selectedTravelPhotoLabel}${selectedLocation.latitude.toFixed(4)}, ${selectedLocation.longitude.toFixed(4)}`
    : "Choose a marker or search result";

  function useCurrentPosition() {
    if (!navigator.geolocation) {
      toast.error("This browser does not support location access.");
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
        toast.error(error.message || "Unable to read your current location.");
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
              label: "Explore",
              to: ROUTES.explore,
              variant: "outline",
            },
            {
              id: "upload",
              icon: <Upload className="size-4" />,
              label: "Upload",
              to: ROUTES.upload,
              variant: "default",
            },
            {
              id: "back",
              icon: <ArrowLeft className="size-4" />,
              label: "Explore",
              to: ROUTES.explore,
              variant: "outline",
            },
          ]}
          brand="Falzo Locations"
          brandIcon={<MapIcon className="size-3.5" />}
          mobileMenuTitle="Locations"
          subtitle="Search places, discover nearby locations, and review location posts."
        />
      }
    >
      <section className="rounded-[2rem] border border-black/6 bg-white px-4 py-5 shadow-[0_18px_50px_-42px_rgb(0_0_0/0.62)] sm:px-6">
        <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
              Travel map
            </p>
            <h1 className="mt-1 max-w-3xl text-3xl font-semibold leading-tight tracking-normal text-[#111] sm:text-4xl">
              Find places, nearby stops, and travel photos by location.
            </h1>
            <p className="mt-3 max-w-2xl text-sm leading-6 text-[#666] sm:text-base">
              Search a city or province, tap a marker, then review the travel
              posts connected to that place.
            </p>
          </div>
          <div className="grid grid-cols-2 gap-2 sm:flex sm:items-center">
            <div className="rounded-2xl border border-black/6 bg-[#f8f8f6] px-4 py-3">
              <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#777]">
                Visible
              </p>
              <p className="mt-1 text-lg font-semibold text-[#111]">
                {mapPoints.length} place{mapPoints.length === 1 ? "" : "s"}
              </p>
            </div>
            <div className="rounded-2xl border border-black/6 bg-[#f8f8f6] px-4 py-3">
              <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#777]">
                Selected
              </p>
              <p className="mt-1 max-w-44 truncate text-lg font-semibold text-[#111]">
                {selectedLocation?.name ?? "None"}
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
                Search
              </p>
              <h2 className="text-2xl font-semibold tracking-normal text-[#111]">
                Where do you want to go?
              </h2>
              <p className="text-sm leading-6 text-[#666]">
                Try a destination name, city, province, or landmark.
              </p>
            </div>
            <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto]">
              <div className="space-y-2">
                <Label htmlFor="location-search">Destination</Label>
                <Input
                  id="location-search"
                  onChange={(event) => setSearchInput(event.target.value)}
                  placeholder="TP.HCM, Ho Chi Minh, Da Nang, Kyoto"
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
                Search
              </Button>
            </div>
          </form>

          <section className="space-y-4 overflow-hidden rounded-2xl border border-black/6 bg-white shadow-[0_14px_38px_-32px_rgb(0_0_0/0.58)]">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="space-y-1 px-5 pt-5 sm:px-6">
                <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
                  Map discovery
                </p>
                <h2 className="text-xl font-semibold tracking-normal text-[#111]">
                  Tap a marker to open a travel stop
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
                  Nearby
                </p>
                <h2 className="text-xl font-semibold tracking-normal text-[#111]">
                  Places around you
                </h2>
                <p className="text-sm leading-6 text-[#666]">
                  Use your current location to find close travel stops.
                </p>
              </div>
              <div className="flex items-end gap-2">
                <div className="w-32 space-y-2">
                  <Label htmlFor="radius">Radius</Label>
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
                  Locate
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
                    Use current location to find nearby places.
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
                Results
              </p>
              <h2 className="text-xl font-semibold tracking-normal text-[#111]">
                Destination matches
              </h2>
              <p className="text-sm leading-6 text-[#666]">
                Select a place to see its map position and travel posts.
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
                  Searching
                </div>
              ) : null}

              {!searchQuery.isFetching && searchQuery.data?.length === 0 ? (
                <div className="app-empty-state">
                  <span className="app-empty-icon">
                    <Search className="size-5" />
                  </span>
                  <p className="text-sm font-semibold text-[#315578]">
                    No locations found.
                  </p>
                </div>
              ) : null}
            </div>
          </section>

          <section className="space-y-4 rounded-2xl border border-black/6 bg-white p-5 shadow-[0_14px_38px_-32px_rgb(0_0_0/0.58)] sm:p-6">
            <div className="space-y-1">
              <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#777]">
                Travel photos
              </p>
              <h2 className="text-xl font-semibold tracking-normal text-[#111]">
                {selectedLocation ? selectedLocation.name : "Select a location"}
              </h2>
              <p className="text-sm leading-6 text-[#666]">
                Photos and notes shared from this destination.
              </p>
            </div>

            <div className="grid gap-3">
              {postsQuery.isFetching ? (
                <div className="flex items-center gap-2 rounded-xl border border-[#d7e5f4] bg-white/90 px-4 py-3 text-sm text-[#6682a1]">
                  <Loader2 className="size-4 animate-spin" />
                  Loading posts
                </div>
              ) : null}

              {postsQuery.data?.map((post) => (
                <article
                  className="overflow-hidden rounded-2xl border border-black/6 bg-white shadow-[0_14px_30px_-24px_rgb(0_0_0/0.48)]"
                  key={post.id}
                >
                  <img
                    alt={post.caption || post.location_name || "Location post"}
                    className="aspect-[4/3] w-full object-cover"
                    loading="lazy"
                    src={post.image_url}
                  />
                  <div className="space-y-1 px-4 py-3">
                    <p className="line-clamp-2 text-sm font-semibold text-[#111]">
                      {post.caption || "Location post"}
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
                    No posts for this location yet.
                  </p>
                  {isAuthenticated ? (
                    <Button asChild size="sm" variant="outline">
                      <Link href={ROUTES.upload}>
                        <Upload className="size-4" />
                        Upload
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
                    No posts for this location yet.
                  </p>
                  {isAuthenticated ? (
                    <Button asChild size="sm" variant="outline">
                      <Link href={ROUTES.upload}>
                        <Upload className="size-4" />
                        Upload
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
