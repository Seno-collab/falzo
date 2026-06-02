"use client";

import {
  Bookmark,
  CalendarDays,
  Bot,
  ChevronDown,
  Flag,
  Heart,
  HelpCircle,
  MapPin,
  Maximize2,
  MessageCircle,
  Pencil,
  Reply,
  Send,
  ShieldCheck,
  TriangleAlert,
  Trash2,
  UsersRound,
  X,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import Link from "next/link";
import type { KeyboardEvent } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { RainbowAvatar } from "@/components/ui/rainbow-avatar";
import type {
  Post,
  PostComment,
  PostTrustSummary,
  PostTrustVoteType,
} from "@/features/posts/types";
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
  bestTimeLabel?: string;
  distanceLabel?: string;
  isSubmittingComment: boolean;
  isSubmittingTrustVote: boolean;
  onOpen: (postId: number) => void;
  onLike: (postId: number) => void;
  onSave: (postId: number) => void;
  onDelete: (postId: number) => void;
  onReport: (postId: number) => void;
  onTrustVote: (postId: number, type: PostTrustVoteType) => void;
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
  "group mb-4 break-inside-avoid overflow-hidden rounded-2xl border border-black/5 bg-white shadow-[0_18px_48px_-38px_rgb(0_0_0/0.6)] transition duration-300 [contain-intrinsic-size:1px_620px] [content-visibility:auto] hover:-translate-y-1 hover:shadow-[0_26px_70px_-42px_rgb(0_0_0/0.72)] sm:rounded-[28px]";

const overlayClass =
  "pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0.03)_0%,rgb(0_0_0/0.02)_42%,rgb(0_0_0/0.62)_100%)] opacity-80 transition group-hover:opacity-100";

const imageFrameClasses = [
  "relative h-[22rem] cursor-zoom-in overflow-hidden bg-[#e9eef3] sm:h-[31rem]",
  "relative h-[17.5rem] cursor-zoom-in overflow-hidden bg-[#e9eef3] sm:h-[24rem]",
  "relative h-[20rem] cursor-zoom-in overflow-hidden bg-[#e9eef3] sm:h-[28rem]",
  "relative h-[16.5rem] cursor-zoom-in overflow-hidden bg-[#e9eef3] sm:h-[22rem]",
  "relative h-[23rem] cursor-zoom-in overflow-hidden bg-[#e9eef3] sm:h-[34rem]",
  "relative h-[18.5rem] cursor-zoom-in overflow-hidden bg-[#e9eef3] sm:h-[25rem]",
];

function getImageFrameClass(index: number) {
  return imageFrameClasses[index % imageFrameClasses.length];
}

function getPostAvatarUrl(post: Post) {
  return post.user_avatar_url || post.avatar_url || "";
}

function getPostCategoryLabel(post: Post) {
  const names = (post.categories ?? [])
    .map((category) => category.name.trim())
    .filter(Boolean);

  if (names.length > 0) {
    return names.slice(0, 3).join(", ");
  }

  return post.category_name || "Travel";
}

function getAuthorInitial(authorName: string) {
  return authorName.trim().charAt(0).toUpperCase() || "U";
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

const trustVoteOptions: Array<{
  type: PostTrustVoteType;
  label: string;
  description: string;
  icon: LucideIcon;
}> = [
  {
    type: "credible",
    label: "Looks real",
    description: "Place and image feel authentic",
    icon: ShieldCheck,
  },
  {
    type: "suspicious",
    label: "Suspicious",
    description: "Something looks manipulated",
    icon: TriangleAlert,
  },
  {
    type: "ai_generated",
    label: "AI generated",
    description: "Looks synthetic or overprocessed",
    icon: Bot,
  },
  {
    type: "wrong_context",
    label: "Wrong context",
    description: "Place, time, or caption seems off",
    icon: Flag,
  },
  {
    type: "unsure",
    label: "Not sure",
    description: "Needs more community review",
    icon: HelpCircle,
  },
];

function getTrustSummary(post: Post): PostTrustSummary {
  return (
    post.trust_summary ?? {
      status: "unreviewed",
      total_count: 0,
      credible_count: 0,
      suspicious_count: 0,
      ai_generated_count: 0,
      wrong_context_count: 0,
      unsure_count: 0,
    }
  );
}

function getTrustBadge(summary: PostTrustSummary) {
  switch (summary.status) {
    case "community_trusted":
      return {
        label: "Community trusted",
        className: "border-[#b7dfc1] bg-[#eef9f0] text-[#236238]",
      };
    case "community_suspicious":
      return {
        label: "Community doubts",
        className: "border-[#f0c2c2] bg-[#fff1f1] text-[#9f2f2f]",
      };
    case "disputed":
      return {
        label: "Disputed",
        className: "border-[#ebd59c] bg-[#fff8df] text-[#7b5a11]",
      };
    case "needs_more_context":
      return {
        label: "Needs context",
        className: "border-[#c8ddf1] bg-[#f2f7fd] text-[#2f6fb8]",
      };
    default:
      return {
        label: "Not reviewed",
        className: "border-black/10 bg-white text-[#555]",
      };
  }
}

function TrustSignals({
  disabled,
  isSubmitting,
  onVote,
  post,
}: Readonly<{
  disabled: boolean;
  isSubmitting: boolean;
  onVote: (type: PostTrustVoteType) => void;
  post: Post;
}>) {
  const summary = getTrustSummary(post);
  const badge = getTrustBadge(summary);
  const concernCount =
    summary.suspicious_count +
    summary.ai_generated_count +
    summary.wrong_context_count;
  const getVoteCount = (type: PostTrustVoteType) => {
    switch (type) {
      case "credible":
        return summary.credible_count;
      case "suspicious":
        return summary.suspicious_count;
      case "ai_generated":
        return summary.ai_generated_count;
      case "wrong_context":
        return summary.wrong_context_count;
      case "unsure":
        return summary.unsure_count;
    }
  };
  const renderVoteButton = (option: (typeof trustVoteOptions)[number]) => {
    const Icon = option.icon;
    const active = summary.viewer_vote === option.type;
    const voteCount = getVoteCount(option.type);

    return (
      <button
        aria-label={`Mark image as ${option.label}: ${option.description}`}
        className={cn(
          "inline-flex min-h-9 shrink-0 items-center gap-1.5 rounded-full border bg-white px-2.5 py-1.5 text-xs font-bold text-[#444] transition hover:border-[#9abfe5] hover:bg-[#f8fbff] hover:text-[#245f9a] disabled:cursor-not-allowed disabled:opacity-55",
          active &&
            "border-[#2f6fb8] bg-[#eef6ff] text-[#245f9a] shadow-[0_12px_28px_-22px_rgb(47_111_184/0.75)]",
        )}
        disabled={disabled || isSubmitting}
        key={option.type}
        onClick={() => onVote(option.type)}
        type="button"
      >
        <span
          className={cn(
            "flex size-6 shrink-0 items-center justify-center rounded-full bg-[#f1f1ef] text-[#666]",
            active && "bg-white text-[#245f9a]",
          )}
        >
          <Icon className="size-3.5" />
        </span>
        <span className="whitespace-nowrap">{option.label}</span>
        <span
          className={cn(
            "shrink-0 rounded-full bg-[#f4f4f2] px-1.5 py-0.5 text-[11px] font-bold leading-none text-[#777]",
            active && "bg-white text-[#245f9a]",
          )}
        >
          {voteCount}
        </span>
      </button>
    );
  };
  const selectedOption = trustVoteOptions.find(
    (option) => option.type === summary.viewer_vote,
  );
  const SelectedIcon = selectedOption?.icon ?? ShieldCheck;

  return (
    <div className="group/trust rounded-2xl border border-black/6 bg-[#f8f8f7] p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <Badge className={badge.className} variant="outline">
          <ShieldCheck className="size-3" />
          {badge.label}
        </Badge>
        <p className="text-xs font-semibold text-[#666]">
          {summary.total_count > 0
            ? `${summary.total_count} ratings / ${summary.credible_count} real / ${concernCount} flags`
            : "No community ratings"}
        </p>
      </div>
      <button
        aria-haspopup="listbox"
        className="mt-3 flex w-full items-center gap-3 rounded-xl border border-black/6 bg-white px-3.5 py-3 text-left transition hover:border-[#9abfe5] hover:bg-[#f8fbff] focus-visible:border-[#2f6fb8] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2f6fb8]/25 disabled:cursor-not-allowed disabled:opacity-55"
        disabled={disabled || isSubmitting}
        type="button"
      >
        <span
          className={cn(
            "flex size-9 shrink-0 items-center justify-center rounded-full bg-[#f1f1ef] text-[#666]",
            selectedOption && "bg-[#eef6ff] text-[#245f9a]",
          )}
        >
          <SelectedIcon className="size-4" />
        </span>
        <span className="min-w-0 flex-1">
          <span className="block text-sm font-bold leading-5 text-[#333]">
            {selectedOption?.label ?? "Choose trust rating"}
          </span>
          <span className="mt-0.5 block text-xs font-semibold leading-4 text-[#777]">
            {selectedOption?.description ?? "Hover to select how this image feels"}
          </span>
        </span>
        <ChevronDown className="size-4 shrink-0 text-[#777] transition group-hover/trust:rotate-180 group-focus-within/trust:rotate-180" />
      </button>
      <div className="flex max-h-0 flex-wrap gap-2 overflow-hidden opacity-0 transition-all duration-200 group-hover/trust:mt-2 group-hover/trust:max-h-32 group-hover/trust:opacity-100 group-focus-within/trust:mt-2 group-focus-within/trust:max-h-32 group-focus-within/trust:opacity-100">
        {trustVoteOptions.map((option) => renderVoteButton(option))}
      </div>
    </div>
  );
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
  bestTimeLabel,
  distanceLabel,
  isSubmittingComment,
  isSubmittingTrustVote,
  onOpen,
  onLike,
  onSave,
  onDelete,
  onReport,
  onTrustVote,
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
  const categoryLabel = getPostCategoryLabel(post);
  const authorAvatarUrl = getPostAvatarUrl(post);
  const isInitialImage = index < 2;
  const frameIndex = index % imageFrameClasses.length;
  const isTallFrame = frameIndex === 0 || frameIndex === 4;
  const savedCount = post.saves_count ?? 0;

  return (
    <article className={cardClass}>
      <div
        className={getImageFrameClass(frameIndex)}
        onClick={() => onOpen(post.id)}
        onKeyDown={(event) => onKeyboardOpen(event, () => onOpen(post.id))}
        role="button"
        tabIndex={0}
      >
        <img
          alt={post.caption || post.location_name || "Destination photo"}
          className="h-full w-full object-cover transition duration-700 ease-out group-hover:scale-[1.045]"
          decoding="async"
          fetchPriority={isInitialImage ? "high" : "low"}
          loading={isInitialImage ? "eager" : "lazy"}
          sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
          src={post.image_url}
        />
        <div className={overlayClass} />
        <div className="absolute left-3 right-3 top-3 flex items-start justify-between gap-2">
          <span className="rounded-full bg-white/88 px-3 py-1 text-xs font-semibold text-[#222] shadow-sm backdrop-blur-xl">
            {categoryLabel}
          </span>
          <div className="flex items-center gap-1 opacity-0 transition duration-200 group-hover:opacity-100 group-focus-within:opacity-100">
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
          <p className="inline-flex max-w-full items-center gap-1.5 rounded-full bg-white/16 px-2.5 py-1 text-xs font-semibold uppercase tracking-[0.14em] text-white/86 backdrop-blur-xl">
            <MapPin className="size-3.5 shrink-0" />
            <span className="truncate">{location}</span>
          </p>
          <h2
            className={cn(
              "mt-1 font-semibold leading-tight tracking-normal",
              isTallFrame
                ? "line-clamp-3 text-2xl sm:text-3xl"
                : "line-clamp-2 text-xl sm:text-2xl",
            )}
          >
            {title}
          </h2>
        </div>
      </div>

      <div className="space-y-3 p-3.5 sm:p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex min-w-0 items-center gap-2.5">
            <Link
              aria-label={`Open ${authorName}'s profile`}
              className="shrink-0 rounded-full"
              href={ROUTES.userProfile(post.user_id)}
            >
              <RainbowAvatar
                alt={authorName}
                fallback={getAuthorInitial(authorName)}
                size="md"
                src={authorAvatarUrl}
              />
            </Link>
            <div className="min-w-0">
              <Link
                className="block truncate text-sm font-semibold text-[#202020] hover:text-[#ff385c]"
                href={ROUTES.userProfile(post.user_id)}
              >
                {authorName}
              </Link>
              <p className="mt-0.5 text-xs font-medium text-[#777]">
                {categoryLabel} -{" "}
                {new Date(post.created_at).toLocaleDateString()}
              </p>
            </div>
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

        <div className="grid grid-cols-3 gap-2 rounded-2xl border border-black/6 bg-[#f8f8f7] p-2">
          <div className="min-w-0 rounded-xl bg-white px-2.5 py-2">
            <p className="flex items-center gap-1 text-[11px] font-semibold uppercase tracking-[0.12em] text-[#777]">
              <MapPin className="size-3" />
              Place
            </p>
            <p className="mt-1 truncate text-xs font-semibold text-[#222]">
              {distanceLabel ?? location}
            </p>
          </div>
          <div className="min-w-0 rounded-xl bg-white px-2.5 py-2">
            <p className="flex items-center gap-1 text-[11px] font-semibold uppercase tracking-[0.12em] text-[#777]">
              <CalendarDays className="size-3" />
              Best
            </p>
            <p className="mt-1 truncate text-xs font-semibold text-[#222]">
              {bestTimeLabel ?? "Golden hour"}
            </p>
          </div>
          <div className="min-w-0 rounded-xl bg-white px-2.5 py-2">
            <p className="flex items-center gap-1 text-[11px] font-semibold uppercase tracking-[0.12em] text-[#777]">
              <UsersRound className="size-3" />
              Saved
            </p>
            <p className="mt-1 truncate text-xs font-semibold text-[#222]">
              {savedCount > 0 ? `${savedCount} saved` : "Start board"}
            </p>
          </div>
        </div>

        <TrustSignals
          disabled={!isAuthenticated}
          isSubmitting={isSubmittingTrustVote}
          onVote={(type) => onTrustVote(post.id, type)}
          post={post}
        />

        <div className="flex flex-col gap-2 sm:flex-row">
          <Button
            aria-label={isSaved ? "Saved to trip" : "Save destination to trip"}
            className={cn(
              "min-w-0 flex-1 rounded-full",
              isSaved
                ? "bg-[#111] text-white hover:bg-[#222]"
                : "bg-[#ff385c] text-white hover:bg-[#e93152]",
            )}
            onClick={() => onSave(post.id)}
            type="button"
          >
            <Bookmark className={cn("size-4", isSaved ? "fill-current" : "")} />
            {isSaved ? "Saved to trip" : "Save to trip"}
          </Button>
          <Button
            aria-label="Open local discussion"
            className={cn(
              "rounded-full",
              activeClass("comment", commentsOpen),
            )}
            onClick={() => onToggleComments(post.id)}
            type="button"
            variant="outline"
          >
            <MessageCircle className="size-4" />
            Local tips
          </Button>
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
    </article>
  );
}
