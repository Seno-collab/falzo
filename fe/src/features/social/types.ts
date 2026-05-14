import type { Post } from "@/features/posts/types";

export type PublicProfile = {
  user_id: number;
  user_name: string;
  full_name?: string;
  created_at: string;
  posts_count: number;
  followers_count: number;
  following_count: number;
  is_following: boolean;
  posts: Post[];
};
