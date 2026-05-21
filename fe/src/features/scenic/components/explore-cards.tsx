"use client";

import {
  Bookmark,
  Flag,
  Heart,
  Maximize2,
  MessageCircle,
  Pencil,
  Reply,
  Send,
  Trash2,
  X,
} from "lucide-react";
import { motion } from "motion/react";
import Link from "next/link";
import type { KeyboardEvent } from "react";
import { Button } from "@/components/ui/button";
import type { Post, PostComment } from "@/features/posts/types";
import { ROUTES } from "@/lib/routes";
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
  replyTarget: PostComment | null;
  editingComment: PostComment | null;
  currentUserId: number | null;
  isSubmittingComment: boolean;
  onOpen: (postId: number) => void;
  onLike: (postId: number) => void;
  onSave: (postId: number) => void;
  onDelete: (postId: number) => void;
  onReport: (postId: number) => void;
  onToggleComments: (postId: number) => void;
  onCommentChange: (postId: number, value: string) => void;
  onCancelReply: (postId: number) => void;
  onCancelEdit: (postId: number) => void;
  onEditComment: (postId: number, comment: PostComment) => void;
  onReplyComment: (postId: number, comment: PostComment) => void;
  onSubmitComment: (postId: number) => void;
  onRegisterCommentInput: (
    postId: number,
    node: HTMLInputElement | null,
  ) => void;
};

const cardClass =
  "group mb-4 break-inside-avoid overflow-hidden rounded-2xl border border-black/5 bg-white shadow-[0_18px_48px_-38px_rgb(0_0_0/0.6)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_26px_70px_-42px_rgb(0_0_0/0.72)] sm:rounded-[28px]";

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

function isCommentEdited(comment: PostComment) {
  return comment.updated_at !== comment.created_at;
}

function Comments({
  comments,
  isLoading,
  onReply,
  onEdit,
  selectedReplyTargetId,
  editingCommentId,
  currentUserId,
}: Readonly<{
  comments: PostComment[];
  isLoading: boolean;
  onReply: (comment: PostComment) => void;
  onEdit: (comment: PostComment) => void;
  selectedReplyTargetId: number | null;
  editingCommentId: number | null;
  currentUserId: number | null;
}>) {
  if (isLoading) {
    return (
      <p className="text-xs font-semibold text-[#555]">Loading comments</p>
    );
  }

  if (comments.length === 0) {
    return (
      <p className="text-xs font-semibold text-[#555]">No comments yet.</p>
    );
  }

  return comments.map((comment) => (
    <div
      className={cn(
        "rounded-xl border px-3 py-2 text-sm text-[#333]",
        editingCommentId === comment.id
          ? "border-[#dec078] bg-[#fff8e6]"
          : selectedReplyTargetId === comment.id
            ? "border-[#8ebae6] bg-[#f2f7fd]"
            : "border-transparent bg-white",
      )}
      key={comment.id}
    >
      <div className="flex items-center justify-between gap-2">
        <Link
          className="min-w-0 truncate text-xs font-semibold text-[#777] hover:text-[#ff385c]"
          href={ROUTES.userProfile(comment.user_id)}
        >
          {comment.user_name || `User #${comment.user_id}`}
        </Link>
        <div className="flex shrink-0 items-center gap-2">
          {currentUserId === comment.user_id ? (
            <button
              className="inline-flex items-center gap-1 text-xs font-semibold text-[#8c6a1f] transition hover:text-[#5e430f]"
              onClick={() => onEdit(comment)}
              type="button"
            >
              <Pencil className="size-3" />
              Edit
            </button>
          ) : null}
          <button
            className="inline-flex items-center gap-1 text-xs font-semibold text-[#5f7894] transition hover:text-[#244c78]"
            onClick={() => onReply(comment)}
            type="button"
          >
            <Reply className="size-3" />
            Reply
          </button>
        </div>
      </div>
      <div className="mt-1 flex items-center gap-2 text-[11px] font-medium text-[#999]">
        <span>{new Date(comment.created_at).toLocaleDateString()}</span>
        {isCommentEdited(comment) ? (
          <span className="rounded-full bg-[#f2f7fd] px-2 py-0.5 text-[#5f7894]">
            edited
          </span>
        ) : null}
      </div>
      {comment.reply_to_comment_id ? (
        <div className="mt-2 rounded-lg border-l-2 border-[#7aa7d9] bg-[#f2f7fd] px-2 py-1.5 text-xs text-[#4b6682]">
          <p className="font-semibold">
            {comment.reply_to_user_name ||
              `User #${comment.reply_to_user_id ?? ""}`}
          </p>
          <p className="mt-0.5 line-clamp-2">{comment.reply_to_content}</p>
        </div>
      ) : null}
      <p className="mt-1 leading-5">{comment.content}</p>
    </div>
  ));
}

