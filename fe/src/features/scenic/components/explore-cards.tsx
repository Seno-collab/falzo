"use client";

import {
  Bookmark,
  Heart,
  Maximize2,
  MessageCircle,
  Send,
} from "lucide-react";
import { motion } from "motion/react";
import type { KeyboardEvent } from "react";
import { Button } from "@/components/ui/button";
import type { Post, PostComment } from "@/features/posts/types";
import { cn } from "@/lib/utils";

type PostCardProps = {
  post: Post;
  index: number;
  isAuthenticated: boolean;
  isLiked: boolean;
  isSaved: boolean;
  commentsOpen: boolean;
  comments: PostComment[];
  isLoadingComments: boolean;
  commentValue: string;
  isSubmittingComment: boolean;
  onOpen: (postId: number) => void;
  onLike: (postId: number) => void;
  onSave: (postId: number) => void;
  onToggleComments: (postId: number) => void;
  onCommentChange: (postId: number, value: string) => void;
  onSubmitComment: (postId: number) => void;
};

const cardClass =
  "group mb-4 break-inside-avoid overflow-hidden rounded-[28px] border border-black/5 bg-white shadow-[0_18px_48px_-38px_rgb(0_0_0/0.6)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_26px_70px_-42px_rgb(0_0_0/0.72)]";

const overlayClass =
  "pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0.02)_0%,rgb(0_0_0/0.02)_48%,rgb(0_0_0/0.44)_100%)] opacity-80 transition group-hover:opacity-100";

function animation(index: number) {
  return {
    duration: 0.34,
    delay: Math.min(index * 0.035, 0.22),
    ease: "easeOut" as const,
  };
}

function onKeyboardOpen(event: KeyboardEvent, callback: () => void) {
  if (event.key !== "Enter" && event.key !== " ") {
    return;
  }

  event.preventDefault();
  callback();
}

function activeClass(type: "heart" | "save" | "comment", active: boolean) {
  if (!active) {
    return "";
  }

  if (type === "heart") {
    return "border-[#ffb3c1] bg-[#fff1f4] text-[#cf2142]";
  }

  return "border-[#c8ddf1] bg-[#f2f7fd] text-[#2f6fb8]";
}

function Comments({
  comments,
  isLoading,
}: {
  comments: PostComment[];
  isLoading: boolean;
}) {
  if (isLoading) {
    return <p className="text-xs font-semibold text-[#555]">Loading comments</p>;
  }

  if (comments.length === 0) {
    return <p className="text-xs font-semibold text-[#555]">No comments yet.</p>;
  }

  return comments.map((comment) => (
    <div
      className="rounded-xl bg-white px-3 py-2 text-sm text-[#333]"
      key={comment.id}
    >
      <p className="text-xs font-semibold text-[#777]">
        {comment.user_name || `User #${comment.user_id}`}
      </p>
      <p className="mt-1 leading-5">{comment.content}</p>
    </div>
  ));
}

function CommentInput({
  disabled,
  isAuthenticated,
  isSubmitting,
  value,
  onChange,
  onSubmit,
}: {
  disabled: boolean;
  isAuthenticated: boolean;
  isSubmitting: boolean;
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
}) {
  return (
    <div className="mt-3 flex items-center gap-2">
      <input
        className="h-9 min-w-0 flex-1 rounded-full border border-black/8 bg-white px-3 text-sm outline-none placeholder:text-[#999] focus:border-[#111]/20"
        disabled={disabled}
        onChange={(event) => onChange(event.target.value)}
        placeholder={isAuthenticated ? "Write a comment" : "Login to comment"}
        value={value}
      />
      <Button
        aria-label={isAuthenticated ? "Submit comment" : "Login to comment"}
        className="rounded-full"
        disabled={isSubmitting}
        onClick={onSubmit}
        size="icon-sm"
        type="button"
        variant="outline"
      >
        <Send className="size-4" />
      </Button>
    </div>
  );
}

