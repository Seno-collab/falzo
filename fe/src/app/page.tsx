"use client";

import {
  Bell,
  Bookmark,
  Camera,
  Grid3X3,
  Heart,
  Map,
  MapPinned,
  MessageCircle,
  PlaneTakeoff,
  Search,
  Share2,
  UserRound,
  X,
} from "lucide-react";
import { AnimatePresence, motion } from "motion/react";
import Link from "next/link";
import {
  useEffect,
  useState,
  type Dispatch,
  type ReactNode,
  type SetStateAction,
} from "react";
import { hasAuthSession } from "@/api/auth.api";
import { ScenicImage } from "@/components/scenic-image";
import { Button } from "@/components/ui/button";
import { ROUTES } from "@/lib/routes";
import { cn } from "@/lib/utils";

const travelPins = [
  {
    id: "highland-morning",
    imageId: "mu-cang-chai-dawn",
    location: "Mu Cang Chai, Vietnam",
    height: "h-[360px]",
    gradient: "bg-[linear-gradient(145deg,#f7d36e,#8dc871_48%,#346c42)]",
    mapX: 72,
    mapY: 43,
  },
  {
    id: "island-blue",
    imageId: "ly-son-coast",
    location: "Ly Son, Vietnam",
    height: "h-[470px]",
    gradient: "bg-[linear-gradient(145deg,#b9f3ff,#58a9d4_52%,#1f4a70)]",
    mapX: 70,
    mapY: 58,
  },
  {
    id: "lantern-street",
    imageId: "kyoto-lantern-night",
    location: "Kyoto, Japan",
    height: "h-[310px]",
    gradient: "bg-[linear-gradient(145deg,#ffd5a4,#e38a60_50%,#713d35)]",
    mapX: 78,
    mapY: 38,
  },
  {
    id: "alpine-lake",
    imageId: "swiss-lake-view",
    location: "Interlaken, Switzerland",
    height: "h-[520px]",
    gradient: "bg-[linear-gradient(145deg,#caffea,#64bbb3_50%,#1d5874)]",
    mapX: 48,
    mapY: 35,
  },
  {
    id: "turkiye-sunset",
    imageId: "istanbul-skyline",
    location: "Istanbul, Turkiye",
    height: "h-[390px]",
    gradient: "bg-[linear-gradient(145deg,#ffd8b1,#d58d50_50%,#62382e)]",
    mapX: 54,
    mapY: 43,
  },
  {
    id: "patagonia-path",
    imageId: "patagonia-trail",
    location: "Torres del Paine, Chile",
    height: "h-[450px]",
    gradient: "bg-[linear-gradient(145deg,#ddf4ff,#72acd2_52%,#264c70)]",
    mapX: 33,
    mapY: 76,
  },
  {
    id: "terrace-garden",
    imageId: "mu-cang-chai-dawn",
    location: "Sapa, Vietnam",
    height: "h-[420px]",
    gradient: "bg-[linear-gradient(145deg,#ffeaa4,#9ec96f_48%,#446f47)]",
    mapX: 71,
    mapY: 45,
  },
  {
    id: "coastal-quiet",
    imageId: "ly-son-coast",
    location: "Da Nang, Vietnam",
    height: "h-[330px]",
    gradient: "bg-[linear-gradient(145deg,#caf7ff,#6bbadd_50%,#2b5875)]",
    mapX: 72,
    mapY: 55,
  },
  {
    id: "old-town-glow",
    imageId: "kyoto-lantern-night",
    location: "Gion, Japan",
    height: "h-[480px]",
    gradient: "bg-[linear-gradient(145deg,#ffd9af,#d17a5c_50%,#593044)]",
    mapX: 79,
    mapY: 39,
  },
  {
    id: "lake-silence",
    imageId: "swiss-lake-view",
    location: "Lucerne, Switzerland",
    height: "h-[350px]",
    gradient: "bg-[linear-gradient(145deg,#ddfff0,#7fc6bf_50%,#38657c)]",
    mapX: 47,
    mapY: 36,
  },
  {
    id: "dome-horizon",
    imageId: "istanbul-skyline",
    location: "Galata, Turkiye",
    height: "h-[430px]",
    gradient: "bg-[linear-gradient(145deg,#ffe1bd,#c98255_52%,#523440)]",
    mapX: 53,
    mapY: 42,
  },
  {
    id: "wild-south",
    imageId: "patagonia-trail",
    location: "Aysen, Chile",
    height: "h-[300px]",
    gradient: "bg-[linear-gradient(145deg,#e3f6ff,#8ab7cf_48%,#344d63)]",
    mapX: 31,
    mapY: 74,
  },
] as const;

