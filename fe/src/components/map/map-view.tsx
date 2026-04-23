"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { MapPinned } from "lucide-react";
import type { TravelLocation } from "@/lib/travel/types";
import { cn } from "@/lib/utils";

type MapViewProps = {
  locations: TravelLocation[];
  selectedLocationId?: string | null;
  onSelectLocation?: (locationId: string) => void;
  className?: string;
};

export function MapView({
  locations,
  selectedLocationId,
  onSelectLocation,
  className,
}: MapViewProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<unknown>(null);
  const markerRefs = useRef<Array<{ marker: { remove: () => void } }>>([]);
  const [mapboxReady, setMapboxReady] = useState(false);

  const token = process.env.NEXT_PUBLIC_MAPBOX_TOKEN;
  const validCoordinates = useMemo(
    () => locations.filter((location) => Number.isFinite(location.lat) && Number.isFinite(location.lng)),
    [locations],
  );

  useEffect(() => {
    if (!token || !containerRef.current || mapRef.current) {
      return;
    }

    let disposed = false;

    const setupMap = async () => {
      const mapboxModule = await import("mapbox-gl");
      if (disposed || !containerRef.current) {
        return;
      }

      mapboxModule.default.accessToken = token;
      const map = new mapboxModule.default.Map({
        container: containerRef.current,
        style: "mapbox://styles/mapbox/streets-v12",
        center: validCoordinates.length
          ? [validCoordinates[0].lng, validCoordinates[0].lat]
          : [106.700981, 10.77689],
        zoom: validCoordinates.length ? 4 : 2,
      });

      mapRef.current = map;
      map.addControl(new mapboxModule.default.NavigationControl({ showCompass: false }), "top-right");
      setMapboxReady(true);
    };

    void setupMap();

    return () => {
      disposed = true;
      markerRefs.current.forEach(({ marker }) => marker.remove());
      markerRefs.current = [];

      if (mapRef.current && typeof mapRef.current === "object" && "remove" in mapRef.current) {
        (mapRef.current as { remove: () => void }).remove();
      }

      mapRef.current = null;
      setMapboxReady(false);
    };
  }, [token, validCoordinates]);

  useEffect(() => {
    if (!mapboxReady || !mapRef.current || !token) {
      return;
    }

    let disposed = false;

    const drawMarkers = async () => {
      const mapboxModule = await import("mapbox-gl");
      if (disposed || !mapRef.current) {
        return;
      }

      markerRefs.current.forEach(({ marker }) => marker.remove());
      markerRefs.current = [];

      const map = mapRef.current as {
        fitBounds: (bounds: [[number, number], [number, number]], options?: { padding?: number }) => void;
      };

      const bounds = new mapboxModule.default.LngLatBounds();

      validCoordinates.forEach((location) => {
        const markerNode = document.createElement("button");
        markerNode.className = cn(
          "h-8 w-8 rounded-full border-2 border-white text-xs font-semibold shadow-md transition",
          selectedLocationId === location.id
            ? "bg-[#1f6fe5] text-white"
            : "bg-[#ffffffdd] text-[#1d3f67] hover:bg-[#eaf2ff]",
        );
        markerNode.textContent = "•";
        markerNode.type = "button";
        markerNode.onclick = () => {
          onSelectLocation?.(location.id);
        };

        const marker = new mapboxModule.default.Marker({ element: markerNode })
          .setLngLat([location.lng, location.lat])
          .addTo(mapRef.current as never);

        markerRefs.current.push({ marker });
        bounds.extend([location.lng, location.lat]);
      });

      if (!bounds.isEmpty()) {
        map.fitBounds(bounds.toArray() as [[number, number], [number, number]], { padding: 64 });
      }
    };

    void drawMarkers();

    return () => {
      disposed = true;
    };
  }, [mapboxReady, onSelectLocation, selectedLocationId, token, validCoordinates]);

  if (!token) {
    return (
      <div className={cn("surface relative overflow-hidden", className)}>
        <div className="absolute inset-0 bg-[linear-gradient(145deg,#eef5ff_0%,#e4f0ff_48%,#dbeaff_100%)]" />
        <div className="relative h-full min-h-[320px] p-4">
          <div className="mb-2 inline-flex items-center gap-1 rounded-full bg-white/80 px-2.5 py-1 text-xs font-semibold text-[#335c8a]">
            <MapPinned className="size-3.5" />
            Set NEXT_PUBLIC_MAPBOX_TOKEN to enable live map
          </div>
          <div className="grid gap-2">
            {validCoordinates.map((location) => (
              <button
                className={cn(
                  "flex items-center justify-between rounded-xl border border-[#cdddf0] bg-white/85 px-3 py-2 text-left text-sm font-semibold text-[#1c3e64]",
                  selectedLocationId === location.id ? "ring-2 ring-[#9cc0ee]" : "",
                )}
                key={location.id}
                onClick={() => onSelectLocation?.(location.id)}
                type="button"
              >
                <span className="truncate">{location.name}</span>
                <span className="text-xs text-[#6a86a7]">{location.subtitle}</span>
              </button>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={cn("surface overflow-hidden", className)}>
      <div className="h-full min-h-[340px]" ref={containerRef} />
    </div>
  );
}
