import { initializeAuthHeader } from "@/features/auth/api";
import {
  buildApiUrl,
  IMAGE_UPLOAD_ENDPOINT,
  POSTS_ENDPOINT,
} from "@/lib/api-config";
import {
  apiDelete,
  apiGet,
  apiPatch,
  apiPost,
  apiPut,
  endpointPath,
} from "@/lib/api-utils";
import type {
  CreatePostCommentPayload,
  CreatePostPayload,
  CreateSavedCollectionPayload,
  Post,
  PostComment,
  PostCommentCreatedEvent,
  PostCreatedEvent,
  PostDeletedEvent,
  PostSort,
  ReportContentPayload,
  SavedCollection,
  UpdatePostCommentPayload,
  UpdatePostPayload,
  UpdateSavedCollectionPayload,
  UploadedImage,
} from "./types";

export async function getPostsApi(params?: {
  page?: number;
  limit?: number;
  search?: string;
  categorySlug?: string;
  feed?: "following";
  sort?: PostSort;
  latitude?: number;
  longitude?: number;
  radiusMeters?: number;
}): Promise<Post[]> {
  initializeAuthHeader();

  const search = params?.search?.trim();
  const categorySlug = params?.categorySlug?.trim();
  return apiGet<Post[]>(POSTS_ENDPOINT, {
    params: {
      page: params?.page ?? 1,
      limit: params?.limit ?? 24,
      ...(search ? { search } : {}),
      ...(categorySlug ? { category_slug: categorySlug } : {}),
      ...(params?.feed ? { feed: params.feed } : {}),
      ...(params?.sort ? { sort: params.sort } : {}),
      ...(typeof params?.latitude === "number" ? { lat: params.latitude } : {}),
      ...(typeof params?.longitude === "number" ? { lng: params.longitude } : {}),
      ...(params?.radiusMeters ? { radius_meters: params.radiusMeters } : {}),
    },
  });
}

export async function getPostDetailApi(postId: number): Promise<Post> {
  initializeAuthHeader();

  return apiGet<Post>(endpointPath(POSTS_ENDPOINT, postId));
}

export function getPostEventsUrl() {
  return buildApiUrl(endpointPath(POSTS_ENDPOINT, "events"));
}

export async function uploadImageApi(file: File): Promise<UploadedImage> {
  initializeAuthHeader();

  const formData = new FormData();
  formData.append("file", file);

  return apiPost<UploadedImage>(IMAGE_UPLOAD_ENDPOINT, formData);
}

export async function createPostApi(payload: CreatePostPayload): Promise<Post> {
  initializeAuthHeader();

  return apiPost<Post>(POSTS_ENDPOINT, payload);
}

export async function updatePostApi(payload: UpdatePostPayload): Promise<Post> {
  initializeAuthHeader();

  const { postId, ...body } = payload;
  return apiPut<Post>(endpointPath(POSTS_ENDPOINT, postId), body);
}

export async function deletePostApi(postId: number): Promise<void> {
  initializeAuthHeader();

  await apiDelete(endpointPath(POSTS_ENDPOINT, postId));
}

export async function hidePostApi(
  postId: number,
  payload: ReportContentPayload,
): Promise<void> {
  initializeAuthHeader();

  await apiPost(endpointPath(POSTS_ENDPOINT, postId, "hide"), payload);
}

