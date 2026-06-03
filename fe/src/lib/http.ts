import axios from "axios";
import { defaultLocale, readCurrentLocale } from "@/i18n/locale";
import { API_BASE_URL } from "@/lib/api-config";

export const http = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    Accept: "application/json",
  },
  timeout: 15_000,
});

http.interceptors.request.use((config) => {
  if (globalThis.window === undefined) {
    config.headers["X-Locale"] = defaultLocale;
    return config;
  }

  config.headers["X-Locale"] = readCurrentLocale();
  return config;
});
