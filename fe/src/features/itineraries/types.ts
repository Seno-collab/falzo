export type ItineraryListItem = {
  id: string;
  title: string;
  slug: string;
  province: string;
  durationDays: number;
  budgetMin: number;
  budgetMax: number;
  travelStyle: string;
  transportation: string;
  coverImageUrl: string;
  stopCount: number;
};

export type ItineraryListPage = {
  items: ItineraryListItem[];
  page: number;
  limit: number;
  total: number;
};

export type ItineraryListParams = {
  province?: string;
  durationDays?: number;
  budgetMax?: number;
  travelStyle?: string;
  page?: number;
  limit?: number;
};

export type ItineraryStop = {
  id: string;
  locationId: string;
  locationName: string;
  latitude: number;
  longitude: number;
  startTime: string;
  endTime: string;
  note: string;
  estimatedCost: number;
  stopOrder: number;
};

export type ItineraryDay = {
  dayNumber: number;
  stops: ItineraryStop[];
};

export type ItineraryDetail = {
  id: string;
  title: string;
  slug: string;
  province: string;
  durationDays: number;
  budgetMin: number;
  budgetMax: number;
  travelStyle: string;
  transportation: string;
  startLocation: string;
  description: string;
  coverImageUrl: string;
  days: ItineraryDay[];
};
