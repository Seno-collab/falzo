import axios from "axios";
import { API_BASE_URL } from "@/lib/api-config";

export const http = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    Accept: "application/json",
  },
});