export async function reportPostApi(
  postId: number,
  payload: ReportContentPayload,
): Promise<void> {
  initializeAuthHeader();

  await apiPost(endpointPath(POSTS_ENDPOINT, postId, "report"), payload);
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

export async function getSavedPostsApi(): Promise<Post[]> {
  initializeAuthHeader();

  return apiGet<Post[]>(endpointPath(POSTS_ENDPOINT, "saved"));
}

export async function getSavedCollectionsApi(): Promise<SavedCollection[]> {
  initializeAuthHeader();

  return apiGet<SavedCollection[]>(
    endpointPath(POSTS_ENDPOINT, "saved-collections"),
  );
}

export async function getPublicSavedCollectionApi(
  shareSlug: string,
): Promise<SavedCollection> {
  initializeAuthHeader();

  return apiGet<SavedCollection>(
    endpointPath(POSTS_ENDPOINT, "saved-collections", "public", shareSlug),
  );
}

export async function createSavedCollectionApi(
  payload: CreateSavedCollectionPayload,
): Promise<SavedCollection> {
  initializeAuthHeader();

  return apiPost<SavedCollection>(
    endpointPath(POSTS_ENDPOINT, "saved-collections"),
    payload,
  );
}

export async function updateSavedCollectionApi(
  payload: UpdateSavedCollectionPayload,
): Promise<SavedCollection> {
  initializeAuthHeader();

  return apiPatch<SavedCollection>(
    endpointPath(POSTS_ENDPOINT, "saved-collections", payload.collectionId),
    { is_public: payload.isPublic },
  );
}

export async function addPostToSavedCollectionApi(
  collectionId: number,
  postId: number,
): Promise<void> {
  initializeAuthHeader();

  await apiPost(
    endpointPath(
      POSTS_ENDPOINT,
      "saved-collections",
      collectionId,
      "posts",
      postId,
    ),
  );
}

export async function removePostFromSavedCollectionApi(
  collectionId: number,
  postId: number,
): Promise<void> {
  initializeAuthHeader();

  await apiDelete(
    endpointPath(
      POSTS_ENDPOINT,
      "saved-collections",
      collectionId,
      "posts",
      postId,
    ),
  );
}

export async function deleteSavedCollectionApi(
  collectionId: number,
): Promise<void> {
  initializeAuthHeader();

  await apiDelete(endpointPath(POSTS_ENDPOINT, "saved-collections", collectionId));
}

export async function getPostCommentsApi(
  postId: number,
): Promise<PostComment[]> {
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
  return buildApiUrl(
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
      typeof payload.created_at !== "string" ||
      typeof payload.updated_at !== "string"
    ) {
      return null;
    }

    return {
      ...payload,
      reply_to_comment_id:
        typeof payload.reply_to_comment_id === "number"
          ? payload.reply_to_comment_id
          : undefined,
      reply_to_user_id:
        typeof payload.reply_to_user_id === "number"
          ? payload.reply_to_user_id
          : undefined,
      reply_to_user_name:
        typeof payload.reply_to_user_name === "string"
          ? payload.reply_to_user_name
          : undefined,
      reply_to_content:
        typeof payload.reply_to_content === "string"
          ? payload.reply_to_content
          : undefined,
    } as PostCommentCreatedEvent;
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
      category_id:
        typeof payload.category_id === "number"
          ? payload.category_id
          : undefined,
      category_name:
        typeof payload.category_name === "string"
          ? payload.category_name
          : undefined,
      category_slug:
        typeof payload.category_slug === "string"
          ? payload.category_slug
          : undefined,
      is_liked: Boolean(payload.is_liked),
      is_saved: Boolean(payload.is_saved),
    } as PostCreatedEvent;
  } catch {
    return null;
  }
}

export function parsePostDeletedEvent(
  event: MessageEvent<string>,
): PostDeletedEvent | null {
  try {
    const payload = JSON.parse(event.data) as Partial<PostDeletedEvent>;
    if (typeof payload.id !== "number" || payload.id <= 0) {
      return null;
    }

    return { id: payload.id };
  } catch {
    return null;
  }
}

export async function createPostCommentApi(
  params: CreatePostCommentPayload,
): Promise<PostComment> {
  initializeAuthHeader();
  const payload: { content: string; reply_to_comment_id?: number } = {
    content: params.content,
  };
  if (params.replyToCommentId) {
    payload.reply_to_comment_id = params.replyToCommentId;
  }

  return apiPost<PostComment>(
    endpointPath(POSTS_ENDPOINT, params.postId, "comments"),
    payload,
  );
}

export async function updatePostCommentApi(
  params: UpdatePostCommentPayload,
): Promise<PostComment> {
  initializeAuthHeader();

  return apiPut<PostComment>(
    endpointPath(POSTS_ENDPOINT, params.postId, "comments", params.commentId),
    {
      content: params.content,
    },
  );
}

export async function deletePostCommentApi(
  postId: number,
  commentId: number,
): Promise<void> {
  initializeAuthHeader();

  await apiDelete(endpointPath(POSTS_ENDPOINT, postId, "comments", commentId));
}

export async function hidePostCommentApi(
  postId: number,
  commentId: number,
  payload: ReportContentPayload,
): Promise<void> {
  initializeAuthHeader();

  await apiPost(
    endpointPath(POSTS_ENDPOINT, postId, "comments", commentId, "hide"),
    payload,
  );
}

export async function reportPostCommentApi(
  postId: number,
  commentId: number,
  payload: ReportContentPayload,
): Promise<void> {
  initializeAuthHeader();

  await apiPost(
    endpointPath(POSTS_ENDPOINT, postId, "comments", commentId, "report"),
    payload,
  );
}
