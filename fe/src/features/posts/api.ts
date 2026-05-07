import { initializeAuthHeader } from "@/features/auth/api";
import {
  apiDelete,
  apiGet,
  apiPost,
  endpointPath,
  envEndpoint,
} from "@/lib/api-utils";
import type {
  CreatePostPayload,
  Post,
  PostComment,
  PostCommentCreatedEvent,
  PostCreatedEvent,
  UploadedImage,
} from "./types";

const API_BASE_URL = (
  process.env.NEXT_PUBLIC_API_BASE_URL ??
  process.env.VITE_API_BASE_URL ??
  "/api"
).trim();
const POSTS_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_POSTS_ENDPOINT,
  process.env.VITE_POSTS_ENDPOINT,
  "/posts/",
);
const IMAGE_UPLOAD_ENDPOINT = envEndpoint(
  process.env.NEXT_PUBLIC_IMAGE_UPLOAD_ENDPOINT,
  process.env.VITE_IMAGE_UPLOAD_ENDPOINT,
  "/images/upload",
);

function buildEventSourceUrl(path: string) {
  if (/^https?:\/\//i.test(path)) {
    return path;
  }

  const normalizedBase = API_BASE_URL.replace(/\/+$/, "");
  const normalizedPath = path.replace(/^\/+/, "");
  if (!normalizedBase || normalizedBase === "/") {
    return `/${normalizedPath}`;
  }

  return `${normalizedBase}/${normalizedPath}`;
}

export async function getPostsApi(params?: {
  page?: number;
  limit?: number;
}): Promise<Post[]> {
  initializeAuthHeader();

  return apiGet<Post[]>(POSTS_ENDPOINT, {
    params: {
      page: params?.page ?? 1,
      limit: params?.limit ?? 24,
    },
  });
}

export async function getPostDetailApi(postId: number): Promise<Post> {
  initializeAuthHeader();

  return apiGet<Post>(endpointPath(POSTS_ENDPOINT, postId));
}

export function getPostEventsUrl() {
  return buildEventSourceUrl(endpointPath(POSTS_ENDPOINT, "events"));
}

export async function uploadImageApi(file: File): Promise<UploadedImage> {
  initializeAuthHeader();

  const formData = new FormData();
  formData.append("file", file);

  return apiPost<UploadedImage>(IMAGE_UPLOAD_ENDPOINT, formData);
}

export async function createPostApi(
  payload: CreatePostPayload,
): Promise<Post> {
  initializeAuthHeader();

  return apiPost<Post>(POSTS_ENDPOINT, payload);
}

export async function likePostApi(postId: number): Promise<void> {
  initializeAuthHeader();

  await apiPost(endpointPath(POSTS_ENDPOINT, postId, "like"));
}

export async function unlikePostApi(postId: number): Promise<void> {
  initializeAuthHeader();

  await apiDelete(endpointPath(POSTS_ENDPOINT, postId, "like"));
}

export async function savePostApi(postId: number): Promise<void> {
  initializeAuthHeader();

  await apiPost(endpointPath(POSTS_ENDPOINT, postId, "save"));
}

export async function unsavePostApi(postId: number): Promise<void> {
  initializeAuthHeader();

  await apiDelete(endpointPath(POSTS_ENDPOINT, postId, "save"));
}

export async function getPostCommentsApi(postId: number): Promise<PostComment[]> {
  return apiGet<PostComment[]>(
    endpointPath(POSTS_ENDPOINT, postId, "comments"),
    {
      params: {
        page: 1,
        limit: 20,
      },
    },
  );
}

export function getPostCommentEventsUrl(postId: number) {
  return buildEventSourceUrl(
    endpointPath(POSTS_ENDPOINT, postId, "comments", "events"),
  );
}

export function parsePostCommentCreatedEvent(
  event: MessageEvent<string>,
): PostCommentCreatedEvent | null {
  try {
    const payload = JSON.parse(event.data) as Partial<PostCommentCreatedEvent>;
    if (
      typeof payload.id !== "number" ||
      typeof payload.post_id !== "number" ||
      typeof payload.user_id !== "number" ||
      typeof payload.user_name !== "string" ||
      typeof payload.content !== "string" ||
      typeof payload.created_at !== "string"
    ) {
      return null;
    }

    return payload as PostCommentCreatedEvent;
  } catch {
    return null;
  }
}

export function parsePostCreatedEvent(
  event: MessageEvent<string>,
): PostCreatedEvent | null {
  try {
    const payload = JSON.parse(event.data) as Partial<PostCreatedEvent>;
    if (
      typeof payload.id !== "number" ||
      typeof payload.user_id !== "number" ||
      typeof payload.user_name !== "string" ||
      typeof payload.image_url !== "string" ||
      typeof payload.caption !== "string" ||
      typeof payload.location_name !== "string" ||
      typeof payload.latitude !== "number" ||
      typeof payload.longitude !== "number" ||
      typeof payload.created_at !== "string"
    ) {
      return null;
    }

    return {
      ...payload,
      is_liked: Boolean(payload.is_liked),
      is_saved: Boolean(payload.is_saved),
    } as PostCreatedEvent;
  } catch {
    return null;
  }
}

export async function createPostCommentApi(params: {
  postId: number;
  content: string;
}): Promise<PostComment> {
  initializeAuthHeader();

  return apiPost<PostComment>(
    endpointPath(POSTS_ENDPOINT, params.postId, "comments"),
    {
      content: params.content,
    },
  );
}
