import { http } from "@/lib/http";
import type { ApiEnvelope } from "@/types/api/response";
import type {
  CreatePostPayload,
  Post,
  PostComment,
  UploadedImage,
} from "./types";

const POSTS_ENDPOINT =
  process.env.NEXT_PUBLIC_POSTS_ENDPOINT ??
  process.env.VITE_POSTS_ENDPOINT ??
  "/posts/";

const IMAGE_UPLOAD_ENDPOINT =
  process.env.NEXT_PUBLIC_IMAGE_UPLOAD_ENDPOINT ??
  process.env.VITE_IMAGE_UPLOAD_ENDPOINT ??
  "/images/upload";

function unwrapResponseData<T>(value: ApiEnvelope<T> | T): T {
  if (
    value &&
    typeof value === "object" &&
    !Array.isArray(value) &&
    "success" in value &&
    "data" in value
  ) {
    return (value as ApiEnvelope<T>).data;
  }

  return value as T;
}

export async function getPostsApi(params?: {
  page?: number;
  limit?: number;
}): Promise<Post[]> {
  const response = await http.get<ApiEnvelope<Post[]> | Post[]>(POSTS_ENDPOINT, {
    params: {
      page: params?.page ?? 1,
      limit: params?.limit ?? 24,
    },
  });

  return unwrapResponseData(response.data);
}

export async function getPostDetailApi(postId: number): Promise<Post> {
  const response = await http.get<ApiEnvelope<Post> | Post>(
    `${POSTS_ENDPOINT.replace(/\/+$/, "")}/${postId}`,
  );

  return unwrapResponseData(response.data);
}

export async function uploadImageApi(file: File): Promise<UploadedImage> {
  const formData = new FormData();
  formData.append("file", file);

  const response = await http.post<ApiEnvelope<UploadedImage> | UploadedImage>(
    IMAGE_UPLOAD_ENDPOINT,
    formData,
  );

  return unwrapResponseData(response.data);
}

export async function createPostApi(
  payload: CreatePostPayload,
): Promise<Post> {
  const response = await http.post<ApiEnvelope<Post> | Post>(
    POSTS_ENDPOINT,
    payload,
  );

  return unwrapResponseData(response.data);
}

export async function likePostApi(postId: number): Promise<void> {
  await http.post(`${POSTS_ENDPOINT.replace(/\/+$/, "")}/${postId}/like`);
}

export async function savePostApi(postId: number): Promise<void> {
  await http.post(`${POSTS_ENDPOINT.replace(/\/+$/, "")}/${postId}/save`);
}

export async function getPostCommentsApi(postId: number): Promise<PostComment[]> {
  const response = await http.get<ApiEnvelope<PostComment[]> | PostComment[]>(
    `${POSTS_ENDPOINT.replace(/\/+$/, "")}/${postId}/comments`,
    {
      params: {
        page: 1,
        limit: 20,
      },
    },
  );

  return unwrapResponseData(response.data);
}

export async function createPostCommentApi(params: {
  postId: number;
  content: string;
}): Promise<PostComment> {
  const response = await http.post<ApiEnvelope<PostComment> | PostComment>(
    `${POSTS_ENDPOINT.replace(/\/+$/, "")}/${params.postId}/comments`,
    {
      content: params.content,
    },
  );

  return unwrapResponseData(response.data);
}
