export type TravelPost = {
  id: string;
  imageUrl: string;
  caption: string;
  locationId: string;
  locationName: string;
  locationSubtitle?: string;
  tags: string[];
  likes: number;
  saves: number;
  createdAt?: string;
};

export type TravelLocation = {
  id: string;
  name: string;
  subtitle?: string;
  lat: number;
  lng: number;
  imageUrl?: string;
  postsCount?: number;
  countryCode?: string;
};

export type PaginatedPosts = {
  data: TravelPost[];
  nextPage: number | null;
};

export type NearbyLocationsParams = {
  lat: number;
  lng: number;
  radiusKm?: number;
};

export type CreatePostPayload = {
  caption: string;
  imageFile: File;
  locationId: string;
  tags: string[];
};
