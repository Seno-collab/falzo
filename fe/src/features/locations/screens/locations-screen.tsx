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
import { getApiErrorMessage } from "@/features/auth/api";
import {
  getLocationPostsApi,
  getNearbyLocationsApi,
} from "@/features/locations/api";
import {
  defaultLocationSearch,
  isGeocodedLocation,
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
  return (
    <button
      className={cn(
        "w-full rounded-xl border px-4 py-3 text-left transition",
        selected
          ? "border-[#2f6fb8] bg-[#eef7ff] shadow-[0_14px_30px_-24px_rgb(47_111_184/0.8)]"
          : "border-[#d7e5f4] bg-white/90 hover:border-[#a9c8e8] hover:bg-[#f8fbff]",
      )}
      onClick={onSelect}
      type="button"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-[#15365a]">
            {location.name}
          </p>
          <p className="mt-1 line-clamp-2 text-xs leading-5 text-[#6682a1]">
            {location.address}
          </p>
        </div>
        {distanceMeters === undefined ? null : (
          <span className="shrink-0 rounded-full bg-[#f2f7fd] px-2.5 py-1 text-xs font-semibold text-[#356792]">
            {formatDistance(distanceMeters)}
          </span>
        )}
      </div>
      <p className="mt-2 text-xs font-medium text-[#7b92ad]">
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

  useEffect(() => {
    document.title = "Locations | Falzo";
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
    enabled: selectedLocation !== null && !isGeocodedLocation(selectedLocation),
    queryKey: ["locations", selectedLocation?.id, "posts"],
    queryFn: () => getLocationPostsApi(selectedLocation?.id ?? ""),
  });

  const nearbyLocations = useMemo<NearbyLocation[]>(
    () => nearbyQuery.data ?? [],
    [nearbyQuery.data],
  );
  const mapPoints = useMemo<MapPoint[]>(() => {
    const points = new Map<string, MapPoint>();

    for (const location of searchQuery.data ?? []) {
      points.set(location.id, {
        id: location.id,
        name: location.name,
        address: location.address,
        latitude: location.latitude,
        longitude: location.longitude,
      });
    }

    for (const item of nearbyLocations) {
      points.set(item.location.id, {
        id: item.location.id,
        name: item.location.name,
        address: item.location.address,
        latitude: item.location.latitude,
        longitude: item.location.longitude,
        distanceMeters: item.distance_meters,
      });
    }

    if (selectedLocation && !points.has(selectedLocation.id)) {
      points.set(selectedLocation.id, {
        id: selectedLocation.id,
        name: selectedLocation.name,
        address: selectedLocation.address,
        latitude: selectedLocation.latitude,
        longitude: selectedLocation.longitude,
      });
    }

    return Array.from(points.values());
  }, [nearbyLocations, searchQuery.data, selectedLocation]);

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
      <section className="grid gap-5 lg:grid-cols-[minmax(0,0.82fr)_minmax(360px,0.58fr)]">
        <div className="space-y-5">
          <form
            className="app-panel space-y-4 rounded-2xl border-[#d6e5f6] bg-white/92 p-5 sm:p-6"
            onSubmit={(event) => {
              event.preventDefault();
              setSubmittedSearch(normalizeLocationSearchQuery(searchInput));
            }}
          >
            <div className="space-y-1">
              <p className="app-kicker">Search</p>
              <h1 className="text-2xl font-semibold tracking-normal text-[#15365a]">
                Find locations
              </h1>
            </div>
            <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto]">
              <div className="space-y-2">
                <Label htmlFor="location-search">Name</Label>
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

          <section className="app-panel space-y-4 rounded-2xl border-[#d6e5f6] bg-white/92 p-5 sm:p-6">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div className="space-y-1">
                <p className="app-kicker">Map</p>
                <h2 className="text-xl font-semibold tracking-normal text-[#15365a]">
                  Explore locations visually
                </h2>
              </div>
              {selectedLocation ? (
                <span className="rounded-full border border-[#d7e5f4] bg-[#f7fbff] px-3 py-1 text-xs font-semibold text-[#356792]">
                  {selectedLocation.name}
                </span>
              ) : null}
            </div>
            <MapClient
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

          <section className="app-panel space-y-4 rounded-2xl border-[#d6e5f6] bg-white/92 p-5 sm:p-6">
            <div className="flex flex-wrap items-end justify-between gap-3">
              <div className="space-y-1">
                <p className="app-kicker">Nearby</p>
                <h2 className="text-xl font-semibold tracking-normal text-[#15365a]">
                  Locations around you
                </h2>
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
              <div className="app-panel-soft flex items-center gap-3 rounded-xl border-[#d7e5f4] bg-[#f7fbff] px-4 py-3 text-sm text-[#385c80]">
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
          <section className="app-panel space-y-4 rounded-2xl border-[#d6e5f6] bg-white/92 p-5 sm:p-6">
            <div className="space-y-1">
              <p className="app-kicker">Results</p>
              <h2 className="text-xl font-semibold tracking-normal text-[#15365a]">
                Search matches
              </h2>
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

          <section className="app-panel space-y-4 rounded-2xl border-[#d6e5f6] bg-white/92 p-5 sm:p-6">
            <div className="space-y-1">
              <p className="app-kicker">Posts</p>
              <h2 className="text-xl font-semibold tracking-normal text-[#15365a]">
                {selectedLocation ? selectedLocation.name : "Select a location"}
              </h2>
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
                  className="overflow-hidden rounded-xl border border-[#d7e5f4] bg-white/95 shadow-[0_14px_30px_-24px_rgb(28_77_128/0.55)]"
                  key={post.id}
                >
                  <img
                    alt={post.caption || post.location_name || "Location post"}
                    className="h-44 w-full object-cover"
                    loading="lazy"
                    src={post.image_url}
                  />
                  <div className="space-y-1 px-4 py-3">
                    <p className="line-clamp-2 text-sm font-semibold text-[#15365a]">
                      {post.caption || "Location post"}
                    </p>
                    <p className="text-xs font-medium text-[#6682a1]">
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
                  <Button asChild size="sm" variant="outline">
                    <Link href={ROUTES.upload}>
                      <Upload className="size-4" />
                      Upload
                    </Link>
                  </Button>
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
                  <Button asChild size="sm" variant="outline">
                    <Link href={ROUTES.upload}>
                      <Upload className="size-4" />
                      Upload
                    </Link>
                  </Button>
                </div>
              ) : null}
            </div>
          </section>
        </aside>
      </section>
    </PageShell>
  );
}
