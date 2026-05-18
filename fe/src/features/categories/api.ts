import { CATEGORIES_ENDPOINT } from "@/lib/api-config";
import { apiGet } from "@/lib/api-utils";
import type { Category } from "./types";

export async function getCategoriesApi(): Promise<Category[]> {
  return apiGet<Category[]>(CATEGORIES_ENDPOINT);
}
