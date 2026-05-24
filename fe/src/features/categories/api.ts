import type { AxiosRequestConfig } from "axios";
import { CATEGORIES_ENDPOINT } from "@/lib/api-config";
import { apiGet } from "@/lib/api-utils";
import type { Category } from "./types";

export async function getCategoriesApi(
  config?: AxiosRequestConfig,
): Promise<Category[]> {
  return apiGet<Category[]>(CATEGORIES_ENDPOINT, config);
}
