import { initializeAuthHeader } from "@/features/auth/api";
import { apiGet, apiPost, endpointPath, envEndpoint } from "@/lib/api-utils";
import type {
  CreatePostPayload,
  Post,
  PostComment,
  UploadedImage,
} from "./types";

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

export async function getPostsApi(params?: {
  page?: number;
  limit?: number;
}): Promise<Post[]> {
  return apiGet<Post[]>(POSTS_ENDPOINT, {
    params: {
      page: params?.page ?? 1,
      limit: params?.limit ?? 24,
    },
  });
}

export async function getPostDetailApi(postId: number): Promise<Post> {
  return apiGet<Post>(endpointPath(POSTS_ENDPOINT, postId));
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

export async function savePostApi(postId: number): Promise<void> {
  initializeAuthHeader();

  await apiPost(endpointPath(POSTS_ENDPOINT, postId, "save"));
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