function CommentInput({
  disabled,
  isAuthenticated,
  isSubmitting,
  replyTarget,
  editingComment,
  value,
  onCancelEdit,
  onCancelReply,
  onChange,
  onInputMount,
  onSubmit,
}: {
  disabled: boolean;
  isAuthenticated: boolean;
  isSubmitting: boolean;
  replyTarget: PostComment | null;
  editingComment: PostComment | null;
  value: string;
  onCancelEdit: () => void;
  onCancelReply: () => void;
  onChange: (value: string) => void;
  onInputMount: (node: HTMLInputElement | null) => void;
  onSubmit: () => void;
}) {
  const placeholder = !isAuthenticated
    ? "Login to comment"
    : editingComment
      ? "Edit comment"
      : replyTarget
        ? `Reply to ${replyTarget.user_name || `User #${replyTarget.user_id}`}`
        : "Write a comment";

  return (
    <div className="mt-3 space-y-2">
      {editingComment ? (
        <div className="flex items-start gap-2 rounded-2xl border border-[#dec078] bg-[#fff8e6] px-3 py-2 text-xs text-[#73551b]">
          <Pencil className="mt-0.5 size-3 shrink-0" />
          <div className="min-w-0 flex-1">
            <p className="font-semibold">Editing comment</p>
            <p className="mt-0.5 truncate">{editingComment.content}</p>
          </div>
          <button
            aria-label="Cancel edit"
            className="rounded-full p-0.5 text-[#8c6a1f] transition hover:bg-white hover:text-[#5e430f]"
            onClick={onCancelEdit}
            type="button"
          >
            <X className="size-3.5" />
          </button>
        </div>
      ) : replyTarget ? (
        <div className="flex items-start gap-2 rounded-2xl border border-[#c8ddf1] bg-[#f2f7fd] px-3 py-2 text-xs text-[#4b6682]">
          <Reply className="mt-0.5 size-3 shrink-0" />
          <div className="min-w-0 flex-1">
            <p className="font-semibold">
              Replying to{" "}
              {replyTarget.user_name || `User #${replyTarget.user_id}`}
            </p>
            <p className="mt-0.5 truncate">{replyTarget.content}</p>
          </div>
          <button
            aria-label="Cancel reply"
            className="rounded-full p-0.5 text-[#5f7894] transition hover:bg-white hover:text-[#244c78]"
            onClick={onCancelReply}
            type="button"
          >
            <X className="size-3.5" />
          </button>
        </div>
      ) : null}
      <div className="flex items-center gap-2">
        <input
          className="h-9 min-w-0 flex-1 rounded-full border border-black/8 bg-white px-3 text-sm outline-none placeholder:text-[#999] focus:border-[#111]/20"
          disabled={disabled}
          onChange={(event) => onChange(event.target.value)}
          placeholder={placeholder}
          ref={onInputMount}
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
  replyTarget,
  editingComment,
  currentUserId,
  isSubmittingComment,
  onOpen,
  onLike,
  onSave,
  onDelete,
  onReport,
  onToggleComments,
  onCommentChange,
  onCancelReply,
  onCancelEdit,
  onEditComment,
  onReplyComment,
  onSubmitComment,
  onRegisterCommentInput,
}: Readonly<PostCardProps>) {
  const title = post.caption || "Travel story";
  const location = post.location_name || "Destination";
  const authorName = post.user_name || `User #${post.user_id}`;
  const categoryLabel = post.category_name || "Travel";

  return (
    <motion.article
      className={cardClass}
      initial={{ opacity: 0, y: 18 }}
      transition={animation(index)}
      viewport={{ amount: 0.12, once: true }}
      whileInView={{ opacity: 1, y: 0 }}
    >
      <div
        className="relative h-72 cursor-zoom-in overflow-hidden bg-[#e9eef3] sm:h-96"
        onClick={() => onOpen(post.id)}
        onKeyDown={(event) => onKeyboardOpen(event, () => onOpen(post.id))}
        role="button"
        tabIndex={0}
      >
        <img
          alt={post.caption || post.location_name || "Destination photo"}
          className="h-full w-full object-cover transition duration-500 ease-out group-hover:scale-[1.035]"
          loading={index < 2 ? "eager" : "lazy"}
          src={post.image_url}
        />
        <div className={overlayClass} />
        <div className="absolute left-3 right-3 top-3 flex items-start justify-between gap-2 opacity-0 transition duration-200 group-hover:opacity-100">
          <span className="rounded-full bg-white/86 px-3 py-1 text-xs font-semibold text-[#222] shadow-sm backdrop-blur-xl">
            {categoryLabel}
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
            {isAuthenticated ? (
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
            ) : null}
          </div>
        </div>
        <div className="absolute inset-x-4 bottom-4 text-white">
          <p className="line-clamp-1 text-xs font-semibold uppercase tracking-[0.16em] text-white/76">
            {location}
          </p>
          <h2 className="mt-1 line-clamp-3 text-xl font-semibold leading-tight tracking-normal sm:text-2xl">
            {title}
          </h2>
        </div>
      </div>

      <div className="space-y-3 p-3.5 sm:p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="min-w-0">
            <Link
              className="block truncate text-sm font-semibold text-[#202020] hover:text-[#ff385c]"
              href={ROUTES.userProfile(post.user_id)}
            >
              {authorName}
            </Link>
            <p className="mt-0.5 text-xs font-medium text-[#777]">
              {categoryLabel} - {new Date(post.created_at).toLocaleDateString()}
            </p>
          </div>
          <div
            className={cn(
              "grid gap-1 sm:flex sm:shrink-0 sm:items-center",
              isAuthenticated ? "grid-cols-5" : "grid-cols-2",
            )}
          >
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
            {isAuthenticated ? (
              <Button
                aria-label={isLiked ? "Liked" : "Like travel post"}
                className={cn("rounded-full", activeClass("heart", isLiked))}
                onClick={() => onLike(post.id)}
                size="icon-sm"
                type="button"
                variant="outline"
              >
                <Heart className={cn("size-4", isLiked ? "fill-current" : "")} />
              </Button>
            ) : null}
            <Button
              aria-label="View comments"
              className={cn(
                "rounded-full",
                activeClass("comment", commentsOpen),
              )}
              onClick={() => onToggleComments(post.id)}
              size="icon-sm"
              type="button"
              variant="outline"
            >
              <MessageCircle className="size-4" />
            </Button>
            {isAuthenticated ? (
              <>
                <Button
                  aria-label={isSaved ? "Saved" : "Save destination"}
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
                {currentUserId === post.user_id ? (
                  <Button
                    aria-label="Delete travel post"
                    className="rounded-full text-[#b4233f]"
                    onClick={() => onDelete(post.id)}
                    size="icon-sm"
                    type="button"
                    variant="outline"
                  >
                    <Trash2 className="size-4" />
                  </Button>
                ) : (
                  <Button
                    aria-label="Report travel post"
                    className="rounded-full"
                    onClick={() => onReport(post.id)}
                    size="icon-sm"
                    type="button"
                    variant="outline"
                  >
                    <Flag className="size-4" />
                  </Button>
                )}
              </>
            ) : null}
          </div>
        </div>

        {commentsOpen ? (
          <div className="rounded-2xl border border-black/6 bg-[#f8f8f7] p-3">
            <div className="space-y-2">
              <Comments
                comments={comments}
                currentUserId={currentUserId}
                editingCommentId={editingComment?.id ?? null}
                isLoading={isLoadingComments}
                onEdit={(comment) => onEditComment(post.id, comment)}
                onReply={(comment) => onReplyComment(post.id, comment)}
                selectedReplyTargetId={replyTarget?.id ?? null}
              />
            </div>
            <CommentInput
              disabled={!isAuthenticated || isSubmittingComment}
              editingComment={editingComment?.id ? editingComment : null}
              isAuthenticated={isAuthenticated}
              isSubmitting={isSubmittingComment}
              onCancelEdit={() => onCancelEdit(post.id)}
              onCancelReply={() => onCancelReply(post.id)}
              onChange={(value) => onCommentChange(post.id, value)}
              onInputMount={(node) => onRegisterCommentInput(post.id, node)}
              onSubmit={() => onSubmitComment(post.id)}
              replyTarget={replyTarget}
              value={commentValue}
            />
          </div>
        ) : null}
      </div>
    </motion.article>
  );
}
