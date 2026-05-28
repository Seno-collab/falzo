"use client";

import { useQuery } from "@tanstack/react-query";
import {
  BadgeCheck,
  Bookmark,
  Camera,
  LockKeyhole,
  MapPinned,
  ShieldCheck,
  Sparkles,
} from "lucide-react";
import { useEffect, useMemo, useState, type ReactNode } from "react";
import { getPostsPageApi } from "@/features/posts/api";

type AuthCarouselFrame = {
  id: string;
  title: string;
  location: string;
  src: string;
};

const fallbackAuthCarouselFrames: AuthCarouselFrame[] = [
  {
    id: "alpine-lake",
    title: "Alpine lake",
    location: "Interlaken, Switzerland",
    src: "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&fm=jpg&q=80&w=1800",
  },
  {
    id: "coastal-cliff",
    title: "Coastal cliff",
    location: "Amalfi Coast, Italy",
    src: "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?auto=format&fit=crop&fm=jpg&q=80&w=1800",
  },
  {
    id: "mountain-road",
    title: "Mountain road",
    location: "Dolomites, Italy",
    src: "https://images.unsplash.com/photo-1501785888041-af3ef285b470?auto=format&fit=crop&fm=jpg&q=80&w=1800",
  },
  {
    id: "forest-valley",
    title: "Forest valley",
    location: "Pacific Northwest, USA",
    src: "https://images.unsplash.com/photo-1469474968028-56623f02e42e?auto=format&fit=crop&fm=jpg&q=80&w=1800",
  },
] as const satisfies AuthCarouselFrame[];

const authCarouselIntervalMs = 4_800;

