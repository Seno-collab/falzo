"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  CalendarDays,
  Compass,
  ImageIcon,
  Loader2,
  UserPlus,
  UserRound,
  UserX,
} from "lucide-react";
import Link from "next/link";
import { useEffect } from "react";
import { toast } from "sonner";
import { EmptyState } from "@/components/feedback/empty-state";
import { LoadingPanel } from "@/components/feedback/loading-panel";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { RainbowAvatar } from "@/components/ui/rainbow-avatar";
import {
  getApiErrorMessage,
  getMeApi,
  hasAuthSession,
} from "@/features/auth/api";
import type { AuthUser } from "@/features/auth/types";
import {
  followUserApi,
  getPublicProfileApi,
  unfollowUserApi,
} from "@/features/social/api";
import { ROUTES } from "@/lib/routes";

function readAuthUserId(user: AuthUser | null | undefined) {
  const rawId = user?.id ?? user?.user_id ?? user?.userId ?? user?.subject;
  const id = Number(rawId);
  return Number.isFinite(id) && id > 0 ? id : null;
}

function formatJoinedDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Member";
  }

  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    year: "numeric",
  }).format(date);
}

export function PublicProfileScreen({
  userId,
}: Readonly<{
  userId: number;
}>) {
  const queryClient = useQueryClient();
  const validUserId = Number.isFinite(userId) && userId > 0;

  useEffect(() => {
    document.title = "Public profile | Falzo";
  }, []);

  const profileQuery = useQuery({
    enabled: validUserId,
    queryKey: ["users", userId, "public-profile"],
    queryFn: ({ signal }) => getPublicProfileApi(userId, { signal }),
    retry: false,
    staleTime: 60_000,
  });

  const meQuery = useQuery({
    enabled: hasAuthSession(),
    queryKey: ["auth", "me", "public-profile"],
    queryFn: ({ signal }) => getMeApi<AuthUser>({ signal }),
    retry: false,
    staleTime: 60_000,
  });

  const followMutation = useMutation({
    mutationFn: async () => {
      const profile = profileQuery.data;
      if (!profile) {
        return;
      }

      if (profile.is_following) {
        await unfollowUserApi(profile.user_id);
      } else {
        await followUserApi(profile.user_id);
      }
    },
    onError: (error) => toast.error(getApiErrorMessage(error)),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["users", userId, "public-profile"],
      });
    },
  });

  const currentUserId = readAuthUserId(meQuery.data);
  const profile = profileQuery.data;
  const profileAvatarUrl = profile?.avatar_url || profile?.avatarUrl || null;
  const canFollow =
    Boolean(profile) && hasAuthSession() && currentUserId !== profile?.user_id;

  return (
    <PageShell
      contentClassName="space-y-5 pb-12"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "explore",
              icon: <Compass className="size-4" />,
              label: "Explore",
              to: ROUTES.explore,
              variant: "outline",
            },
            {
              id: "back",
              icon: <ArrowLeft className="size-4" />,
              label: "Back",
              to: ROUTES.explore,
              variant: "outline",
            },
          ]}
          brand="Public profile"
          brandIcon={<UserRound className="size-3.5" />}
          mobileMenuTitle="Profile"
          subtitle="View posts and follow creators."
        />
      }
    >
      {!validUserId ? (
        <EmptyState
          description="The profile link does not contain a valid user id."
          title="Invalid profile"
        />
      ) : profileQuery.isLoading ? (
        <LoadingPanel
          description="Loading creator profile and public posts."
          title="Loading profile"
        />
      ) : profileQuery.isError || !profile ? (
        <EmptyState
          description={getApiErrorMessage(profileQuery.error)}
          title="Profile not found"
        />
      ) : (
        <>
          <section className="app-panel overflow-hidden rounded-2xl border-[#d6e5f6] bg-white/94">
            <div className="flex flex-col gap-5 p-5 sm:flex-row sm:items-center sm:justify-between sm:p-7">
              <div className="flex min-w-0 items-center gap-4">
                <RainbowAvatar
                  alt={profile.full_name || profile.user_name}
                  fallback={profile.user_name.slice(0, 2).toUpperCase()}
                  size="lg"
                  src={profileAvatarUrl}
                />
                <div className="min-w-0">
                  <Badge>Creator</Badge>
                  <h1 className="mt-2 truncate text-3xl font-semibold tracking-normal text-[#143052]">
                    {profile.full_name || profile.user_name}
                  </h1>
                  <div className="mt-2 flex flex-wrap items-center gap-3 text-sm font-medium text-[#5f7f9f]">
                    <span>@{profile.user_name}</span>
                    <span className="inline-flex items-center gap-1.5">
                      <CalendarDays className="size-4" />
                      {formatJoinedDate(profile.created_at)}
                    </span>
                  </div>
                </div>
              </div>

              {canFollow ? (
                <Button
                  disabled={followMutation.isPending}
                  onClick={() => followMutation.mutate()}
                  type="button"
                  variant={profile.is_following ? "outline" : "default"}
                >
                  {followMutation.isPending ? (
                    <Loader2 className="size-4 animate-spin" />
                  ) : profile.is_following ? (
                    <UserX className="size-4" />
                  ) : (
                    <UserPlus className="size-4" />
                  )}
                  {profile.is_following ? "Unfollow" : "Follow"}
                </Button>
              ) : null}
            </div>

            <div className="grid border-[#dce8f4] border-t sm:grid-cols-3">
              {[
                ["Posts", profile.posts_count],
                ["Followers", profile.followers_count],
                ["Following", profile.following_count],
              ].map(([label, value]) => (
                <div
                  className="border-[#dce8f4] border-t px-5 py-4 first:border-t-0 sm:border-l sm:border-t-0 sm:first:border-l-0"
                  key={label}
                >
                  <p className="text-2xl font-semibold text-[#143052]">
                    {value}
                  </p>
                  <p className="mt-1 text-xs font-semibold uppercase tracking-[0.14em] text-[#7892ad]">
                    {label}
                  </p>
                </div>
              ))}
            </div>
          </section>

          {profile.posts.length > 0 ? (
            <section className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {profile.posts.map((post) => (
                <Link
                  className="group overflow-hidden rounded-2xl border border-[#d6e5f6] bg-white shadow-[0_18px_42px_-34px_rgb(32_72_116/0.7)] transition hover:-translate-y-0.5 hover:shadow-[0_22px_48px_-34px_rgb(32_72_116/0.9)]"
                  href={ROUTES.explore}
                  key={post.id}
                >
                  <div className="aspect-4/5 overflow-hidden bg-[#edf4fb]">
                    <img
                      alt={post.caption || post.location_name || "Post image"}
                      className="h-full w-full object-cover transition duration-500 group-hover:scale-[1.035]"
                      decoding="async"
                      fetchPriority="low"
                      loading="lazy"
                      sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 22rem"
                      src={post.image_url}
                    />
                  </div>
                  <div className="space-y-2 p-4">
                    <p className="line-clamp-2 text-sm font-semibold text-[#15365a]">
                      {post.caption || "Community post"}
                    </p>
                    <p className="text-xs font-medium text-[#6682a1]">
                      {post.location_name || "Uploaded"}
                      {post.category_name ? ` / ${post.category_name}` : ""}
                    </p>
                  </div>
                </Link>
              ))}
            </section>
          ) : (
            <EmptyState
              description="This creator has not published posts yet."
              icon={<ImageIcon className="size-5" />}
              title="No public posts"
            />
          )}
        </>
      )}
    </PageShell>
  );
}
