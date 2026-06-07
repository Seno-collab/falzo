export type Location = {
  id: string;
  name: string;
  address: string;
  latitude: number;
  longitude: number;
  post_ids?: number[];
  post_count?: number;
};

export type NearbyLocation = {
  location: Location;
  distance_meters: number;
};

export type LocationPost = {
  id: string;
  user_id: number;
  image_url: string;
  caption: string;
  location_name: string;
  latitude: number;
  longitude: number;
};

export type PlaceDetail = {
  id: string;
  name: string;
  slug: string;
  province: string;
  district: string;
  address: string;
  latitude: number;
  longitude: number;
  imageUrl?: string;
  description: string;
  bestTimeToVisit: string;
  estimatedCostMin: number;
  estimatedCostMax: number;
  travelStyles: string[];
  suitableFor: string[];
  warningNote: string;
  isHiddenGem: boolean;
  ratingReality: number | null;
  ratingPhoto: number | null;
};
