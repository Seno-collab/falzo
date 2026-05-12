"use client";

import { DomEvent, latLngBounds } from "leaflet";
import { useEffect, useMemo } from "react";
import {
  CircleMarker,
  MapContainer,
  Popup,
  TileLayer,
  useMap as useLeafletMap,
  useMapEvents,
} from "react-leaflet";
import {
  coordinatesToLatLng,
  formatDistance,
  getInitialCenter,
} from "./map-utils";
import styles from "./map.module.css";
import type { Coordinates, FalzoMapProps, MapPoint } from "./types";

function ViewportController({
  currentPosition,
  points,
  selectedPoint,
  zoom,
}: Readonly<{
  currentPosition?: Coordinates | null;
  points: MapPoint[];
  selectedPoint?: MapPoint;
  zoom: number;
}>) {
  const map = useLeafletMap();

  useEffect(() => {
    if (selectedPoint) {
      map.flyTo(coordinatesToLatLng(selectedPoint), Math.max(zoom, 14), {
        duration: 0.65,
      });
      return;
    }

    const positions = points.map(coordinatesToLatLng);

    if (currentPosition) {
      positions.push(coordinatesToLatLng(currentPosition));
    }

    if (positions.length > 1) {
      map.fitBounds(latLngBounds(positions), {
        animate: true,
        maxZoom: 14,
        padding: [36, 36],
      });
      return;
    }

    if (positions.length === 1) {
      map.flyTo(positions[0], zoom, { duration: 0.65 });
    }
  }, [currentPosition, map, points, selectedPoint, zoom]);

  return null;
}

function CurrentPositionMarker({
  currentPosition,
}: Readonly<{
  currentPosition?: Coordinates | null;
}>) {
  if (!currentPosition) {
    return null;
  }

  return (
    <CircleMarker
      center={coordinatesToLatLng(currentPosition)}
      className={styles.currentMarker}
      pathOptions={{
        color: "#2f6fb8",
        fillColor: "#2f6fb8",
        fillOpacity: 0.22,
        weight: 2,
      }}
      radius={14}
    >
      <Popup>
        <strong>Your position</strong>
      </Popup>
    </CircleMarker>
  );
}

function MapClickController({
  onSelectCoordinates,
}: Readonly<{
  onSelectCoordinates?: (coordinates: Coordinates) => void;
}>) {
  useMapEvents({
    click: (event) => {
      onSelectCoordinates?.({
        latitude: event.latlng.lat,
        longitude: event.latlng.lng,
      });
    },
  });

  return null;
}

function LocationMarker({
  onSelect,
  point,
  selected,
}: Readonly<{
  onSelect?: (point: MapPoint) => void;
  point: MapPoint;
  selected: boolean;
}>) {
  const distance = formatDistance(point.distanceMeters);

  return (
    <CircleMarker
      center={coordinatesToLatLng(point)}
      eventHandlers={{
        click: (event) => {
          DomEvent.stopPropagation(event.originalEvent);
          onSelect?.(point);
        },
      }}
      pathOptions={{
        color: selected ? "#d97706" : "#2f6fb8",
        fillColor: selected ? "#f59e0b" : "#4f8ec8",
        fillOpacity: selected ? 0.9 : 0.74,
        weight: selected ? 4 : 2,
      }}
      radius={selected ? 11 : 8}
    >
      <Popup>
        <div className={styles.popup}>
          <strong>{point.name}</strong>
          {point.address ? <span>{point.address}</span> : null}
          {distance ? <span>{distance}</span> : null}
        </div>
      </Popup>
    </CircleMarker>
  );
}

export default function FalzoMap({
  className,
  currentPosition,
  height = "default",
  onSelectCoordinates,
  onSelectPoint,
  points = [],
  selectedPointId,
  zoom = 12,
}: Readonly<FalzoMapProps>) {
  const selectedPoint = useMemo(
    () => points.find((point) => point.id === selectedPointId),
    [points, selectedPointId],
  );
  const center = getInitialCenter({ currentPosition, points, selectedPoint });

  return (
    <div
      className={`${styles.wrapper} ${
        height === "compact" ? styles.compact : ""
      } ${className ?? ""}`}
    >
      <MapContainer
        center={center}
        className={styles.map}
        scrollWheelZoom={false}
        zoom={zoom}
      >
        <ViewportController
          currentPosition={currentPosition}
          points={points}
          selectedPoint={selectedPoint}
          zoom={zoom}
        />
        <MapClickController onSelectCoordinates={onSelectCoordinates} />
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <CurrentPositionMarker currentPosition={currentPosition} />
        {points.map((point) => (
          <LocationMarker
            key={point.id}
            onSelect={onSelectPoint}
            point={point}
            selected={point.id === selectedPointId}
          />
        ))}
      </MapContainer>
    </div>
  );
}