export function ExplorePostCard({
  post,
  index,
  isAuthenticated,
  isLiked,
  isSaved,
  commentsOpen,
  comments,
  isLoadingComments,
  commentValue,
  isSubmittingComment,
  onOpen,
  onLike,
  onSave,
  onToggleComments,
  onCommentChange,
  onSubmitComment,
}: PostCardProps) {
  const title = post.caption || "Community post";
  const location = post.location_name || "Uploaded";
  const authorName = post.user_name || `User #${post.user_id}`;

  return (
    <motion.article
      className={cardClass}
      initial={{ opacity: 0, y: 18 }}
      transition={animation(index)}
      viewport={{ amount: 0.12, once: true }}
      whileInView={{ opacity: 1, y: 0 }}
    >
      <div
        className="relative h-96 cursor-zoom-in overflow-hidden bg-[#e9eef3]"
        onClick={() => onOpen(post.id)}
        onKeyDown={(event) => onKeyboardOpen(event, () => onOpen(post.id))}
        role="button"
        tabIndex={0}
      >
        <img
          alt={post.caption || post.location_name || "Uploaded post"}
          className="h-full w-full object-cover transition duration-500 ease-out group-hover:scale-[1.035]"
          loading={index < 2 ? "eager" : "lazy"}
          src={post.image_url}
        />
        <div className={overlayClass} />
        <div className="absolute left-3 right-3 top-3 flex items-start justify-between gap-2 opacity-0 transition duration-200 group-hover:opacity-100">
          <span className="rounded-full bg-white/86 px-3 py-1 text-xs font-semibold text-[#222] shadow-sm backdrop-blur-xl">
            Community
          </span>
          <div className="flex items-center gap-1">
            <Button
              aria-label="Open image"
              className="rounded-full bg-white/86 text-[#222] shadow-sm backdrop-blur-xl hover:bg-white"
              onClick={(event) => {
                event.stopPropagation();
                onOpen(post.id);
              }}
              size="icon-sm"
              type="button"
              variant="ghost"
            >
              <Maximize2 className="size-4" />
            </Button>
            <Button
              aria-label={isLiked ? "Liked" : "Like"}
              className={cn(
                "rounded-full shadow-sm backdrop-blur-xl",
                isLiked
                  ? "bg-[#ff385c] text-white hover:bg-[#e93152]"
                  : "bg-white/86 text-[#222] hover:bg-white",
              )}
              onClick={(event) => {
                event.stopPropagation();
                onLike(post.id);
              }}
              size="icon-sm"
              type="button"
              variant="ghost"
            >
              <Heart className={cn("size-4", isLiked ? "fill-current" : "")} />
            </Button>
          </div>
        </div>
        <div className="absolute inset-x-4 bottom-4 text-white">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-white/76">
            {location}
          </p>
          <h2 className="mt-1 text-2xl font-semibold tracking-normal">
            {title}
          </h2>
        </div>
      </div>

      <div className="space-y-3 p-4">
        <div className="flex items-center justify-between gap-3">
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold text-[#202020]">
              {authorName}
            </p>
            <p className="mt-0.5 text-xs font-medium text-[#777]">
              {new Date(post.created_at).toLocaleDateString()}
            </p>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            <Button
              aria-label="Open image"
              className="rounded-full"
              onClick={() => onOpen(post.id)}
              size="icon-sm"
              type="button"
              variant="outline"
            >
              <Maximize2 className="size-4" />
            </Button>
            <Button
              aria-label={isLiked ? "Liked" : "Like post"}
              className={cn("rounded-full", activeClass("heart", isLiked))}
              onClick={() => onLike(post.id)}
              size="icon-sm"
              type="button"
              variant="outline"
            >
              <Heart className={cn("size-4", isLiked ? "fill-current" : "")} />
            </Button>
            <Button
              aria-label="View comments"
              className={cn("rounded-full", activeClass("comment", commentsOpen))}
              onClick={() => onToggleComments(post.id)}
              size="icon-sm"
              type="button"
              variant="outline"
            >
              <MessageCircle className="size-4" />
            </Button>
            <Button
              aria-label={isSaved ? "Saved" : "Save post"}
              className={cn("rounded-full", activeClass("save", isSaved))}
              onClick={() => onSave(post.id)}
              size="icon-sm"
              type="button"
              variant="outline"
            >
              <Bookmark
                className={cn("size-4", isSaved ? "fill-current" : "")}
              />
            </Button>
          </div>
        </div>

        {commentsOpen ? (
          <div className="rounded-2xl border border-black/6 bg-[#f8f8f7] p-3">
            <div className="space-y-2">
              <Comments comments={comments} isLoading={isLoadingComments} />
            </div>
            <CommentInput
              disabled={!isAuthenticated || isSubmittingComment}
              isAuthenticated={isAuthenticated}
              isSubmitting={isSubmittingComment}
              onChange={(value) => onCommentChange(post.id, value)}
              onSubmit={() => onSubmitComment(post.id)}
              value={commentValue}
            />
          </div>
        ) : null}
      </div>
    </motion.article>
  );
}
