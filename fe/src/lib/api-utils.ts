import type { AxiosRequestConfig } from "axios";
import { http } from "@/lib/http";
import type { ApiEnvelope } from "@/types/api/response";

export {
  buildApiUrl,
  endpointPath,
  envEndpoint,
  trimTrailingSlashes,
} from "@/lib/api-config";

export function unwrapApiData<T>(value: ApiEnvelope<T> | T): T {
  if (
    value &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    "success" in value &&
    "data" in value
  ) {
    return value.data;
  }

  return value;
}

export async function apiGet<T>(url: string, config?: AxiosRequestConfig) {
  const response = await http.get<ApiEnvelope<T> | T>(url, config);
  return unwrapApiData(response.data);
}

export async function apiPost<T>(
  url: string,
  data?: unknown,
  config?: AxiosRequestConfig,
) {
  const response = await http.post<ApiEnvelope<T> | T>(url, data, config);
  return unwrapApiData(response.data);
}

export async function apiPut<T>(
  url: string,
  data?: unknown,
  config?: AxiosRequestConfig,
) {
  const response = await http.put<ApiEnvelope<T> | T>(url, data, config);
  return unwrapApiData(response.data);
}

export async function apiPatch<T>(
  url: string,
  data?: unknown,
  config?: AxiosRequestConfig,
) {
  const response = await http.patch<ApiEnvelope<T> | T>(url, data, config);
  return unwrapApiData(response.data);
}

export async function apiDelete<T>(url: string, config?: AxiosRequestConfig) {
  const response = await http.delete<ApiEnvelope<T> | T>(url, config);
  return unwrapApiData(response.data);
}
