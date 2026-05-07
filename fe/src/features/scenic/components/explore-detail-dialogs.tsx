"use client";

import { Bookmark, Heart, MessageCircle, Pencil, Reply, Send, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { Post, PostComment } from "@/features/posts/types";
import { cn } from "@/lib/utils";

type PostDetailDialogProps = {
  open: boolean;
  post: Post | null;
  comments: PostComment[];
  isAuthenticated: boolean;
  isChatOpen: boolean;
  isLoadingComments: boolean;
  isLiked: boolean;
  isSaved: boolean;
  isCommentPending: boolean;
  commentValue: string;
  replyTarget: PostComment | null;
  editingComment: PostComment | null;
  currentUserId: number | null;
  onClose: () => void;
  onLike: () => void;
  onSave: () => void;
  onLoadComments: () => void;
  onCancelReply: () => void;
  onCancelEdit: () => void;
  onCommentChange: (value: string) => void;
  onEditComment: (comment: PostComment) => void;
  onRegisterCommentInput: (node: HTMLInputElement | null) => void;
  onReplyComment: (comment: PostComment) => void;
  onSubmitComment: () => void;
};

function selectedClass(type: "heart" | "save", active: boolean) {
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

function DetailComments({
  comments,
  isChatOpen,
  isLoading,
  onReplyComment,
  selectedReplyTargetId,
  editingCommentId,
  currentUserId,
  onEditComment,
}: Readonly<{
  comments: PostComment[];
  isChatOpen: boolean;
  isLoading: boolean;
  onReplyComment: (comment: PostComment) => void;
  selectedReplyTargetId: number | null;
  editingCommentId: number | null;
  currentUserId: number | null;
  onEditComment: (comment: PostComment) => void;
}>) {
  if (isLoading) {
    return (
      <p className="text-sm font-semibold text-[#555]">Loading comments</p>
    );
  }

  if (!isChatOpen) {
    return (
      <div className="rounded-2xl border border-black/6 bg-white px-4 py-5 text-sm font-semibold text-[#555]">
        Open comments.
      </div>
    );
  }

  if (comments.length === 0) {
    return (
      <div className="rounded-2xl border border-black/6 bg-white px-4 py-5 text-sm font-semibold text-[#555]">
        No comments yet.
      </div>
    );
  }

  return comments.map((comment) => (
    <div
      className={cn(
        "rounded-2xl border px-4 py-3 text-sm text-[#333]",
        editingCommentId === comment.id
          ? "border-[#dec078] bg-[#fff8e6]"
          : selectedReplyTargetId === comment.id
          ? "border-[#8ebae6] bg-[#f2f7fd]"
          : "border-black/5 bg-white",
      )}
      key={comment.id}
    >
      <div className="flex items-center justify-between gap-3">
        <p className="min-w-0 truncate text-xs font-semibold text-[#777]">
          {comment.user_name || `User #${comment.user_id}`}
        </p>
        <div className="flex shrink-0 items-center gap-2">
          <p className="text-xs font-medium text-[#999]">
            {new Date(comment.created_at).toLocaleDateString()}
          </p>
          {isCommentEdited(comment) ? (
            <span className="rounded-full bg-[#f2f7fd] px-2 py-0.5 text-[11px] font-medium text-[#5f7894]">
              edited
            </span>
          ) : null}
          {currentUserId === comment.user_id ? (
            <button
              className="inline-flex items-center gap-1 text-xs font-semibold text-[#8c6a1f] transition hover:text-[#5e430f]"
              onClick={() => onEditComment(comment)}
              type="button"
            >
              <Pencil className="size-3" />
              Edit
            </button>
          ) : null}
          <button
            className="inline-flex items-center gap-1 text-xs font-semibold text-[#5f7894] transition hover:text-[#244c78]"
            onClick={() => onReplyComment(comment)}
            type="button"
          >
            <Reply className="size-3" />
            Reply
          </button>
        </div>
      </div>
      {comment.reply_to_comment_id ? (
        <div className="mt-3 rounded-xl border-l-2 border-[#7aa7d9] bg-[#f2f7fd] px-3 py-2 text-xs text-[#4b6682]">
          <p className="font-semibold">
            {comment.reply_to_user_name ||
              `User #${comment.reply_to_user_id ?? ""}`}
          </p>
          <p className="mt-1 line-clamp-2">{comment.reply_to_content}</p>
        </div>
      ) : null}
      <p className="mt-2 leading-6">{comment.content}</p>
    </div>
  ));
}

export function PostDetailDialog({
  open,
  post,
  comments,
  isAuthenticated,
  isChatOpen,
  isLoadingComments,
  isLiked,
  isSaved,
  isCommentPending,
  commentValue,
  replyTarget,
  editingComment,
  currentUserId,
  onClose,
  onLike,
  onSave,
  onLoadComments,
  onCancelReply,
  onCancelEdit,
  onCommentChange,
  onEditComment,
  onRegisterCommentInput,
  onReplyComment,
  onSubmitComment,
}: Readonly<PostDetailDialogProps>) {
  const authorName = post?.user_name || (post ? `User #${post.user_id}` : "");
  const commentPlaceholder = !isChatOpen
    ? "Open chat"
    : !isAuthenticated
      ? "Login to comment"
      : editingComment
        ? "Edit comment"
      : replyTarget
        ? `Reply to ${replyTarget.user_name || `User #${replyTarget.user_id}`}`
        : "Write a comment";

  return (
    <Dialog onOpenChange={(nextOpen) => !nextOpen && onClose()} open={open}>
      <DialogContent className="max-h-[92vh] w-[min(96vw,76rem)] overflow-hidden border-white/16 bg-[#101010] p-0 text-white">
        <div className="grid max-h-[92vh] overflow-hidden lg:grid-cols-[minmax(0,1.25fr)_minmax(360px,0.72fr)]">
          <div className="min-h-[56vh] bg-black lg:min-h-[82vh]">
            {post ? (
              <img
                alt={post.caption || post.location_name || "Post detail"}
                className="h-full max-h-[82vh] w-full object-contain"
                src={post.image_url}
              />
            ) : (
              <div className="flex h-full min-h-[56vh] items-center justify-center text-sm font-semibold text-white/68">
                Loading post
              </div>
            )}
          </div>

          <aside className="flex max-h-[92vh] min-h-0 flex-col bg-[#f7f7f5] text-[#1f1f1f]">
            <div className="border-b border-black/6 p-5">
              <DialogHeader>
                <DialogTitle className="text-xl leading-7 text-[#111]">
                  {post?.caption || "Community post"}
                </DialogTitle>
                <DialogDescription className="text-sm text-[#777]">
                  {post?.location_name || "Uploaded"} - {authorName}
                </DialogDescription>
              </DialogHeader>

              <div className="mt-4 flex items-center gap-2">
                <Button
                  aria-label={isLiked ? "Liked post" : "Like post"}
                  className={cn(
                    "rounded-full",
                    selectedClass("heart", isLiked),
                  )}
                  disabled={!post}
                  onClick={onLike}
                  size="icon-sm"
                  type="button"
                  variant="outline"
                >
                  <Heart
                    className={cn("size-4", isLiked ? "fill-current" : "")}
                  />
                </Button>
                <Button
                  aria-label={isSaved ? "Saved post" : "Save post"}
                  className={cn("rounded-full", selectedClass("save", isSaved))}
                  disabled={!post}
                  onClick={onSave}
                  size="icon-sm"
                  type="button"
                  variant="outline"
                >
                  <Bookmark
                    className={cn("size-4", isSaved ? "fill-current" : "")}
                  />
                </Button>
                <Button
                  aria-label="Focus comment input"
                  className={cn(
                    "rounded-full",
                    isChatOpen ? selectedClass("save", true) : "",
                  )}
                  disabled={!post}
                  onClick={onLoadComments}
                  size="icon-sm"
                  type="button"
                  variant="outline"
                >
                  <MessageCircle className="size-4" />
                </Button>
              </div>
            </div>

            <div className="min-h-0 flex-1 overflow-y-auto p-5">
              <div className="space-y-3">
                <DetailComments
                  comments={comments}
                  isChatOpen={isChatOpen}
                  isLoading={isLoadingComments}
                  currentUserId={currentUserId}
                  editingCommentId={editingComment?.id ?? null}
                  onEditComment={onEditComment}
                  onReplyComment={onReplyComment}
                  selectedReplyTargetId={replyTarget?.id ?? null}
                />
              </div>
            </div>

            <div className="border-t border-black/6 bg-white p-4">
              {editingComment ? (
                <div className="mb-3 flex items-start gap-2 rounded-2xl border border-[#dec078] bg-[#fff8e6] px-3 py-2 text-xs text-[#73551b]">
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
                <div className="mb-3 flex items-start gap-2 rounded-2xl border border-[#c8ddf1] bg-[#f2f7fd] px-3 py-2 text-xs text-[#4b6682]">
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
                  className="h-10 min-w-0 flex-1 rounded-full border border-black/8 bg-white px-4 text-sm outline-none placeholder:text-[#999] focus:border-[#111]/20"
                  disabled={
                    !isAuthenticated || !isChatOpen || isCommentPending || !post
                  }
                  onChange={(event) => onCommentChange(event.target.value)}
                  placeholder={commentPlaceholder}
                  ref={onRegisterCommentInput}
                  value={commentValue}
                />
                <Button
                  aria-label={
                    isAuthenticated ? "Submit comment" : "Login to comment"
                  }
                  className="rounded-full"
                  disabled={!isChatOpen || isCommentPending || !post}
                  onClick={onSubmitComment}
                  size="icon-sm"
                  type="button"
                  variant="outline"
                >
                  <Send className="size-4" />
                </Button>
              </div>
            </div>
          </aside>
        </div>
      </DialogContent>
    </Dialog>
  );
}
