import { apiGet, envEndpoint } from "@/lib/api-utils";
import type { Category } from "./types";

const CATEGORIES_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_CATEGORIES_ENDPOINT,
  process.env.VITE_CATEGORIES_ENDPOINT,
  "/categories/",
);

export async function getCategoriesApi(): Promise<Category[]> {
  return apiGet<Category[]>(CATEGORIES_ENDPOINT);
}
