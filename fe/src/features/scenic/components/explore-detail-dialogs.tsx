"use client";

import { Bookmark, Heart, MessageCircle, Send } from "lucide-react";
import { ScenicImage } from "@/components/scenic-image";
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

type PinDetailData = {
  id: string;
  imageId: string;
  title: string;
  author: string;
  city: string;
  gradient: string;
};

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
  onClose: () => void;
  onLike: () => void;
  onSave: () => void;
  onLoadComments: () => void;
  onCommentChange: (value: string) => void;
  onSubmitComment: () => void;
};

type PinDetailDialogProps = {
  open: boolean;
  pin: PinDetailData | null;
  isSaved: boolean;
  onClose: () => void;
  onToggleSaved: () => void;
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

function DetailComments({
  comments,
  isChatOpen,
  isLoading,
}: Readonly<{
  comments: PostComment[];
  isChatOpen: boolean;
  isLoading: boolean;
}>) {
  if (!isChatOpen) {
    return (
      <div className="rounded-2xl border border-black/6 bg-white px-4 py-5 text-sm font-semibold text-[#555]">
        Chat is closed.
      </div>
    );
  }

  if (isLoading) {
    return <p className="text-sm font-semibold text-[#555]">Loading comments</p>;
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
      className="rounded-2xl border border-black/5 bg-white px-4 py-3 text-sm text-[#333]"
      key={comment.id}
    >
      <div className="flex items-center justify-between gap-3">
        <p className="text-xs font-semibold text-[#777]">
          User #{comment.user_id}
        </p>
        <p className="text-xs font-medium text-[#999]">
          {new Date(comment.created_at).toLocaleDateString()}
        </p>
      </div>
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
  onClose,
  onLike,
  onSave,
  onLoadComments,
  onCommentChange,
  onSubmitComment,
}: Readonly<PostDetailDialogProps>) {
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
                  {post?.location_name || "Uploaded"} - User #
                  {post?.user_id ?? ""}
                </DialogDescription>
              </DialogHeader>

              <div className="mt-4 flex items-center gap-2">
                <Button
                  aria-label={isLiked ? "Liked post" : "Like post"}
                  className={cn("rounded-full", selectedClass("heart", isLiked))}
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
                />
              </div>
            </div>

            <div className="border-t border-black/6 bg-white p-4">
              <div className="flex items-center gap-2">
                <input
                  className="h-10 min-w-0 flex-1 rounded-full border border-black/8 bg-white px-4 text-sm outline-none placeholder:text-[#999] focus:border-[#111]/20"
                  disabled={
                    !isAuthenticated || !isChatOpen || isCommentPending || !post
                  }
                  onChange={(event) => onCommentChange(event.target.value)}
                  placeholder={
                    isChatOpen
                      ? isAuthenticated
                        ? "Write a comment"
                        : "Login to comment"
                      : "Open chat"
                  }
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

export function PinDetailDialog({
  open,
  pin,
  isSaved,
  onClose,
  onToggleSaved,
}: Readonly<PinDetailDialogProps>) {
  return (
    <Dialog onOpenChange={(nextOpen) => !nextOpen && onClose()} open={open}>
      <DialogContent className="max-h-[92vh] w-[min(96vw,72rem)] overflow-hidden border-white/16 bg-[#101010] p-0 text-white">
        <div className="grid max-h-[92vh] overflow-hidden lg:grid-cols-[minmax(0,1.2fr)_340px]">
          <div
            className={cn(
              "min-h-[56vh] bg-black lg:min-h-[82vh]",
              pin?.gradient,
            )}
          >
            {pin ? (
              <ScenicImage
                alt={pin.title}
                className="h-full max-h-[82vh] w-full object-contain"
                id={pin.imageId}
                sizes="96vw"
              />
            ) : null}
          </div>
          <aside className="flex flex-col bg-[#f7f7f5] p-5 text-[#1f1f1f]">
            <DialogHeader>
              <DialogTitle className="text-xl leading-7 text-[#111]">
                {pin?.title || "Image"}
              </DialogTitle>
              <DialogDescription className="text-sm text-[#777]">
                {pin?.city || "Explore"} - {pin?.author || ""}
              </DialogDescription>
            </DialogHeader>

            <div className="mt-4 flex items-center gap-2">
              <Button
                aria-label={isSaved ? "Saved image" : "Save image"}
                className={cn("rounded-full", selectedClass("heart", isSaved))}
                disabled={!pin}
                onClick={onToggleSaved}
                size="icon-sm"
                type="button"
                variant="outline"
              >
                <Bookmark
                  className={cn("size-4", isSaved ? "fill-current" : "")}
                />
              </Button>
              <Button
                aria-label="Save favorite"
                className={cn("rounded-full", selectedClass("heart", isSaved))}
                disabled={!pin}
                onClick={onToggleSaved}
                size="icon-sm"
                type="button"
                variant="outline"
              >
                <Heart
                  className={cn("size-4", isSaved ? "fill-current" : "")}
                />
              </Button>
            </div>

            <div className="mt-5 rounded-2xl border border-black/6 bg-white px-4 py-5 text-sm font-semibold text-[#555]">
              Comments are available on community posts from the backend feed.
            </div>
          </aside>
        </div>
      </DialogContent>
    </Dialog>
  );
}
