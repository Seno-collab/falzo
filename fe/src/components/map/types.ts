export type Coordinates = {
  latitude: number;
  longitude: number;
};

export type MapPoint = Coordinates & {
  id: string;
  name: string;
  address?: string;
  count?: number;
  distanceMeters?: number;
  imageUrl?: string;
  postIds?: number[];
};

export type FalzoMapProps = {
  className?: string;
  currentPosition?: Coordinates | null;
  currentPositionLabel?: string;
  height?: "default" | "compact" | "large";
  onSelectCoordinates?: (coordinates: Coordinates) => void;
  onSelectPoint?: (point: MapPoint) => void;
  points?: MapPoint[];
  selectedPointId?: string | null;
  zoom?: number;
};
