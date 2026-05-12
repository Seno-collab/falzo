"use client";

import {
  Bookmark,
  Heart,
  LogInIcon,
  MessageCircle,
  Pencil,
  Reply,
  Send,
  X,
} from "lucide-react";
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

function isCommentEdited(comment: PostComment) {
  return comment.updated_at !== comment.created_at;
}

function getCommentAuthor(comment: PostComment) {
  return comment.user_name || `User #${comment.user_id}`;
}

function formatCommentDate(comment: PostComment) {
  return new Date(comment.created_at).toLocaleDateString();
}

function getCommentThreads(comments: PostComment[]) {
  const commentsById = new Map<number, PostComment>();
  const repliesByParentId = new Map<number, PostComment[]>();
  const threadRoots: PostComment[] = [];

  for (const comment of comments) {
    commentsById.set(comment.id, comment);
  }

  for (const comment of comments) {
    const parentId = comment.reply_to_comment_id;
    if (parentId && commentsById.has(parentId)) {
      repliesByParentId.set(parentId, [
        ...(repliesByParentId.get(parentId) ?? []),
        comment,
      ]);
      continue;
    }

    threadRoots.push(comment);
  }

  return threadRoots.map((comment) => ({
    comment,
    replies: repliesByParentId.get(comment.id) ?? [],
  }));
}

