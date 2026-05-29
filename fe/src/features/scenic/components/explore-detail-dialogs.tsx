"use client";

import {
  Bookmark,
  ChevronLeft,
  ChevronRight,
  Heart,
  LogInIcon,
  MessageCircle,
  Pencil,
  Reply,
  RotateCcw,
  Send,
  X,
  ZoomIn,
  ZoomOut,
} from "lucide-react";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import type { MouseEvent, PointerEvent, ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { RainbowAvatar } from "@/components/ui/rainbow-avatar";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { Post, PostComment } from "@/features/posts/types";
import { ROUTES } from "@/lib/routes";
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
  onNextPost?: () => void;
  onPreviousPost?: () => void;
  onRegisterCommentInput: (node: HTMLInputElement | null) => void;
  onReplyComment: (comment: PostComment) => void;
  onSubmitComment: () => void;
  carouselLabel?: string;
};

function isCommentEdited(comment: PostComment) {
  return comment.updated_at !== comment.created_at;
}

function getPostAvatarUrl(post: Post | null) {
  return post?.user_avatar_url || post?.avatar_url || "";
}

function getPostCategoryLabel(post: Post | null) {
  const names = (post?.categories ?? [])
    .map((category) => category.name.trim())
    .filter(Boolean);

  if (names.length > 0) {
    return names.slice(0, 3).join(", ");
  }

  return post?.category_name || "Travel";
}

