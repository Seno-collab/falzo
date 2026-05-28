import {
  BadgeCheck,
  Bookmark,
  Camera,
  LockKeyhole,
  MapPinned,
  ShieldCheck,
  Sparkles,
} from "lucide-react";
import type { ReactNode } from "react";
import { ScenicImage } from "@/components/scenic-image";

const previewFrames = [
  {
    id: "mu-cang-chai-dawn",
    label: "Saved route",
    className: "left-0 top-8 h-40 w-44 rotate-[-6deg]",
  },
  {
    id: "ly-son-coast",
    label: "Hidden cove",
    className: "right-0 top-0 h-52 w-36 rotate-[5deg]",
  },
  {
    id: "kyoto-lantern-night",
    label: "Night idea",
    className: "bottom-0 left-20 h-36 w-48 rotate-[2deg]",
  },
] as const;

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
  return (
    <main className="relative min-h-svh overflow-x-hidden bg-[#f7f7f5] text-[#1f1f1f]">
      <ScenicImage
        alt={title}
        className="absolute inset-0 h-full w-full scale-[1.03] object-cover opacity-[0.24]"
        fetchPriority="high"
        id="patagonia-trail"
        loading="eager"
        sizes="100vw"
      />
      <div className="absolute inset-0 bg-[linear-gradient(115deg,rgb(247_247_245/0.98)_0%,rgb(247_247_245/0.94)_42%,rgb(239_246_252/0.82)_72%,rgb(247_247_245/0.72)_100%)]" />
      <div className="absolute inset-0 bg-[repeating-linear-gradient(135deg,rgb(255_255_255/0.34)_0px,rgb(255_255_255/0.34)_1px,transparent_1px,transparent_18px)] opacity-55" />
      <div className="absolute inset-x-0 bottom-0 h-2/5 bg-[linear-gradient(0deg,rgb(215_229_244/0.68)_0%,rgb(247_247_245/0)_100%)]" />
      <div className="absolute right-0 top-0 hidden h-full w-[42vw] border-[#d6e5f6]/70 border-l bg-white/22 backdrop-blur-[1px] lg:block" />

      <div className="relative z-10 flex min-h-svh flex-col">
        <header className="px-4 py-3 sm:px-6 sm:py-4 lg:px-8">{topbar}</header>

        <section className="flex flex-1 items-start justify-center px-4 pb-[calc(env(safe-area-inset-bottom)+2rem)] pt-1 sm:items-center sm:px-6 sm:pb-8 sm:pt-4 lg:px-8">
          <div className="grid w-full max-w-5xl gap-6 lg:grid-cols-[minmax(0,440px)_minmax(0,1fr)] lg:items-center">
            <div>
              <div className="mb-5 text-center">
                <div className="mx-auto mb-3 inline-flex size-11 items-center justify-center rounded-full border border-[#c8ddf1] bg-white text-[#2f6fb8] shadow-[0_18px_38px_-28px_rgb(32_72_116/0.7)]">
                  <ShieldCheck className="size-5" />
                </div>
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[#7892ad]">
                  {label}
                </p>
                <h1 className="mt-2 text-4xl font-semibold tracking-normal text-[#143052]">
                  {title}
                </h1>
                <p className="mx-auto mt-2 max-w-sm text-sm leading-6 text-[#5f7894]">
                  {subtitle}
                </p>
              </div>

              <div className="rounded-3xl border border-[#d6e5f6] bg-white/90 p-5 shadow-[0_30px_86px_-46px_rgb(32_72_116/0.82)] backdrop-blur-2xl sm:p-6">
                <div className="space-y-6">
                  {children}
                  {footer ? <div>{footer}</div> : null}
                </div>
              </div>

              <div className="mx-auto mt-5 flex w-fit items-center gap-2 rounded-full border border-[#d6e5f6] bg-white/88 px-3 py-1.5 text-xs font-medium text-[#5f7894] shadow-[0_14px_34px_-28px_rgb(32_72_116/0.62)] backdrop-blur-xl">
                <BadgeCheck className="size-3.5 text-[#2f6fb8]" />
                {note ||
                  "Explore first. Save when a place feels worth remembering."}
              </div>
            </div>

            <aside className="hidden lg:block">
              <div className="relative min-h-[560px] overflow-hidden rounded-[2rem] border border-white/70 bg-[#102018] p-5 text-white shadow-[0_34px_90px_-54px_rgb(15_42_71/0.82)]">
                <ScenicImage
                  alt=""
                  className="absolute inset-0 h-full w-full scale-105 object-cover opacity-70"
                  id="swiss-lake-view"
                  loading="lazy"
                  sizes="50vw"
                />
                <div className="absolute inset-0 bg-[linear-gradient(180deg,rgb(7_20_18/0.28)_0%,rgb(7_20_18/0.72)_58%,rgb(7_20_18/0.9)_100%)]" />
                <div className="relative z-10 flex h-full min-h-[520px] flex-col justify-between">
                  <div className="flex items-center justify-between gap-3">
                    <div className="inline-flex items-center gap-2 rounded-full border border-white/22 bg-white/16 px-3 py-1.5 text-xs font-semibold text-white backdrop-blur-xl">
                      <Sparkles className="size-3.5 text-[#ffcf5a]" />
                      Members see more
                    </div>
                    <div className="inline-flex size-10 items-center justify-center rounded-full bg-white text-[#143052]">
                      <LockKeyhole className="size-4" />
                    </div>
                  </div>

                  <div className="relative my-8 min-h-[260px]">
                    {previewFrames.map((frame) => (
                      <div
                        className={`absolute overflow-hidden rounded-2xl border border-white/24 bg-white/14 p-1 shadow-[0_22px_48px_-30px_rgb(0_0_0/0.72)] backdrop-blur-sm ${frame.className}`}
                        key={frame.id}
                      >
                        <ScenicImage
                          alt=""
                          className="h-full w-full rounded-[0.85rem] object-cover"
                          id={frame.id}
                          loading="lazy"
                          sizes="220px"
                        />
                        <span className="absolute bottom-3 left-3 inline-flex items-center gap-1.5 rounded-full bg-white/90 px-2.5 py-1 text-[11px] font-bold text-[#143052]">
                          <Camera className="size-3" />
                          {frame.label}
                        </span>
                      </div>
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
                          className="flex items-center gap-3 rounded-2xl border border-white/14 bg-white/12 px-3 py-3 text-sm font-medium leading-5 text-white/88 backdrop-blur-xl"
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