export function AuthShell({
  topbar,
  label,
  title,
  subtitle,
  points,
  note,
  children,
  footer,
}: Readonly<{
  topbar: ReactNode;
  label: string;
  title: string;
  subtitle: string;
  points: string[];
  note: string;
  children: ReactNode;
  footer?: ReactNode;
}>) {
  const [activeFrameIndex, setActiveFrameIndex] = useState(0);
  const exploreImagesQuery = useQuery({
    queryKey: ["posts", "auth-carousel-images"],
    queryFn: ({ signal }) =>
      getPostsPageApi({
        limit: 8,
        sort: "popular",
        signal,
      }),
    retry: 1,
    staleTime: 60_000,
  });
  const exploreFrames = useMemo<AuthCarouselFrame[]>(
    () =>
      (exploreImagesQuery.data?.items ?? [])
        .filter((post) => post.image_url.trim())
        .slice(0, 4)
        .map((post) => ({
          id: `post-${post.id}`,
          title: post.caption || post.location_name || "Explore post",
          location: post.location_name || "Falzo Explore",
          src: post.image_url,
        })),
    [exploreImagesQuery.data?.items],
  );
  const authCarouselFrames =
    exploreFrames.length >= 3 ? exploreFrames : fallbackAuthCarouselFrames;
  const activeFrame = authCarouselFrames[activeFrameIndex];
  const previewFrames = useMemo(
    () =>
      authCarouselFrames.map((_, index) => {
        const originalIndex =
          (activeFrameIndex + index) % authCarouselFrames.length;

        return {
          frame: authCarouselFrames[originalIndex],
          originalIndex,
        };
      }),
    [activeFrameIndex, authCarouselFrames],
  );

  useEffect(() => {
    if (activeFrameIndex < authCarouselFrames.length) {
      return;
    }

    setActiveFrameIndex(0);
  }, [activeFrameIndex, authCarouselFrames.length]);

  useEffect(() => {
    if (authCarouselFrames.length <= 1) {
      return undefined;
    }

    const interval = globalThis.setInterval(() => {
      setActiveFrameIndex(
        (currentIndex) => (currentIndex + 1) % authCarouselFrames.length,
      );
    }, authCarouselIntervalMs);

    return () => globalThis.clearInterval(interval);
  }, [authCarouselFrames.length]);

  return (
    <main className="relative min-h-svh overflow-x-hidden bg-[#f7f7f5] text-[#1f1f1f]">
      <img
        alt={title}
        className="absolute inset-0 h-full w-full scale-[1.03] object-cover opacity-[0.24]"
        decoding="async"
        fetchPriority="high"
        loading="eager"
        sizes="100vw"
        src={activeFrame.src}
      />
      <div className="absolute inset-0 bg-[linear-gradient(115deg,rgb(247_247_245/0.98)_0%,rgb(247_247_245/0.94)_42%,rgb(239_246_252/0.82)_72%,rgb(247_247_245/0.72)_100%)]" />
      <div className="absolute inset-0 bg-[repeating-linear-gradient(135deg,rgb(255_255_255/0.34)_0px,rgb(255_255_255/0.34)_1px,transparent_1px,transparent_18px)] opacity-55" />
      <div className="absolute inset-x-0 bottom-0 h-2/5 bg-[linear-gradient(0deg,rgb(215_229_244/0.68)_0%,rgb(247_247_245/0)_100%)]" />
      <div className="absolute right-0 top-0 hidden h-full w-[42vw] border-[#d6e5f6]/70 border-l bg-white/22 backdrop-blur-[1px] lg:block" />

      <div className="relative z-10 flex min-h-svh flex-col">
        <header className="px-4 py-3 sm:px-6 sm:py-4 lg:px-8">{topbar}</header>

        <section className="flex flex-1 items-start justify-center px-4 pb-[calc(env(safe-area-inset-bottom)+2rem)] pt-1 sm:items-center sm:px-6 sm:pb-8 sm:pt-4 lg:px-8">
          <div className="grid w-full max-w-5xl gap-6 lg:grid-cols-[minmax(0,440px)_minmax(0,1fr)] lg:items-center">
            <div className="lg:translate-y-3">
              <div className="mb-4 text-center">
                <div className="mx-auto mb-2.5 inline-flex size-9 items-center justify-center rounded-full bg-white/70 text-[#2f6fb8]">
                  <ShieldCheck className="size-4" />
                </div>
                <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-[#7892ad]">
                  {label}
                </p>
                <h1 className="mt-1.5 text-2xl font-semibold tracking-normal text-[#143052] sm:text-[1.75rem]">
                  {title}
                </h1>
                <p className="mx-auto mt-1.5 max-w-xs text-xs leading-5 text-[#5f7894]">
                  {subtitle}
                </p>
              </div>

              <div className="rounded-3xl border border-[#d6e5f6] bg-white/90 p-5 shadow-[0_30px_86px_-46px_rgb(32_72_116/0.82)] backdrop-blur-2xl sm:p-6">
                <div className="space-y-6">
                  {children}
                  {footer ? <div>{footer}</div> : null}
                </div>
              </div>

              <div className="mx-auto mt-5 flex w-fit items-center gap-2 rounded-full bg-white/60 px-3 py-1.5 text-xs font-medium text-[#5f7894] backdrop-blur-xl">
                <BadgeCheck className="size-3.5 text-[#2f6fb8]" />
                {note ||
                  "Explore first. Save when a place feels worth remembering."}
              </div>
            </div>

            <aside className="hidden lg:block">
              <div className="relative min-h-[560px] overflow-hidden rounded-[2rem] border border-white/70 bg-[#102018] p-5 text-white shadow-[0_34px_90px_-54px_rgb(15_42_71/0.82)]">
                {authCarouselFrames.map((frame, index) => (
                  <img
                    alt=""
                    aria-hidden="true"
                    className={`absolute inset-0 h-full w-full scale-105 object-cover transition-opacity duration-700 ${
                      index === activeFrameIndex
                        ? "opacity-[0.72]"
                        : "opacity-0"
                    }`}
                    decoding="async"
                    key={frame.id}
                    loading={index === 0 ? "eager" : "lazy"}
                    sizes="50vw"
                    src={frame.src}
                  />
                ))}
                <div className="absolute inset-0 bg-[linear-gradient(180deg,rgb(7_20_18/0.28)_0%,rgb(7_20_18/0.72)_58%,rgb(7_20_18/0.9)_100%)]" />
                <div className="relative z-10 flex h-full min-h-[520px] flex-col justify-between">
                  <div className="flex items-center justify-between gap-3">
                    <div className="inline-flex items-center gap-2 rounded-full bg-white/16 px-3 py-1.5 text-xs font-semibold text-white backdrop-blur-xl">
                      <Sparkles className="size-3.5 text-[#ffcf5a]" />
                      Members see more
                    </div>
                    <div className="inline-flex size-10 items-center justify-center rounded-full bg-white text-[#143052]">
                      <LockKeyhole className="size-4" />
                    </div>
                  </div>

                  <div className="relative my-8 min-h-[260px]">
                    {previewFrames
                      .slice(0, 3)
                      .map(({ frame, originalIndex }, index) => (
                      <button
                        aria-label={`View ${frame.title}`}
                        className={`absolute overflow-hidden rounded-2xl border bg-white/14 p-1 text-left shadow-[0_22px_48px_-30px_rgb(0_0_0/0.72)] backdrop-blur-sm transition duration-300 hover:-translate-y-1 ${
                          index === 0
                            ? "left-0 top-8 h-40 w-44 rotate-[-6deg] border-white/70"
                            : index === 1
                              ? "right-0 top-0 h-52 w-36 rotate-[5deg] border-white/24"
                              : "bottom-0 left-20 h-36 w-48 rotate-[2deg] border-white/24"
                        }`}
                        key={frame.id}
                        onClick={() => setActiveFrameIndex(originalIndex)}
                        type="button"
                      >
                        <img
                          alt=""
                          className="h-full w-full rounded-[0.85rem] object-cover"
                          decoding="async"
                          loading="lazy"
                          sizes="220px"
                          src={frame.src}
                        />
                        {/* <span className="absolute bottom-3 left-3 inline-flex items-center gap-1.5 rounded-full bg-white/90 px-2.5 py-1 text-[11px] font-bold text-[#143052]">
                          <Camera className="size-3" />
                          {frame.title}
                        </span> */}
                      </button>
                    ))}
                  </div>

                  <div className="space-y-4">
                    <div>
                      <p className="text-xs font-bold uppercase tracking-[0.18em] text-white/62">
                        After access
                      </p>
                      <h2 className="mt-2 max-w-sm text-3xl font-semibold leading-tight tracking-normal">
                        Turn browsing into a personal travel board.
                      </h2>
                      <p className="mt-2 text-sm font-medium text-white/72">
                        Now showing {activeFrame.location}
                      </p>
                    </div>

                    <div className="grid gap-2">
                      {(points.length > 0
                        ? points
                        : [
                            "Save destinations that catch your eye.",
                            "Follow fresh posts from other travelers.",
                            "Build a board before planning the trip.",
                          ]
                      ).map((point, index) => (
                        <div
                          className="flex items-center gap-3 rounded-2xl bg-white/12 px-3 py-3 text-sm font-medium leading-5 text-white/88 backdrop-blur-xl"
                          key={point}
                        >
                          <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-white text-[#143052]">
                            {index === 0 ? (
                              <Bookmark className="size-4" />
                            ) : index === 1 ? (
                              <MapPinned className="size-4" />
                            ) : (
                              <Sparkles className="size-4" />
                            )}
                          </span>
                          {point}
                        </div>
                      ))}
                    </div>

                    <div className="flex items-center gap-2">
                      {authCarouselFrames.map((frame, index) => (
                        <button
                          aria-label={`Show ${frame.title}`}
                          className={`h-2.5 rounded-full transition-all ${
                            index === activeFrameIndex
                              ? "w-8 bg-white"
                              : "w-2.5 bg-white/38 hover:bg-white/68"
                          }`}
                          key={frame.id}
                          onClick={() => setActiveFrameIndex(index)}
                          type="button"
                        />
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </aside>
          </div>
        </section>
      </div>
    </main>
  );
}
