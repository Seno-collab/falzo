export type Post = {
  id: number;
  user_id: number;
  user_name: string;
  image_url: string;
  caption: string;
  location_name: string;
  latitude: number;
  longitude: number;
  is_liked: boolean;
  is_saved: boolean;
  created_at: string;
};

export type UploadedImage = {
  id: number;
  url: string;
  object_key: string;
};

export type PostComment = {
  id: number;
  post_id: number;
  user_id: number;
  user_name: string;
  content: string;
  created_at: string;
};

export type PostCommentCreatedEvent = PostComment;

export type PostCreatedEvent = Post;

export type CreatePostPayload = {
  image_url: string;
  caption: string;
  location_name: string;
  latitude: number;
  longitude: number;
};
