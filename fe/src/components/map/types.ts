export type Coordinates = {
  latitude: number;
  longitude: number;
};

export type MapPoint = Coordinates & {
  id: string;
  name: string;
  address?: string;
  distanceMeters?: number;
};

export type FalzoMapProps = {
  className?: string;
  currentPosition?: Coordinates | null;
  height?: "default" | "compact";
  onSelectCoordinates?: (coordinates: Coordinates) => void;
  onSelectPoint?: (point: MapPoint) => void;
  points?: MapPoint[];
  selectedPointId?: string | null;
  zoom?: number;
};