type TravelPin = (typeof travelPins)[number];
type SavedView = "grid" | "map";

export default function RootPage() {
  const [authenticated, setAuthenticated] = useState(false);
  const [savedPins, setSavedPins] = useState<Set<string>>(
    new Set(["island-blue", "alpine-lake", "old-town-glow", "patagonia-path"]),
  );
  const [likedPins, setLikedPins] = useState<Set<string>>(new Set());
  const [selectedPin, setSelectedPin] = useState<TravelPin | null>(null);
  const [savedView, setSavedView] = useState<SavedView>("grid");
  const [previewPinId, setPreviewPinId] = useState<string>("alpine-lake");

  useEffect(() => {
    document.title = "Falzo Travel | Visual Discovery";
    setAuthenticated(hasAuthSession());
  }, []);

  useEffect(() => {
    if (!selectedPin) {
      document.body.style.overflow = "";
      return;
    }

    document.body.style.overflow = "hidden";

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setSelectedPin(null);
      }
    }

    window.addEventListener("keydown", handleKeyDown);

    return () => {
      document.body.style.overflow = "";
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [selectedPin]);

  function toggleSet(
    setter: Dispatch<SetStateAction<Set<string>>>,
    id: string,
  ) {
    setter((current) => {
      const next = new Set(current);

      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }

      return next;
    });
  }

  const savedPlaces = travelPins.filter((pin) => savedPins.has(pin.id));
  const previewPin =
    savedPlaces.find((pin) => pin.id === previewPinId) ?? savedPlaces[0];

  return (
    <main className="min-h-screen bg-[#f8f7f4] text-[#171717]">
      <header className="sticky top-0 z-40 border-b border-black/5 bg-[#f8f7f4]/82 backdrop-blur-2xl">
        <div className="mx-auto flex max-w-385 items-center gap-2 px-3 py-3 sm:px-5 lg:px-8">
          <Link
            aria-label="Falzo Travel home"
            className="inline-flex size-10 items-center justify-center rounded-full bg-[#171717] text-white shadow-[0_14px_34px_-24px_rgb(0_0_0/0.8)] transition hover:scale-[1.03]"
            href={ROUTES.home}
          >
            <Camera className="size-4" />
          </Link>

          <div className="relative flex-1">
            <Search className="-translate-y-1/2 pointer-events-none absolute left-4 top-1/2 size-4 text-[#777]" />
            <input
              className="h-11 w-full rounded-full border border-black/6 bg-white/92 px-11 text-sm font-medium text-[#1e1e1e] shadow-[0_14px_36px_-30px_rgb(0_0_0/0.5)] outline-none transition placeholder:text-[#8b8b8b] focus:border-black/10 focus:bg-white focus:shadow-[0_22px_52px_-38px_rgb(0_0_0/0.7)]"
              placeholder="Search destinations"
              type="search"
            />
          </div>

          <div className="hidden items-center gap-1 sm:flex">
            <Button
              aria-label="Notifications"
              className="rounded-full"
              size="icon-sm"
              type="button"
              variant="ghost"
            >
              <Bell className="size-4" />
            </Button>
            <Button
              aria-label="Profile"
              asChild
              className="rounded-full"
              size="icon-sm"
              variant="outline"
            >
              <Link href={authenticated ? ROUTES.dashboard : ROUTES.login}>
                <UserRound className="size-4" />
              </Link>
            </Button>
          </div>
        </div>
      </header>

      <section className="mx-auto max-w-385 px-4 pb-5 pt-7 sm:px-6 lg:px-8">
        <div className="flex items-end justify-between gap-5">
          <div className="space-y-2">
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-[#777]">
              Travel inspiration
            </p>
            <h1 className="max-w-3xl text-4xl font-semibold tracking-normal text-[#111] sm:text-5xl">
              Explore places by image.
            </h1>
          </div>
          <Button
            asChild
            className="hidden rounded-full bg-[#171717] text-white hover:bg-[#2a2a2a] md:inline-flex"
          >
            <Link href={ROUTES.explore}>Explore</Link>
          </Button>
        </div>
      </section>

      <section className="mx-auto max-w-385 px-4 pb-16 sm:px-6 lg:px-8">
        <div className="columns-1 gap-4 sm:columns-2 lg:columns-3 2xl:columns-4">
          {travelPins.map((pin, index) => {
            const isSaved = savedPins.has(pin.id);
            const isLiked = likedPins.has(pin.id);

            return (
              <motion.article
                className="group relative mb-4 break-inside-avoid cursor-zoom-in overflow-hidden rounded-[30px] bg-white shadow-[0_18px_55px_-42px_rgb(0_0_0/0.65)] ring-1 ring-black/[0.04] transition duration-300 hover:-translate-y-1 hover:shadow-[0_30px_90px_-48px_rgb(0_0_0/0.78)]"
                initial={{ opacity: 0, y: 18 }}
                key={pin.id}
                onClick={() => setSelectedPin(pin)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    setSelectedPin(pin);
                  }
                }}
                role="button"
                tabIndex={0}
                transition={{
                  delay: Math.min(index * 0.035, 0.24),
                  duration: 0.36,
                  ease: "easeOut",
                }}
                viewport={{ amount: 0.1, once: true }}
                whileInView={{ opacity: 1, y: 0 }}
              >
                <div
                  className={cn(
                    "relative overflow-hidden",
                    pin.height,
                    pin.gradient,
                  )}
                >
                  <ScenicImage
                    alt={pin.location}
                    className="h-full w-full object-cover transition duration-700 ease-out group-hover:scale-[1.035]"
                    id={pin.imageId}
                    sizes="(max-width: 640px) 92vw, (max-width: 1024px) 46vw, (max-width: 1536px) 31vw, 23vw"
                  />
                  <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0)_0%,rgb(0_0_0/0)_46%,rgb(0_0_0/0.46)_100%)]" />

                  <div className="absolute inset-x-3 top-3 flex justify-end gap-2 opacity-0 transition duration-200 group-hover:opacity-100">
                    <button
                      aria-label={isLiked ? "Unlike" : "Like"}
                      className={cn(
                        "inline-flex size-9 items-center justify-center rounded-full bg-white/86 text-[#1f1f1f] shadow-sm backdrop-blur-xl transition hover:bg-white hover:scale-105",
                        isLiked ? "text-[#ff385c]" : "",
                      )}
                      onClick={(event) => {
                        event.stopPropagation();
                        toggleSet(setLikedPins, pin.id);
                      }}
                      type="button"
                    >
                      <Heart
                        className={cn("size-4", isLiked ? "fill-current" : "")}
                      />
                    </button>
                    <button
                      aria-label="Comment"
                      className="inline-flex size-9 items-center justify-center rounded-full bg-white/86 text-[#1f1f1f] shadow-sm backdrop-blur-xl transition hover:scale-105 hover:bg-white"
                      onClick={(event) => event.stopPropagation()}
                      type="button"
                    >
                      <MessageCircle className="size-4" />
                    </button>
                    <button
                      aria-label={isSaved ? "Remove save" : "Save"}
                      className={cn(
                        "inline-flex size-9 items-center justify-center rounded-full bg-white/86 text-[#1f1f1f] shadow-sm backdrop-blur-xl transition hover:bg-white hover:scale-105",
                        isSaved ? "text-[#ff385c]" : "",
                      )}
                      onClick={(event) => {
                        event.stopPropagation();
                        toggleSet(setSavedPins, pin.id);
                      }}
                      type="button"
                    >
                      <Bookmark
                        className={cn("size-4", isSaved ? "fill-current" : "")}
                      />
                    </button>
                  </div>

                  <div className="absolute bottom-3 left-3">
                    <span className="inline-flex max-w-[calc(100%-1.5rem)] items-center rounded-full bg-white/84 px-3 py-1.5 text-xs font-semibold text-[#202020] shadow-sm backdrop-blur-xl">
                      {pin.location}
                    </span>
                  </div>
                </div>
              </motion.article>
            );
          })}
        </div>
      </section>

      <section className="mx-auto max-w-[1540px] px-4 pb-20 sm:px-6 lg:px-8">
        <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div className="max-w-xl space-y-2">
            <p className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.18em] text-[#777]">
              <PlaneTakeoff className="size-3.5" />
              Saved places
            </p>
            <h2 className="text-3xl font-semibold tracking-normal text-[#111] sm:text-4xl">
              Your future trips, gathered in one quiet place.
            </h2>
          </div>

          <div className="inline-flex w-fit rounded-full border border-black/[0.06] bg-white p-1 shadow-[0_18px_44px_-36px_rgb(0_0_0/0.5)]">
            <button
              className={cn(
                "inline-flex h-9 items-center gap-2 rounded-full px-4 text-sm font-semibold transition",
                savedView === "grid"
                  ? "bg-[#171717] text-white"
                  : "text-[#666] hover:text-[#171717]",
              )}
              onClick={() => setSavedView("grid")}
              type="button"
            >
              <Grid3X3 className="size-4" />
              Grid
            </button>
            <button
              className={cn(
                "inline-flex h-9 items-center gap-2 rounded-full px-4 text-sm font-semibold transition",
                savedView === "map"
                  ? "bg-[#171717] text-white"
                  : "text-[#666] hover:text-[#171717]",
              )}
              onClick={() => setSavedView("map")}
              type="button"
            >
              <Map className="size-4" />
              Map
            </button>
          </div>
        </div>

        {savedPlaces.length > 0 ? (
          <AnimatePresence mode="wait">
            {savedView === "grid" ? (
              <motion.div
                animate={{ opacity: 1, y: 0 }}
                className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4"
                exit={{ opacity: 0, y: 10 }}
                initial={{ opacity: 0, y: 10 }}
                key="saved-grid"
                transition={{ duration: 0.24, ease: "easeOut" }}
              >
                {savedPlaces.map((pin) => (
                  <button
                    className="group overflow-hidden rounded-[28px] bg-white text-left shadow-[0_20px_58px_-42px_rgb(0_0_0/0.62)] ring-1 ring-black/[0.04] transition hover:-translate-y-1 hover:shadow-[0_30px_80px_-48px_rgb(0_0_0/0.72)]"
                    key={`saved-${pin.id}`}
                    onClick={() => setSelectedPin(pin)}
                    type="button"
                  >
                    <div
                      className={cn(
                        "relative h-72 overflow-hidden",
                        pin.gradient,
                      )}
                    >
                      <ScenicImage
                        alt={pin.location}
                        className="h-full w-full object-cover transition duration-700 group-hover:scale-[1.035]"
                        id={pin.imageId}
                        sizes="(max-width: 640px) 92vw, (max-width: 1024px) 46vw, 24vw"
                      />
                      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0)_0%,rgb(0_0_0/0.42)_100%)]" />
                      <span className="absolute bottom-3 left-3 rounded-full bg-white/86 px-3 py-1.5 text-xs font-semibold text-[#202020] backdrop-blur-xl">
                        {pin.location}
                      </span>
                    </div>
                  </button>
                ))}
              </motion.div>
            ) : (
              <motion.div
                animate={{ opacity: 1, y: 0 }}
                className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]"
                exit={{ opacity: 0, y: 10 }}
                initial={{ opacity: 0, y: 10 }}
                key="saved-map"
                transition={{ duration: 0.24, ease: "easeOut" }}
              >
                <div className="relative min-h-[520px] overflow-hidden rounded-[34px] border border-black/[0.06] bg-[#dfe9df] shadow-[0_28px_90px_-56px_rgb(0_0_0/0.65)]">
                  <div className="absolute inset-0 bg-[radial-gradient(circle_at_22%_30%,rgb(255_255_255/0.72)_0_7%,transparent_8%),radial-gradient(circle_at_68%_22%,rgb(255_255_255/0.5)_0_6%,transparent_7%),radial-gradient(circle_at_56%_74%,rgb(255_255_255/0.48)_0_9%,transparent_10%),linear-gradient(135deg,#e7efe5_0%,#d4e5d9_42%,#c7d9df_100%)]" />
                  <div className="absolute left-[12%] top-[18%] h-[72%] w-[68%] rotate-[-10deg] rounded-[55%] border border-white/55 bg-[#c7dac4]/68 blur-[1px]" />
                  <div className="absolute left-[43%] top-[8%] h-[78%] w-[38%] rotate-[18deg] rounded-[48%] border border-white/40 bg-[#c3d8e1]/72 blur-[1px]" />
                  <div className="absolute inset-0 opacity-45 [background-image:linear-gradient(rgba(255,255,255,0.42)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.42)_1px,transparent_1px)] [background-size:42px_42px]" />
                  <div className="absolute inset-x-0 top-0 h-24 bg-[linear-gradient(180deg,rgb(255_255_255/0.58),transparent)]" />
                  <div className="absolute bottom-5 left-5 rounded-full bg-white/72 px-3 py-1.5 text-xs font-semibold text-[#47604d] shadow-sm backdrop-blur-xl">
                    {savedPlaces.length} saved pins
                  </div>

                  {savedPlaces.map((pin) => {
                    const isPreview = previewPin?.id === pin.id;

                    return (
                      <button
                        aria-label={`Preview ${pin.location}`}
                        className="absolute -translate-x-1/2 -translate-y-1/2"
                        key={`map-${pin.id}`}
                        onClick={() => setPreviewPinId(pin.id)}
                        onMouseEnter={() => setPreviewPinId(pin.id)}
                        style={{ left: `${pin.mapX}%`, top: `${pin.mapY}%` }}
                        type="button"
                      >
                        <span
                          className={cn(
                            "absolute left-1/2 top-1/2 size-10 -translate-x-1/2 -translate-y-1/2 rounded-full bg-[#ff385c]/18 transition",
                            isPreview
                              ? "scale-125 opacity-100"
                              : "scale-75 opacity-0",
                          )}
                        />
                        <span
                          className={cn(
                            "relative inline-flex size-9 items-center justify-center rounded-full bg-[#171717] text-white shadow-[0_14px_34px_-18px_rgb(0_0_0/0.85)] ring-4 ring-white/78 transition hover:scale-110",
                            isPreview ? "scale-110 bg-[#ff385c]" : "",
                          )}
                        >
                          <MapPinned className="size-4" />
                        </span>
                      </button>
                    );
                  })}
                </div>

                <aside className="overflow-hidden rounded-[34px] bg-[#171717] text-white shadow-[0_28px_90px_-56px_rgb(0_0_0/0.82)]">
                  {previewPin ? (
                    <div>
                      <div
                        className={cn(
                          "relative h-80 overflow-hidden",
                          previewPin.gradient,
                        )}
                      >
                        <ScenicImage
                          alt={previewPin.location}
                          className="h-full w-full object-cover"
                          fetchPriority="high"
                          id={previewPin.imageId}
                          loading="eager"
                          sizes="360px"
                        />
                        <div className="absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0)_0%,rgb(0_0_0/0.58)_100%)]" />
                        <span className="absolute bottom-4 left-4 rounded-full bg-white/88 px-3 py-1.5 text-xs font-semibold text-[#171717] backdrop-blur-xl">
                          Preview
                        </span>
                      </div>
                      <div className="space-y-4 p-5">
                        <div>
                          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-white/48">
                            Wishlist stop
                          </p>
                          <h3 className="mt-1 text-2xl font-semibold tracking-normal">
                            {previewPin.location}
                          </h3>
                        </div>
                        <div className="flex gap-2">
                          <Button
                            className="rounded-full border-white/16 bg-white/10 text-white hover:bg-white/16"
                            onClick={() => setSelectedPin(previewPin)}
                            type="button"
                            variant="outline"
                          >
                            Open
                          </Button>
                          <Button
                            className="rounded-full bg-white text-[#171717] hover:bg-white/90"
                            type="button"
                          >
                            <MapPinned className="size-4" />
                            Plan route
                          </Button>
                        </div>
                      </div>
                    </div>
                  ) : null}
                </aside>
              </motion.div>
            )}
          </AnimatePresence>
        ) : (
          <div className="rounded-[30px] border border-dashed border-black/12 bg-white/70 p-10 text-center text-sm font-medium text-[#777]">
            Save a destination to start building your travel wishlist.
          </div>
        )}
      </section>

      <AnimatePresence>
        {selectedPin ? (
          <motion.div
            animate={{ opacity: 1 }}
            className="fixed inset-0 z-50 grid place-items-center bg-black/88 px-4 py-5 text-white backdrop-blur-md sm:px-6"
            exit={{ opacity: 0 }}
            initial={{ opacity: 0 }}
            onClick={() => setSelectedPin(null)}
            role="dialog"
            aria-modal="true"
            aria-label={`${selectedPin.location} image viewer`}
            transition={{ duration: 0.22, ease: "easeOut" }}
          >
            <button
              aria-label="Close image viewer"
              className="absolute right-4 top-4 z-20 inline-flex size-10 items-center justify-center rounded-full bg-white/10 text-white backdrop-blur-xl transition hover:bg-white/18 sm:right-6 sm:top-6"
              onClick={() => setSelectedPin(null)}
              type="button"
            >
              <X className="size-5" />
            </button>

            <motion.div
              animate={{ opacity: 1, scale: 1, y: 0 }}
              className="relative h-[88vh] w-full max-w-295 overflow-hidden rounded-[34px] bg-neutral-900 shadow-[0_44px_120px_-42px_rgb(0_0_0/0.95)] ring-1 ring-white/10"
              exit={{ opacity: 0, scale: 0.97, y: 12 }}
              initial={{ opacity: 0, scale: 0.94, y: 18 }}
              onClick={(event) => event.stopPropagation()}
              transition={{ duration: 0.34, ease: [0.16, 1, 0.3, 1] }}
            >
              <motion.div
                animate={{ scale: 1.025 }}
                className={cn("absolute inset-0", selectedPin.gradient)}
                initial={{ scale: 1 }}
                transition={{ duration: 0.7, ease: "easeOut" }}
              >
                <ScenicImage
                  alt={selectedPin.location}
                  className="h-full w-full object-cover"
                  fetchPriority="high"
                  id={selectedPin.imageId}
                  loading="eager"
                  sizes="100vw"
                />
              </motion.div>
              <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(90deg,rgb(0_0_0/0.24)_0%,rgb(0_0_0/0)_28%,rgb(0_0_0/0.18)_100%),linear-gradient(180deg,rgb(0_0_0/0)_0%,rgb(0_0_0/0)_50%,rgb(0_0_0/0.68)_100%)]" />

              <div className="-translate-y-1/2 absolute right-3 top-1/2 flex flex-col gap-3 sm:right-5">
                <ViewerAction
                  active={likedPins.has(selectedPin.id)}
                  icon={<Heart className="size-5" />}
                  label="Like"
                  onClick={() => toggleSet(setLikedPins, selectedPin.id)}
                />
                <ViewerAction
                  icon={<MessageCircle className="size-5" />}
                  label="Comment"
                />
                <ViewerAction
                  icon={<Share2 className="size-5" />}
                  label="Share"
                />
                <ViewerAction
                  active={savedPins.has(selectedPin.id)}
                  icon={<Bookmark className="size-5" />}
                  label="Save"
                  onClick={() => toggleSet(setSavedPins, selectedPin.id)}
                />
              </div>

              <motion.div
                animate={{ opacity: 1, y: 0 }}
                className="absolute inset-x-0 bottom-0 flex flex-col gap-3 p-5 sm:flex-row sm:items-end sm:justify-between sm:p-7"
                initial={{ opacity: 0, y: 16 }}
                transition={{ delay: 0.12, duration: 0.28, ease: "easeOut" }}
              >
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.2em] text-white/62">
                    Now viewing
                  </p>
                  <h2 className="mt-1 text-3xl font-semibold tracking-normal text-white sm:text-5xl">
                    {selectedPin.location}
                  </h2>
                </div>
                <Button
                  className="w-fit rounded-full border-white/18 bg-white/12 text-white backdrop-blur-xl hover:bg-white/18"
                  type="button"
                  variant="outline"
                >
                  <MapPinned className="size-4" />
                  View on map
                </Button>
              </motion.div>
            </motion.div>
          </motion.div>
        ) : null}
      </AnimatePresence>
    </main>
  );
}

function ViewerAction({
  active = false,
  icon,
  label,
  onClick,
}: Readonly<{
  active?: boolean;
  icon: ReactNode;
  label: string;
  onClick?: () => void;
}>) {
  return (
    <button
      aria-label={label}
      className={cn(
        "group/action flex flex-col items-center gap-1 text-[11px] font-semibold text-white/80 transition hover:text-white",
        active ? "text-[#ff4968]" : "",
      )}
      onClick={onClick}
      type="button"
    >
      <span
        className={cn(
          "inline-flex size-12 items-center justify-center rounded-full bg-white/12 shadow-[0_18px_38px_-24px_rgb(0_0_0/0.9)] backdrop-blur-xl transition group-hover/action:scale-105 group-hover/action:bg-white/20",
          active ? "bg-white text-[#ff385c]" : "",
        )}
      >
        {icon}
      </span>
      {label}
    </button>
  );
}