function CommentAvatar({ name }: Readonly<{ name: string }>) {
  return (
    <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-[#111] text-xs font-bold uppercase text-white">
      {name.trim().charAt(0) || "U"}
    </span>
  );
}

function CommentBubble({
  comment,
  compact = false,
  currentUserId,
  editingCommentId,
  onEditComment,
  onReplyComment,
  selectedReplyTargetId,
}: Readonly<{
  comment: PostComment;
  compact?: boolean;
  currentUserId: number | null;
  editingCommentId: number | null;
  onEditComment: (comment: PostComment) => void;
  onReplyComment: (comment: PostComment) => void;
  selectedReplyTargetId: number | null;
}>) {
  const author = getCommentAuthor(comment);
  const isEditing = editingCommentId === comment.id;
  const isReplyTarget = selectedReplyTargetId === comment.id;

  return (
    <div
      className={cn(
        "flex gap-3 rounded-2xl px-3 py-2.5 transition",
        isEditing ? "bg-[#fff6dc]" : "",
        isReplyTarget ? "bg-[#edf6ff]" : "",
      )}
    >
      <CommentAvatar name={author} />
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
          <p className="min-w-0 truncate text-sm font-bold text-[#171717]">
            {author}
          </p>
          <span className="text-xs font-medium text-[#8a8a8a]">
            {formatCommentDate(comment)}
          </span>
          {isCommentEdited(comment) ? (
            <span className="rounded-full bg-[#f1f1f1] px-2 py-0.5 text-[11px] font-semibold text-[#777]">
              edited
            </span>
          ) : null}
        </div>

        {comment.reply_to_comment_id && !compact ? (
          <div className="mt-2 rounded-xl border-l-2 border-[#d7d7d7] bg-[#f6f6f6] px-3 py-2 text-xs text-[#666]">
            <p className="font-semibold">
              {comment.reply_to_user_name ||
                `User #${comment.reply_to_user_id ?? ""}`}
            </p>
            <p className="mt-0.5 line-clamp-2">{comment.reply_to_content}</p>
          </div>
        ) : null}

        <p className="mt-1 whitespace-pre-wrap break-words text-sm leading-6 text-[#202020]">
          {comment.content}
        </p>

        <div className="mt-2 flex items-center gap-4">
          <button
            className="text-xs font-bold text-[#777] transition hover:text-[#111]"
            onClick={() => onReplyComment(comment)}
            type="button"
          >
            Reply
          </button>
          {currentUserId === comment.user_id ? (
            <button
              className="inline-flex items-center gap-1 text-xs font-bold text-[#8c6a1f] transition hover:text-[#5e430f]"
              onClick={() => onEditComment(comment)}
              type="button"
            >
              <Pencil className="size-3" />
              Edit
            </button>
          ) : null}
        </div>
      </div>
    </div>
  );
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
      <div className="rounded-2xl bg-[#f4f4f4] px-4 py-5 text-sm font-bold text-[#555]">
        Loading comments
      </div>
    );
  }

  if (!isChatOpen) {
    return (
      <div className="rounded-2xl bg-[#f4f4f4] px-4 py-5 text-sm font-bold text-[#555]">
        Open comments.
      </div>
    );
  }

  if (comments.length === 0) {
    return (
      <div className="rounded-2xl bg-[#f4f4f4] px-4 py-5 text-sm font-bold text-[#555]">
        No comments yet.
      </div>
    );
  }

  return getCommentThreads(comments).map(({ comment, replies }) => (
    <div className="space-y-1" key={comment.id}>
      <CommentBubble
        comment={comment}
        currentUserId={currentUserId}
        editingCommentId={editingCommentId}
        onEditComment={onEditComment}
        onReplyComment={onReplyComment}
        selectedReplyTargetId={selectedReplyTargetId}
      />
      {replies.length > 0 ? (
        <div className="ml-12 border-l border-[#e5e5e5] pl-3">
          <p className="px-3 py-1 text-xs font-bold text-[#8a8a8a]">
            {replies.length} repl{replies.length === 1 ? "y" : "ies"}
          </p>
          <div className="space-y-1">
            {replies.map((reply) => (
              <CommentBubble
                comment={reply}
                compact
                currentUserId={currentUserId}
                editingCommentId={editingCommentId}
                key={reply.id}
                onEditComment={onEditComment}
                onReplyComment={onReplyComment}
                selectedReplyTargetId={selectedReplyTargetId}
              />
            ))}
          </div>
        </div>
      ) : null}
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
  const getCommentPlaceholder = () => {
    if (!isChatOpen) return "Open chat";
    if (!isAuthenticated)
      return "Login to comment";
    if (editingComment) return "Edit comment";
    if (replyTarget)
      return `Reply to ${replyTarget.user_name || `User #${replyTarget.user_id}`}`;
    return "Write a comment";
  };
  const commentPlaceholder = getCommentPlaceholder();
  const commentCount = comments.length;

  return (
    <Dialog onOpenChange={(nextOpen) => !nextOpen && onClose()} open={open}>
      <DialogContent className="h-[min(94vh,860px)] w-[min(98vw,86rem)] overflow-hidden border-white/16 bg-[#050505] p-0 text-white">
        <div className="grid h-full min-h-0 overflow-hidden lg:grid-cols-[minmax(0,1fr)_minmax(360px,0.45fr)]">
          <div className="relative min-h-[48vh] bg-black lg:min-h-0">
            {post ? (
              <img
                alt={post.caption || post.location_name || "Post detail"}
                className="h-full w-full object-contain"
                src={post.image_url}
              />
            ) : (
              <div className="flex h-full min-h-[48vh] items-center justify-center text-sm font-semibold text-white/68">
                Loading post
              </div>
            )}

            <div className="pointer-events-none absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/86 via-black/30 to-transparent px-5 pb-5 pt-24">
              <div className="max-w-2xl">
                <p className="text-sm font-bold text-white">{authorName}</p>
                <p className="mt-2 line-clamp-3 text-sm leading-6 text-white/90">
                  {post?.caption || "Community post"}
                </p>
                <p className="mt-2 text-xs font-semibold text-white/60">
                  {post?.location_name || "Uploaded"}
                </p>
              </div>
            </div>

            <div className="absolute bottom-6 right-4 flex flex-col items-center gap-3">
              <button
                aria-label={isLiked ? "Liked post" : "Like post"}
                className={cn(
                  "flex size-12 items-center justify-center rounded-full bg-white/12 text-white shadow-lg backdrop-blur-xl transition hover:bg-white/20",
                  isLiked ? "text-[#ff4d6d]" : "",
                )}
                disabled={!post}
                onClick={onLike}
                type="button"
              >
                <Heart
                  className={cn("size-5", isLiked ? "fill-current" : "")}
                />
              </button>
              <button
                aria-label="Open comments"
                className={cn(
                  "flex size-12 items-center justify-center rounded-full bg-white/12 text-white shadow-lg backdrop-blur-xl transition hover:bg-white/20",
                  isChatOpen ? "bg-white text-[#111] hover:bg-white" : "",
                )}
                disabled={!post}
                onClick={onLoadComments}
                type="button"
              >
                <MessageCircle className="size-5" />
              </button>
              <button
                aria-label={isSaved ? "Saved post" : "Save post"}
                className={cn(
                  "flex size-12 items-center justify-center rounded-full bg-white/12 text-white shadow-lg backdrop-blur-xl transition hover:bg-white/20",
                  isSaved ? "text-[#f8c14a]" : "",
                )}
                disabled={!post}
                onClick={onSave}
                type="button"
              >
                <Bookmark
                  className={cn("size-5", isSaved ? "fill-current" : "")}
                />
              </button>
            </div>
          </div>

          <aside className="flex min-h-0 flex-col bg-white text-[#1f1f1f]">
            <div className="border-b border-black/8 px-5 py-4">
              <DialogHeader>
                <DialogTitle className="text-lg leading-7 text-[#111]">
                  Comments
                </DialogTitle>
                <DialogDescription className="text-sm text-[#777]">
                  {commentCount} comment{commentCount === 1 ? "" : "s"} grouped
                  by conversation.
                </DialogDescription>
              </DialogHeader>
            </div>

            <div className="min-h-0 flex-1 overflow-y-auto px-3 py-4">
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

            <div className="border-t border-black/8 bg-white p-4">
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
                  {isAuthenticated ? (
                    <Send className="size-4" />
                  ) : (
                    <LogInIcon className="size-4 text-muted-foreground" />
                  )}
                </Button>
              </div>
            </div>
          </aside>
        </div>
      </DialogContent>
    </Dialog>
  );
}
