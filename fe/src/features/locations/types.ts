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
