import type { AxiosRequestConfig } from "axios";
import { ITINERARIES_ENDPOINT } from "@/lib/api-config";
import { apiGet, endpointPath } from "@/lib/api-utils";
import type {
  ItineraryDetail,
  ItineraryListPage,
  ItineraryListParams,
} from "@/features/itineraries/types";

function cleanParams(params: ItineraryListParams) {
  return Object.fromEntries(
    Object.entries(params).filter(([, value]) => {
      if (typeof value === "string") {
        return value.trim().length > 0;
      }
      return value !== undefined && value !== null && value !== 0;
    }),
  );
}

export function getItinerariesApi(
  params: ItineraryListParams,
  config?: AxiosRequestConfig,
): Promise<ItineraryListPage> {
  return apiGet<ItineraryListPage>(ITINERARIES_ENDPOINT, {
    ...config,
    params: cleanParams(params),
  });
}

export function getItineraryBySlugApi(
  slug: string,
  config?: AxiosRequestConfig,
): Promise<ItineraryDetail> {
  return apiGet<ItineraryDetail>(endpointPath(ITINERARIES_ENDPOINT, slug), config);
}
