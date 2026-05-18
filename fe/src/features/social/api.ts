import { initializeAuthHeader } from "@/features/auth/api";
import { USERS_ENDPOINT } from "@/lib/api-config";
import {
  apiDelete,
  apiGet,
  apiPost,
  endpointPath,
} from "@/lib/api-utils";
import type { PublicProfile } from "./types";

export async function getPublicProfileApi(userId: number) {
  initializeAuthHeader();
  return apiGet<PublicProfile>(endpointPath(USERS_ENDPOINT, userId));
}

export async function followUserApi(userId: number) {
  initializeAuthHeader();
  await apiPost(endpointPath(USERS_ENDPOINT, userId, "follow"));
}

export async function unfollowUserApi(userId: number) {
  initializeAuthHeader();
  await apiDelete(endpointPath(USERS_ENDPOINT, userId, "follow"));
}

export async function blockUserApi(userId: number) {
  initializeAuthHeader();
  await apiPost(endpointPath(USERS_ENDPOINT, userId, "block"));
}

export async function unblockUserApi(userId: number) {
  initializeAuthHeader();
  await apiDelete(endpointPath(USERS_ENDPOINT, userId, "block"));
}
