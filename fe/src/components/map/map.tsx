"use client";

import { divIcon, DomEvent, latLngBounds } from "leaflet";
import { useEffect, useMemo, useState } from "react";
import {
  CircleMarker,
  MapContainer,
  Marker,
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

function escapeHtml(value: string) {
  return value.replaceAll(
    /[&<>"']/g,
    (character) =>
      ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': "&quot;",
        "'": "&#39;",
      })[character] ?? character,
  );
}

function photoMarkerHtml(point: MapPoint, selected: boolean) {
  const markerClass = selected
    ? "falzo-map-marker is-selected"
    : "falzo-map-marker";
  const countBadge =
    point.count && point.count > 1
      ? `<span class="falzo-photo-marker-count">${point.count}</span>`
      : "";

  return `<div class="${markerClass}"><span class="falzo-photo-marker"><img alt="" src="${escapeHtml(
    point.imageUrl ?? "",
  )}" />${countBadge}</span><span class="falzo-map-label">${escapeHtml(
    point.name,
  )}</span></div>`;
}

function placeMarkerHtml(point: MapPoint, selected: boolean) {
  const markerClass = selected
    ? "falzo-map-marker is-selected"
    : "falzo-map-marker";

  return `<div class="${markerClass}"><span class="falzo-place-pin"><span></span></span><span class="falzo-map-label">${escapeHtml(
    point.name,
  )}</span></div>`;
}

function markerScale(zoom: number) {
  return Math.min(1.08, Math.max(0.52, 0.52 + (zoom - 5) * 0.07));
}

function ZoomObserver({
  onZoomChange,
}: Readonly<{
  onZoomChange: (zoom: number) => void;
}>) {
  const map = useMapEvents({
    zoom: (event) => onZoomChange(event.target.getZoom()),
    zoomend: (event) => onZoomChange(event.target.getZoom()),
  });

  useEffect(() => {
    onZoomChange(map.getZoom());
  }, [map, onZoomChange]);

  return null;
}

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

    if (currentPosition) {
      if (points.length === 0) {
        map.flyTo(coordinatesToLatLng(currentPosition), zoom, {
          duration: 0.65,
        });
      }
      return;
    }

    const positions = points.map(coordinatesToLatLng);

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
  label,
}: Readonly<{
  currentPosition?: Coordinates | null;
  label?: string;
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
        <strong>{label ?? "Selected area"}</strong>
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
  markerZoom,
  onSelect,
  point,
  selected,
}: Readonly<{
  markerZoom: number;
  onSelect?: (point: MapPoint) => void;
  point: MapPoint;
  selected: boolean;
}>) {
  const distance = formatDistance(point.distanceMeters);
  const scale = markerScale(markerZoom) * (selected ? 1.08 : 1);

  if (point.imageUrl) {
    const icon = divIcon({
      className: "falzo-photo-marker-shell",
      html: `<div style="--falzo-marker-scale:${scale.toFixed(
        3,
      )}">${photoMarkerHtml(point, selected)}</div>`,
      iconAnchor: [32, 32],
      iconSize: [128, 92],
      popupAnchor: [0, -32],
    });

    return (
      <Marker
        eventHandlers={{
          click: (event) => {
            DomEvent.stopPropagation(event.originalEvent);
            onSelect?.(point);
          },
        }}
        icon={icon}
        position={coordinatesToLatLng(point)}
      >
        <Popup>
          <div className={styles.popup}>
            <img
              alt={point.name}
              className={styles.popupPhoto}
              src={point.imageUrl}
            />
            <strong>{point.name}</strong>
            {point.address ? <span>{point.address}</span> : null}
            {distance ? <span>{distance}</span> : null}
          </div>
        </Popup>
      </Marker>
    );
  }

  const icon = divIcon({
    className: "falzo-place-marker-shell",
    html: `<div style="--falzo-marker-scale:${scale.toFixed(
      3,
    )}">${placeMarkerHtml(point, selected)}</div>`,
    iconAnchor: [64, 46],
    iconSize: [128, 76],
    popupAnchor: [0, -36],
  });

  return (
    <Marker
      eventHandlers={{
        click: (event) => {
          DomEvent.stopPropagation(event.originalEvent);
          onSelect?.(point);
        },
      }}
      icon={icon}
      position={coordinatesToLatLng(point)}
    >
      <Popup>
        <div className={styles.popup}>
          <strong>{point.name}</strong>
          {point.address ? <span>{point.address}</span> : null}
          {distance ? <span>{distance}</span> : null}
        </div>
      </Popup>
    </Marker>
  );
}

export default function FalzoMap({
  className,
  currentPosition,
  currentPositionLabel,
  height = "default",
  onSelectCoordinates,
  onSelectPoint,
  points = [],
  selectedPointId,
  zoom = 12,
}: Readonly<FalzoMapProps>) {
  const [markerZoom, setMarkerZoom] = useState(zoom);
  const selectedPoint = useMemo(
    () => points.find((point) => point.id === selectedPointId),
    [points, selectedPointId],
  );
  const center = getInitialCenter({ currentPosition, points, selectedPoint });

  return (
    <div
      className={`${styles.wrapper} ${
        height === "compact" ? styles.compact : ""
      } ${height === "large" ? styles.large : ""} ${className ?? ""}`}
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
        <ZoomObserver onZoomChange={setMarkerZoom} />
        <MapClickController onSelectCoordinates={onSelectCoordinates} />
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <CurrentPositionMarker
          currentPosition={currentPosition}
          label={currentPositionLabel}
        />
        {points.map((point) => (
          <LocationMarker
            key={point.id}
            markerZoom={markerZoom}
            onSelect={onSelectPoint}
            point={point}
            selected={point.id === selectedPointId}
          />
        ))}
      </MapContainer>
    </div>
  );
}
