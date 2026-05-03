import axios from "axios";

const apiBaseUrl =
  process.env.NEXT_PUBLIC_API_BASE_URL ??
  process.env.VITE_API_BASE_URL ??
  "/api";

export const http = axios.create({
  baseURL: apiBaseUrl,
  headers: {
    Accept: "application/json",
  },
});
