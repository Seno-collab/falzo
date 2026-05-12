export type AppNotificationType = "image.uploaded" | "post.commented" | "post.created";

export type AppNotification = {
  id: string;
  user_id?: number;
  actor_user_id?: number;
  actor_name?: string;
  type: AppNotificationType | string;
  title: string;
  body: string;
  resource?: string;
  resource_id?: string;
  post_id?: number;
  image_id?: number;
  created_at: string;
};
