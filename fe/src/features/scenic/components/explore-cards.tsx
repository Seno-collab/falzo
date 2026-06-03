"use client";

import {
  Bookmark,
  CalendarDays,
  Bot,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
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
  X,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
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
import { useI18n } from "@/i18n/locale-provider";
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

function getPostImageUrls(post: Post) {
  const urls = (post.image_urls ?? []).map((url) => url.trim()).filter(Boolean);
  return urls.length > 0 ? urls : [post.image_url];
}

type ExploreCopy = ReturnType<typeof useI18n>["messages"]["explorePage"];

function formatCardTemplate(
  template: string,
  values: Record<string, string | number>,
) {
  return Object.entries(values).reduce(
    (text, [key, value]) => text.replaceAll(`{${key}}`, String(value)),
    template,
  );
}

function getPostCategoryLabel(post: Post, copy: ExploreCopy) {
  const names = (post.categories ?? [])
    .map((category) => category.name.trim())
    .filter(Boolean);

  if (names.length > 0) {
    return names.slice(0, 3).join(", ");
  }

  return post.category_name || copy.categoryTravel;
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

function getTrustVoteOptions(copy: ExploreCopy): Array<{
  type: PostTrustVoteType;
  label: string;
  description: string;
  icon: LucideIcon;
}> {
  return [
    {
      type: "credible",
      label: copy.trustLooksReal,
      description: copy.trustLooksRealDescription,
      icon: ShieldCheck,
    },
    {
      type: "suspicious",
      label: copy.trustSuspicious,
      description: copy.trustSuspiciousDescription,
      icon: TriangleAlert,
    },
    {
      type: "ai_generated",
      label: copy.trustAiGenerated,
      description: copy.trustAiGeneratedDescription,
      icon: Bot,
    },
    {
      type: "wrong_context",
      label: copy.trustWrongContext,
      description: copy.trustWrongContextDescription,
      icon: Flag,
    },
    {
      type: "unsure",
      label: copy.trustUnsure,
      description: copy.trustUnsureDescription,
      icon: HelpCircle,
    },
  ];
}

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

function getTrustBadge(summary: PostTrustSummary, copy: ExploreCopy) {
  switch (summary.status) {
    case "community_trusted":
      return {
        label: copy.trustCommunityTrusted,
        className: "border-[#b7dfc1] bg-[#eef9f0] text-[#236238]",
      };
    case "community_suspicious":
      return {
        label: copy.trustCommunityDoubts,
        className: "border-[#f0c2c2] bg-[#fff1f1] text-[#9f2f2f]",
      };
    case "disputed":
      return {
        label: copy.trustDisputed,
        className: "border-[#ebd59c] bg-[#fff8df] text-[#7b5a11]",
      };
    case "needs_more_context":
      return {
        label: copy.trustNeedsContext,
        className: "border-[#c8ddf1] bg-[#f2f7fd] text-[#2f6fb8]",
      };
    default:
      return {
        label: copy.trustNotReviewed,
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
  const { messages } = useI18n();
  const copy = messages.explorePage;
  const trustVoteOptions = getTrustVoteOptions(copy);
  const [isPickerOpen, setIsPickerOpen] = useState(false);
  const summary = getTrustSummary(post);
  const badge = getTrustBadge(summary, copy);
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
        aria-label={formatCardTemplate(copy.trustRatingAria, {
          description: option.description,
          label: option.label,
        })}
        className={cn(
          "inline-flex min-h-9 shrink-0 items-center gap-1.5 rounded-full border bg-white px-2.5 py-1.5 text-xs font-bold text-[#444] transition hover:border-[#9abfe5] hover:bg-[#f8fbff] hover:text-[#245f9a] disabled:cursor-not-allowed disabled:opacity-55",
          active &&
            "border-[#2f6fb8] bg-[#eef6ff] text-[#245f9a] shadow-[0_12px_28px_-22px_rgb(47_111_184/0.75)]",
        )}
        disabled={disabled || isSubmitting}
        key={option.type}
        onClick={() => {
          onVote(option.type);
          setIsPickerOpen(false);
        }}
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
            ? formatCardTemplate(copy.trustSummary, {
                credible: summary.credible_count,
                flags: concernCount,
                total: summary.total_count,
              })
            : copy.noCommunityRatings}
        </p>
      </div>
      <button
        aria-expanded={isPickerOpen}
        aria-haspopup="listbox"
        className="mt-3 flex w-full items-center gap-3 rounded-xl border border-black/6 bg-white px-3.5 py-3 text-left transition hover:border-[#9abfe5] hover:bg-[#f8fbff] focus-visible:border-[#2f6fb8] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#2f6fb8]/25 disabled:cursor-not-allowed disabled:opacity-55"
        disabled={disabled || isSubmitting}
        onClick={() => setIsPickerOpen((current) => !current)}
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
            {selectedOption?.label ?? copy.chooseTrustRating}
          </span>
          <span className="mt-0.5 block text-xs font-semibold leading-4 text-[#777]">
            {selectedOption?.description ?? copy.trustRatingHint}
          </span>
        </span>
        <ChevronDown
          className={cn(
            "size-4 shrink-0 text-[#777] transition sm:group-hover/trust:rotate-180 sm:group-focus-within/trust:rotate-180",
            isPickerOpen && "rotate-180 sm:rotate-0",
          )}
        />
      </button>
      <div
        className={cn(
          "flex flex-wrap gap-2 overflow-hidden transition-all duration-200 sm:max-h-0 sm:opacity-0 sm:group-hover/trust:mt-2 sm:group-hover/trust:max-h-32 sm:group-hover/trust:opacity-100 sm:group-focus-within/trust:mt-2 sm:group-focus-within/trust:max-h-32 sm:group-focus-within/trust:opacity-100",
          isPickerOpen ? "mt-2 max-h-32 opacity-100" : "max-h-0 opacity-0",
        )}
      >
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
  const { messages } = useI18n();
  const copy = messages.explorePage;
  if (isLoading) {
    return (
      <p className="text-xs font-semibold text-[#555]">
        {copy.loadingComments}
      </p>
    );
  }

  if (comments.length === 0) {
    return (
      <p className="text-xs font-semibold text-[#555]">{copy.noCommentsYet}</p>
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
              {copy.edit}
            </button>
          ) : null}
          <button
            className="inline-flex items-center gap-1 text-xs font-semibold text-[#5f7894] transition hover:text-[#244c78]"
            onClick={() => onReply(comment)}
            type="button"
          >
            <Reply className="size-3" />
            {copy.reply}
          </button>
        </div>
      </div>
      <div className="mt-1 flex items-center gap-2 text-[11px] font-medium text-[#999]">
        <span>{new Date(comment.created_at).toLocaleDateString()}</span>
        {isCommentEdited(comment) ? (
          <span className="rounded-full bg-[#f2f7fd] px-2 py-0.5 text-[#5f7894]">
            {copy.edited}
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
  const { messages } = useI18n();
  const copy = messages.explorePage;
  const placeholder = !isAuthenticated
    ? copy.loginToComment
    : editingComment
      ? copy.editComment
      : replyTarget
        ? `${copy.replyingTo} ${
            replyTarget.user_name || `User #${replyTarget.user_id}`
          }`
        : copy.writeComment;

  return (
    <div className="mt-3 space-y-2">
      {editingComment ? (
        <div className="flex items-start gap-2 rounded-2xl border border-[#dec078] bg-[#fff8e6] px-3 py-2 text-xs text-[#73551b]">
          <Pencil className="mt-0.5 size-3 shrink-0" />
          <div className="min-w-0 flex-1">
            <p className="font-semibold">{copy.editingComment}</p>
            <p className="mt-0.5 truncate">{editingComment.content}</p>
          </div>
          <button
            aria-label={copy.cancelEdit}
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
              {copy.replyingTo}{" "}
              {replyTarget.user_name || `User #${replyTarget.user_id}`}
            </p>
            <p className="mt-0.5 truncate">{replyTarget.content}</p>
          </div>
          <button
            aria-label={copy.cancelReply}
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
          aria-label={isAuthenticated ? copy.submitComment : copy.loginToComment}
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
  const { messages } = useI18n();
  const copy = messages.explorePage;
  const title = post.caption || copy.travelStoryFallback;
  const location = post.location_name || copy.destinationFallback;
  const authorName = post.user_name || `User #${post.user_id}`;
  const categoryLabel = getPostCategoryLabel(post, copy);
  const authorAvatarUrl = getPostAvatarUrl(post);
  const imageUrls = getPostImageUrls(post);
  const [activeImageIndex, setActiveImageIndex] = useState(0);
  const activeImageUrl = imageUrls[activeImageIndex] ?? post.image_url;
  const isInitialImage = index < 2;
  const frameIndex = index % imageFrameClasses.length;
  const isTallFrame = frameIndex === 0 || frameIndex === 4;
  const savedCount = post.saves_count ?? 0;
  const [isCardActive, setIsCardActive] = useState(false);
  const revealCardActions = () => setIsCardActive(true);
  const hasMultipleImages = imageUrls.length > 1;
  const showPreviousImage = () =>
    setActiveImageIndex((current) =>
      current <= 0 ? imageUrls.length - 1 : current - 1,
    );
  const showNextImage = () =>
    setActiveImageIndex((current) => (current + 1) % imageUrls.length);

  return (
    <article
      className={cardClass}
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) {
          setIsCardActive(false);
        }
      }}
      onMouseEnter={revealCardActions}
      onMouseLeave={() => setIsCardActive(false)}
    >
      <div
        className={getImageFrameClass(frameIndex)}
        onClick={() => {
          if (!isCardActive) {
            revealCardActions();
            return;
          }

          onOpen(post.id);
        }}
        onKeyDown={(event) => onKeyboardOpen(event, () => onOpen(post.id))}
        role="button"
        tabIndex={0}
      >
        <img
          alt={post.caption || post.location_name || copy.destinationPhoto}
          className="h-full w-full object-cover transition duration-700 ease-out group-hover:scale-[1.045]"
          decoding="async"
          fetchPriority={isInitialImage ? "high" : "low"}
          loading={isInitialImage ? "eager" : "lazy"}
          sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
          src={activeImageUrl}
        />
        <div className={overlayClass} />
        <div className="absolute left-3 right-3 top-3 flex items-start justify-between gap-2">
          <span
            className={cn(
              "rounded-full bg-white/88 px-3 py-1 text-xs font-semibold text-[#222] opacity-0 shadow-sm backdrop-blur-xl transition sm:group-hover:opacity-100 sm:group-focus-within:opacity-100",
              isCardActive && "opacity-100",
            )}
          >
            {hasMultipleImages
              ? `${categoryLabel} / ${activeImageIndex + 1}/${imageUrls.length}`
              : categoryLabel}
          </span>
          <div
            className={cn(
              "flex items-center gap-1 opacity-0 transition duration-200 group-hover:opacity-100 group-focus-within:opacity-100",
              isCardActive && "opacity-100",
            )}
          >
            <Button
              aria-label={copy.openImage}
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
                aria-label={isLiked ? copy.liked : copy.like}
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
        {hasMultipleImages ? (
          <div
            className={cn(
              "-translate-y-1/2 absolute inset-x-3 top-1/2 flex items-center justify-between opacity-0 transition duration-200 group-hover:opacity-100 group-focus-within:opacity-100",
              isCardActive && "opacity-100",
            )}
          >
            <button
              aria-label={copy.previousImageInPost}
              className="flex size-9 items-center justify-center rounded-full bg-black/34 text-white shadow-lg backdrop-blur-xl transition hover:bg-black/48"
              onClick={(event) => {
                event.stopPropagation();
                revealCardActions();
                showPreviousImage();
              }}
              type="button"
            >
              <ChevronLeft className="size-4" />
            </button>
            <button
              aria-label={copy.nextImageInPost}
              className="flex size-9 items-center justify-center rounded-full bg-black/34 text-white shadow-lg backdrop-blur-xl transition hover:bg-black/48"
              onClick={(event) => {
                event.stopPropagation();
                revealCardActions();
                showNextImage();
              }}
              type="button"
            >
              <ChevronRight className="size-4" />
            </button>
          </div>
        ) : null}
        <div className="absolute inset-x-4 bottom-4 text-white">
          <div className="mb-2 flex min-w-0 items-center gap-2">
            <RainbowAvatar
              alt={authorName}
              className="shadow-[0_12px_26px_-18px_rgb(255_255_255/0.45),0_0_0_1px_rgb(255_255_255/0.24)]"
              fallback={getAuthorInitial(authorName)}
              size="sm"
              src={authorAvatarUrl}
            />
            <div className="min-w-0">
              <p className="truncate text-sm font-bold leading-5">
                {authorName}
              </p>
              <p className="hidden truncate text-xs font-semibold text-white/68 sm:block">
                {location}
              </p>
            </div>
          </div>
          <h2
            className={cn(
              "mt-1 font-semibold leading-tight tracking-normal",
              isTallFrame
                ? "line-clamp-3 text-2xl sm:text-3xl"
                : "line-clamp-2 text-xl sm:text-2xl",
              isCardActive ? "block" : "hidden sm:block",
            )}
          >
            {title}
          </h2>
        </div>
      </div>

      <div
        className={cn(
          "max-h-0 space-y-3 overflow-hidden p-0 opacity-0 transition-all duration-200 group-hover:max-h-[30rem] group-hover:p-3.5 group-hover:opacity-100 group-focus-within:max-h-[30rem] group-focus-within:p-3.5 group-focus-within:opacity-100 sm:group-hover:p-4 sm:group-focus-within:p-4",
          isCardActive && "max-h-[30rem] p-3.5 opacity-100 sm:p-4",
        )}
      >
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex min-w-0 items-center gap-2.5">
            <Link
              aria-label={`${messages.common.profile}: ${authorName}`}
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
        </div>

        <div className="flex flex-wrap gap-2 rounded-2xl border border-black/6 bg-[#f8f8f7] p-2">
          <Button
            aria-label={copy.openImage}
            className="h-8 rounded-full px-3 text-xs"
            onClick={() => onOpen(post.id)}
            size="sm"
            type="button"
            variant="outline"
          >
            <Maximize2 className="size-3.5" />
            {copy.open}
          </Button>
          {isAuthenticated ? (
            <Button
              aria-label={isLiked ? copy.liked : copy.likeTravelPost}
              className={cn(
                "h-8 rounded-full px-3 text-xs",
                activeClass("heart", isLiked),
              )}
              onClick={() => onLike(post.id)}
              size="sm"
              type="button"
              variant="outline"
            >
              <Heart className={cn("size-3.5", isLiked ? "fill-current" : "")} />
              {isLiked ? copy.liked : copy.like}
            </Button>
          ) : null}
          <Button
            aria-label={copy.viewComments}
            className={cn(
              "h-8 rounded-full px-3 text-xs",
              activeClass("comment", commentsOpen),
            )}
            onClick={() => onToggleComments(post.id)}
            size="sm"
            type="button"
            variant="outline"
          >
            <MessageCircle className="size-3.5" />
            {copy.tips}
            {comments.length > 0 ? (
              <span className="rounded-full bg-black/6 px-1.5 text-[11px]">
                {comments.length}
              </span>
            ) : null}
          </Button>
          {isAuthenticated ? (
            <Button
              aria-label={isSaved ? copy.navSaved : copy.saveDestination}
              className={cn(
                "h-8 rounded-full px-3 text-xs",
                isSaved
                  ? "border-[#111] bg-[#111] text-white hover:bg-[#222]"
                  : activeClass("save", isSaved),
              )}
              onClick={() => onSave(post.id)}
              size="sm"
              type="button"
              variant="outline"
            >
              <Bookmark
                className={cn("size-3.5", isSaved ? "fill-current" : "")}
              />
              {isSaved ? copy.navSaved : copy.save}
            </Button>
          ) : null}
          {isAuthenticated ? (
            currentUserId === post.user_id ? (
              <Button
                aria-label={copy.deleteTravelPost}
                className="h-8 rounded-full px-3 text-xs text-[#b4233f]"
                onClick={() => onDelete(post.id)}
                size="sm"
                type="button"
                variant="outline"
              >
                <Trash2 className="size-3.5" />
                {copy.delete}
              </Button>
            ) : (
              <Button
                aria-label={copy.reportTravelPost}
                className="h-8 rounded-full px-3 text-xs"
                onClick={() => onReport(post.id)}
                size="sm"
                type="button"
                variant="outline"
              >
                <Flag className="size-3.5" />
                {copy.report}
              </Button>
            )
          ) : null}
        </div>

        <div className="flex flex-wrap items-center gap-1.5 text-xs font-semibold text-[#555]">
          <span className="inline-flex max-w-full items-center gap-1 rounded-full bg-[#f8f8f7] px-2.5 py-1">
            <MapPin className="size-3.5 shrink-0 text-[#777]" />
            <span className="truncate">{distanceLabel ?? location}</span>
          </span>
          <span className="inline-flex items-center gap-1 rounded-full bg-[#f8f8f7] px-2.5 py-1">
            <CalendarDays className="size-3.5 text-[#777]" />
            {bestTimeLabel ?? copy.goldenHour}
          </span>
          <span className="inline-flex items-center gap-1 rounded-full bg-[#f8f8f7] px-2.5 py-1">
            <Bookmark className="size-3.5 text-[#777]" />
            {savedCount > 0
              ? formatCardTemplate(copy.savedCount, { count: savedCount })
              : copy.noSaves}
          </span>
        </div>

        <TrustSignals
          disabled={!isAuthenticated}
          isSubmitting={isSubmittingTrustVote}
          onVote={(type) => onTrustVote(post.id, type)}
          post={post}
        />

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