function getAuthorInitial(authorName: string) {
  return authorName.trim().charAt(0).toUpperCase() || "U";
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

function CommentAvatar({
  name,
}: Readonly<{ name: string | null | undefined }>) {
  const initial = name?.trim().charAt(0) || "U";

  return (
    <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-[#111] text-xs font-bold uppercase text-white">
      {initial}
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
          <Link
            className="min-w-0 truncate text-sm font-bold text-[#171717] hover:text-[#ff385c]"
            href={ROUTES.userProfile(comment.user_id)}
          >
            {author}
          </Link>
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

        <p className="mt-1 whitespace-pre-wrap wrap-break-word text-sm leading-6 text-[#202020]">
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
  onNextPost,
  onPreviousPost,
  onRegisterCommentInput,
  onReplyComment,
  onSubmitComment,
  carouselLabel,
}: Readonly<PostDetailDialogProps>) {
  const authorName = post?.user_name || (post ? `User #${post.user_id}` : "");
  const authorAvatarUrl = getPostAvatarUrl(post);
  const categoryLabel = getPostCategoryLabel(post);
  const dragStartXRef = useRef<number | null>(null);
  const didNavigateDragRef = useRef(false);
  const [isImageZoomed, setIsImageZoomed] = useState(false);
  const [zoomOrigin, setZoomOrigin] = useState({ x: 50, y: 50 });
  const canNavigatePosts = Boolean(onPreviousPost || onNextPost);
  const getCommentPlaceholder = () => {
    if (!isChatOpen) return "Open comments";
    if (!isAuthenticated) return "Login to comment";
    if (editingComment) return "Edit comment";
    if (replyTarget)
      return `Reply to ${replyTarget.user_name || `User #${replyTarget.user_id}`}`;
    return "Write a comment";
  };
  const commentPlaceholder = getCommentPlaceholder();
  const commentCount = comments.length;
  let commentComposerNotice: ReactNode = null;

  if (editingComment) {
    commentComposerNotice = (
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
    );
  } else if (replyTarget) {
    commentComposerNotice = (
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
    );
  }

  useEffect(() => {
    setIsImageZoomed(false);
    setZoomOrigin({ x: 50, y: 50 });
  }, [open, post?.id]);

  function updateZoomOrigin(event: PointerEvent<HTMLElement> | MouseEvent<HTMLElement>) {
    const rect = event.currentTarget.getBoundingClientRect();
    const x = ((event.clientX - rect.left) / rect.width) * 100;
    const y = ((event.clientY - rect.top) / rect.height) * 100;
    setZoomOrigin({
      x: Math.max(8, Math.min(92, x)),
      y: Math.max(8, Math.min(92, y)),
    });
  }

  function toggleImageZoom(event: MouseEvent<HTMLButtonElement>) {
    if (!post || didNavigateDragRef.current) {
      didNavigateDragRef.current = false;
      return;
    }

    updateZoomOrigin(event);
    setIsImageZoomed((current) => !current);
  }

  function resetImageZoom() {
    setIsImageZoomed(false);
    setZoomOrigin({ x: 50, y: 50 });
  }

  function handleImagePointerDown(event: PointerEvent<HTMLButtonElement>) {
    if (!canNavigatePosts) {
      return;
    }

    dragStartXRef.current = event.clientX;
    didNavigateDragRef.current = false;
    if (event.currentTarget.hasPointerCapture?.(event.pointerId) === false) {
      event.currentTarget.setPointerCapture(event.pointerId);
    }
  }

  function handleImagePointerUp(event: PointerEvent<HTMLButtonElement>) {
    const startX = dragStartXRef.current;
    dragStartXRef.current = null;
    if (startX === null || !canNavigatePosts || didNavigateDragRef.current) {
      return;
    }

    const deltaX = event.clientX - startX;
    if (Math.abs(deltaX) < 48) {
      return;
    }

    if (deltaX < 0) {
      didNavigateDragRef.current = true;
      onNextPost?.();
      return;
    }

    didNavigateDragRef.current = true;
    onPreviousPost?.();
  }

  return (
    <Dialog onOpenChange={(nextOpen) => !nextOpen && onClose()} open={open}>
      <DialogContent
        className="h-[min(94svh,860px)] w-[calc(100vw-1rem)] overflow-hidden border-white/16 bg-[#050505]/94 p-0 text-white shadow-[0_36px_110px_-44px_rgb(0_0_0/0.92)] backdrop-blur-xl sm:w-[min(98vw,86rem)]"
        overlayClassName="overflow-hidden bg-black/72 backdrop-blur-0"
        overlayChildren={
          post ? (
            <div className="pointer-events-none absolute inset-0">
              <img
                alt=""
                aria-hidden="true"
                className="h-full w-full scale-110 object-cover opacity-55 blur-2xl saturate-125"
                decoding="async"
                src={post.image_url}
              />
              <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_48%,rgb(0_0_0/0.18)_0%,rgb(0_0_0/0.42)_46%,rgb(0_0_0/0.82)_100%)]" />
              <div className="absolute inset-0 bg-[linear-gradient(90deg,rgb(0_0_0/0.62)_0%,rgb(0_0_0/0.2)_48%,rgb(0_0_0/0.66)_100%)]" />
            </div>
          ) : null
        }
      >
        <div className="flex h-full min-h-0 flex-col overflow-hidden lg:grid lg:grid-cols-[minmax(0,1fr)_minmax(360px,0.45fr)]">
          <div className="relative h-[42svh] min-h-0 shrink-0 overflow-hidden bg-black lg:h-auto lg:min-h-0 lg:shrink">
            {post ? (
              <button
                aria-label={
                  isImageZoomed
                    ? "Zoom out destination photo"
                    : "Zoom in destination photo"
                }
                className={cn(
                  "h-full w-full touch-pan-y overflow-hidden bg-black",
                  isImageZoomed ? "cursor-zoom-out" : "cursor-zoom-in",
                )}
                onClick={toggleImageZoom}
                onPointerDown={handleImagePointerDown}
                onPointerMove={(event) => {
                  if (isImageZoomed) {
                    updateZoomOrigin(event);
                  }
                }}
                onPointerUp={handleImagePointerUp}
                type="button"
              >
                <img
                  alt={
                    post.caption || post.location_name || "Destination detail"
                  }
                  className="pointer-events-none h-full w-full select-none object-contain transition duration-500 ease-out"
                  decoding="async"
                  draggable={false}
                  fetchPriority="high"
                  src={post.image_url}
                  style={{
                    transform: isImageZoomed ? "scale(1.8)" : "scale(1)",
                    transformOrigin: `${zoomOrigin.x}% ${zoomOrigin.y}%`,
                  }}
                />
              </button>
            ) : (
              <div className="flex h-full min-h-[48vh] items-center justify-center text-sm font-semibold text-white/68">
                Loading travel post
              </div>
            )}

            <div className="absolute left-3 top-3 max-w-[calc(100%-7rem)] rounded-2xl border border-white/14 bg-black/42 px-3 py-2 text-white shadow-lg backdrop-blur-xl sm:left-5 sm:top-5">
              <div className="flex items-center gap-2">
                <span className="flex size-7 items-center justify-center rounded-full bg-white text-[#111]">
                  <ZoomIn className="size-3.5" />
                </span>
                <div className="min-w-0">
                  <p className="text-xs font-bold uppercase tracking-[0.14em] text-white/68">
                    Travel lens
                  </p>
                  <p className="truncate text-sm font-semibold">
                    {isImageZoomed
                      ? "Move around to inspect the scene"
                      : "Click the photo to zoom into details"}
                  </p>
                </div>
              </div>
            </div>

            <div className="absolute right-3 top-3 flex items-center gap-1 rounded-full border border-white/14 bg-black/42 p-1 shadow-lg backdrop-blur-xl sm:right-5 sm:top-5">
              <button
                aria-label={isImageZoomed ? "Zoom out" : "Zoom in"}
                className="flex size-9 items-center justify-center rounded-full text-white transition hover:bg-white/16"
                disabled={!post}
                onClick={() => setIsImageZoomed((current) => !current)}
                type="button"
              >
                {isImageZoomed ? (
                  <ZoomOut className="size-4" />
                ) : (
                  <ZoomIn className="size-4" />
                )}
              </button>
              <button
                aria-label="Reset image view"
                className="flex size-9 items-center justify-center rounded-full text-white transition hover:bg-white/16 disabled:opacity-40"
                disabled={!post || !isImageZoomed}
                onClick={resetImageZoom}
                type="button"
              >
                <RotateCcw className="size-4" />
              </button>
            </div>

            {onPreviousPost ? (
              <button
                aria-label="Previous post"
                className="-translate-y-1/2 absolute left-3 top-1/2 flex size-11 items-center justify-center rounded-full bg-white/14 text-white shadow-lg backdrop-blur-xl transition hover:bg-white/24 sm:left-5 sm:size-12"
                onClick={onPreviousPost}
                type="button"
              >
                <ChevronLeft className="size-5" />
              </button>
            ) : null}

            {onNextPost ? (
              <button
                aria-label="Next post"
                className="-translate-y-1/2 absolute right-3 top-1/2 flex size-11 items-center justify-center rounded-full bg-white/14 text-white shadow-lg backdrop-blur-xl transition hover:bg-white/24 sm:right-5 sm:size-12"
                onClick={onNextPost}
                type="button"
              >
                <ChevronRight className="size-5" />
              </button>
            ) : null}

            {carouselLabel ? (
              <div className="-translate-x-1/2 absolute left-1/2 top-4 rounded-full bg-black/48 px-3 py-1 text-xs font-bold text-white/86 backdrop-blur-xl">
                {carouselLabel}
              </div>
            ) : null}

            <div className="absolute inset-x-0 bottom-0 bg-linear-to-t from-black/86 via-black/30 to-transparent px-4 pb-4 pt-20 sm:px-5 sm:pb-5 sm:pt-24">
              <div className="max-w-2xl">
                <div className="flex min-w-0 items-center gap-2.5">
                  <RainbowAvatar
                    alt={authorName}
                    className="shadow-[0_12px_26px_-18px_rgb(255_255_255/0.45),0_0_0_1px_rgb(255_255_255/0.24)]"
                    fallback={getAuthorInitial(authorName)}
                    size="md"
                    src={authorAvatarUrl}
                  />
                  {post ? (
                    <Link
                      className="pointer-events-auto truncate text-sm font-bold text-white hover:text-[#ff8aa0]"
                      href={ROUTES.userProfile(post.user_id)}
                    >
                      {authorName}
                    </Link>
                  ) : (
                    <p className="truncate text-sm font-bold text-white">
                      {authorName}
                    </p>
                  )}
                </div>
                <p className="mt-2 line-clamp-2 text-sm leading-5 text-white/90 sm:line-clamp-3 sm:leading-6">
                  {post?.caption || "Travel story"}
                </p>
                <p className="mt-2 text-xs font-semibold text-white/60">
                  {post?.location_name || "Destination"} / {categoryLabel}
                </p>
              </div>
            </div>

            <div className="absolute bottom-4 right-3 flex flex-col items-center gap-2 sm:bottom-6 sm:right-4 sm:gap-3">
              {isAuthenticated ? (
                <button
                  aria-label={isLiked ? "Liked" : "Like travel post"}
                  className={cn(
                    "flex size-10 items-center justify-center rounded-full bg-white/12 text-white shadow-lg backdrop-blur-xl transition hover:bg-white/20 sm:size-12",
                    isLiked ? "text-[#ff4d6d]" : "",
                  )}
                  disabled={!post}
                  onClick={onLike}
                  type="button"
                >
                  <Heart
                    className={cn(
                      "size-4 sm:size-5",
                      isLiked ? "fill-current" : "",
                    )}
                  />
                </button>
              ) : null}
              <button
                aria-label="Open comments"
                className={cn(
                  "flex size-10 items-center justify-center rounded-full bg-white/12 text-white shadow-lg backdrop-blur-xl transition hover:bg-white/20 sm:size-12",
                  isChatOpen ? "bg-white text-[#111] hover:bg-white" : "",
                )}
                disabled={!post}
                onClick={onLoadComments}
                type="button"
              >
                <MessageCircle className="size-4 sm:size-5" />
              </button>
              {isAuthenticated ? (
                <button
                  aria-label={isSaved ? "Saved" : "Save destination"}
                  className={cn(
                    "flex size-10 items-center justify-center rounded-full bg-white/12 text-white shadow-lg backdrop-blur-xl transition hover:bg-white/20 sm:size-12",
                    isSaved ? "text-[#f8c14a]" : "",
                  )}
                  disabled={!post}
                  onClick={onSave}
                  type="button"
                >
                  <Bookmark
                    className={cn(
                      "size-4 sm:size-5",
                      isSaved ? "fill-current" : "",
                    )}
                  />
                </button>
              ) : null}
            </div>
          </div>

          <aside className="flex min-h-0 flex-1 flex-col bg-white text-[#1f1f1f] lg:flex-auto">
            <div className="border-b border-black/8 px-4 py-3 sm:px-5 sm:py-4">
              <DialogHeader>
                <DialogTitle className="text-lg leading-7 text-[#111]">
                  Comments
                </DialogTitle>
                <DialogDescription className="text-sm text-[#777]">
                  {commentCount} comment{commentCount === 1 ? "" : "s"} about
                  this destination.
                </DialogDescription>
              </DialogHeader>
            </div>

            <div className="min-h-0 flex-1 overflow-y-auto px-3 py-3 sm:py-4">
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

            <div className="border-t border-black/8 bg-white p-3 sm:p-4">
              {commentComposerNotice}
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
