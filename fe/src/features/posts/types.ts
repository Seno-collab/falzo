export type Post = {
  id: number;
  user_id: number;
  user_name: string;
  category_id?: number;
  category_name?: string;
  category_slug?: string;
  image_url: string;
  caption: string;
  location_name: string;
  latitude: number;
  longitude: number;
  is_liked: boolean;
  is_saved: boolean;
  status?: string;
  likes_count?: number;
  comments_count?: number;
  saves_count?: number;
  created_at: string;
};

export type SavedCollection = {
  id: number;
  user_id: number;
  name: string;
  share_slug: string;
  is_public: boolean;
  posts: Post[];
  post_count: number;
  created_at: string;
  updated_at: string;
  status?: string;
};

export type UploadedImage = {
  id: number;
  url: string;
  object_key: string;
};

export type CheckedImage = {
  valid: boolean;
  mime_type: string;
  size: number;
  width: number;
  height: number;
};

export type PostComment = {
  id: number;
  post_id: number;
  user_id: number;
  user_name: string;
  content: string;
  reply_to_comment_id?: number;
  reply_to_user_id?: number;
  reply_to_user_name?: string;
  reply_to_content?: string;
  created_at: string;
  updated_at: string;
};

export type PostCommentCreatedEvent = PostComment;

export type PostCreatedEvent = Post;

export type PostDeletedEvent = {
  id: number;
};

export type CreatePostPayload = {
  image_url: string;
  caption: string;
  category_id?: number;
  location_name: string;
  latitude: number;
  longitude: number;
};

export type UpdatePostPayload = CreatePostPayload & {
  postId: number;
};

export type CreatePostCommentPayload = {
  postId: number;
  content: string;
  replyToCommentId?: number;
};

export type UpdatePostCommentPayload = {
  postId: number;
  commentId: number;
  content: string;
};

export type CreateSavedCollectionPayload = {
  name: string;
  is_public?: boolean;
};

export type UpdateSavedCollectionPayload = {
  collectionId: number;
  isPublic: boolean;
};

export type ReportContentPayload = {
  reason: string;
};

export type PostSort = "newest" | "popular" | "trending" | "nearby";
