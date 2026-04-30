"use client";

import {
  Bell,
  Bookmark,
  Camera,
  ChevronDown,
  Heart,
  Home,
  Menu,
  Plus,
  Search,
  SlidersHorizontal,
  Sparkles,
  UserRound,
} from "lucide-react";
import { motion } from "motion/react";
import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { ScenicImage } from "@/components/scenic-image";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const collections = [
  "All",
  "Homes",
  "Travel",
  "Food",
  "Style",
  "Wellness",
  "Architecture",
  "Outdoors",
] as const;

const pins = [
  {
    id: "rice-terrace-light",
    imageId: "mu-cang-chai-dawn",
    title: "Morning terraces",
    author: "Nora Field",
    city: "Mu Cang Chai",
    collection: "Travel",
    saves: "18.4k",
    height: "h-[340px]",
    gradient: "bg-[linear-gradient(145deg,#f6d36b,#80b96e_50%,#2f6d51)]",
  },
  {
    id: "coastal-villa",
    imageId: "ly-son-coast",
    title: "Blue coast hideaway",
    author: "Casa Atelier",
    city: "Ly Son",
    collection: "Homes",
    saves: "9.7k",
    height: "h-[430px]",
    gradient: "bg-[linear-gradient(145deg,#95e4f6,#3f92c9_54%,#173d66)]",
  },
  {
    id: "lantern-dinner",
    imageId: "kyoto-lantern-night",
    title: "Lantern dinner room",
    author: "Mika Studio",
    city: "Kyoto",
    collection: "Food",
    saves: "12.1k",
    height: "h-[300px]",
    gradient: "bg-[linear-gradient(145deg,#ffd0a1,#d77a58_54%,#5e2f36)]",
  },
  {
    id: "lake-spa",
    imageId: "swiss-lake-view",
    title: "Lakehouse reset",
    author: "Alpine Edit",
    city: "Interlaken",
    collection: "Wellness",
    saves: "22.9k",
    height: "h-[470px]",
    gradient: "bg-[linear-gradient(145deg,#c7ffe6,#63b9b4_52%,#1f5c78)]",
  },
  {
    id: "sunset-domes",
    imageId: "istanbul-skyline",
    title: "Copper hour skyline",
    author: "Ayla Notes",
    city: "Istanbul",
    collection: "Architecture",
    saves: "7.6k",
    height: "h-[360px]",
    gradient: "bg-[linear-gradient(145deg,#ffd7ad,#d48a4d_54%,#623729)]",
  },
  {
    id: "patagonia-layer",
    imageId: "patagonia-trail",
    title: "Patagonia trail layers",
    author: "Wilder Co.",
    city: "Torres del Paine",
    collection: "Outdoors",
    saves: "15.8k",
    height: "h-[410px]",
    gradient: "bg-[linear-gradient(145deg,#d8f2ff,#6aa7d2_52%,#244c72)]",
  },
  {
    id: "linen-suite",
    imageId: "swiss-lake-view",
    title: "Quiet linen suite",
    author: "Room Service",
    city: "Lucerne",
    collection: "Style",
    saves: "6.3k",
    height: "h-[320px]",
    gradient: "bg-[linear-gradient(145deg,#f8f2e8,#c7d4cc_50%,#718986)]",
  },
  {
    id: "market-table",
    imageId: "kyoto-lantern-night",
    title: "Market table palette",
    author: "Supper Club",
    city: "Osaka",
    collection: "Food",
    saves: "10.2k",
    height: "h-[390px]",
    gradient: "bg-[linear-gradient(145deg,#ffddb5,#f09268_50%,#884437)]",
  },
  {
    id: "pool-edge",
    imageId: "ly-son-coast",
    title: "Pool edge afternoon",
    author: "Stay Index",
    city: "Da Nang",
    collection: "Homes",
    saves: "13.5k",
    height: "h-[450px]",
    gradient: "bg-[linear-gradient(145deg,#b6f2ff,#5aaed6_50%,#295b76)]",
  },
  {
    id: "glass-cabin",
    imageId: "patagonia-trail",
    title: "Glass cabin plans",
    author: "Northline",
    city: "Aysen",
    collection: "Architecture",
    saves: "8.9k",
    height: "h-[335px]",
    gradient: "bg-[linear-gradient(145deg,#e0f4ff,#8ab7ce_48%,#314a5f)]",
  },
  {
    id: "garden-breakfast",
    imageId: "mu-cang-chai-dawn",
    title: "Garden breakfast",
    author: "Soft Serve",
    city: "Sapa",
    collection: "Wellness",
    saves: "11.6k",
    height: "h-[380px]",
    gradient: "bg-[linear-gradient(145deg,#fff0a8,#91c777_50%,#3f7350)]",
  },
  {
    id: "evening-walk",
    imageId: "istanbul-skyline",
    title: "Evening walk edit",
    author: "Rumi Lane",
    city: "Galata",
    collection: "Travel",
    saves: "19.1k",
    height: "h-[485px]",
    gradient: "bg-[linear-gradient(145deg,#ffd9b8,#bd7654_52%,#51333f)]",
  },
] as const;

export default function ExploreRoutePage() {
  const [activeCollection, setActiveCollection] =
    useState<(typeof collections)[number]>("All");
  const [savedPins, setSavedPins] = useState<Set<string>>(new Set());

  useEffect(() => {
    document.title = "Falzo Explore | Visual Inspiration";
  }, []);

  const visiblePins = useMemo(() => {
    if (activeCollection === "All") {
      return pins;
    }

    return pins.filter((pin) => pin.collection === activeCollection);
  }, [activeCollection]);

  function toggleSaved(id: string) {
    setSavedPins((current) => {
      const next = new Set(current);

      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }

      return next;
    });
  }

  return (
    <main className="min-h-screen bg-[#f7f7f5] text-[#1f1f1f]">
      <header className="sticky top-0 z-40 border-b border-black/6 bg-[#f7f7f5]/86 backdrop-blur-2xl">
        <div className="mx-auto flex w-full max-w-370 items-center gap-2 px-3 py-3 sm:px-5 lg:px-8">
          <Link
            aria-label="Home"
            className="inline-flex size-10 items-center justify-center rounded-full bg-[#111] text-white shadow-[0_14px_30px_-20px_rgb(0_0_0/0.72)] transition hover:scale-[1.03]"
            href="/"
          >
            <Camera className="size-4" />
          </Link>

          <nav className="hidden items-center gap-1 md:flex">
            <Button asChild className="rounded-full" size="sm" variant="ghost">
              <Link href="/">
                <Home className="size-4" />
                Home
              </Link>
            </Button>
            <Button
              className="rounded-full bg-[#111] text-white hover:bg-[#222]"
              size="sm"
            >
              Explore
            </Button>
          </nav>

          <div className="relative ml-1 flex-1">
            <Search className="-translate-y-1/2 pointer-events-none absolute left-4 top-1/2 size-4 text-[#777]" />
            <input
              className="h-11 w-full rounded-full border border-black/6 bg-white px-11 text-sm text-[#1f1f1f] shadow-[0_12px_32px_-28px_rgb(0_0_0/0.45)] outline-none transition placeholder:text-[#8a8a8a] focus:border-black/10 focus:bg-white focus:shadow-[0_18px_40px_-30px_rgb(0_0_0/0.58)]"
              placeholder="Search places, rooms, tables, textures"
              type="search"
            />
            <Button
              aria-label="Search filters"
              className="-translate-y-1/2 absolute right-1.5 top-1/2 rounded-full"
              size="icon-sm"
              type="button"
              variant="ghost"
            >
              <SlidersHorizontal className="size-4" />
            </Button>
          </div>

          <div className="hidden items-center gap-1 sm:flex">
            <Button
              aria-label="Create"
              className="rounded-full"
              size="icon-sm"
              type="button"
              variant="ghost"
            >
              <Plus className="size-4" />
            </Button>
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
              className="rounded-full"
              size="icon-sm"
              type="button"
              variant="outline"
            >
              <UserRound className="size-4" />
            </Button>
          </div>

          <Button
            aria-label="Menu"
            className="rounded-full sm:hidden"
            size="icon-sm"
            type="button"
            variant="ghost"
          >
            <Menu className="size-4" />
          </Button>
        </div>
      </header>

      <section className="mx-auto w-full max-w-[1480px] px-4 pb-4 pt-6 sm:px-6 lg:px-8">
        <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
          <div className="max-w-3xl">
            <div className="mb-3 inline-flex items-center gap-2 rounded-full border border-black/[0.06] bg-white px-3 py-1.5 text-xs font-semibold text-[#555] shadow-[0_12px_30px_-24px_rgb(0_0_0/0.35)]">
              <Sparkles className="size-3.5 text-[#ff385c]" />
              Curated today
            </div>
            <h1 className="max-w-2xl text-4xl font-semibold tracking-normal text-[#111] sm:text-5xl lg:text-6xl">
              Fresh visual ideas for beautiful stays and memorable travel.
            </h1>
          </div>

          <div className="flex items-center gap-2 lg:justify-end">
            <Button
              className="rounded-full border-black/[0.08] bg-white"
              type="button"
              variant="outline"
            >
              Trending
              <ChevronDown className="size-4" />
            </Button>
            <Button
              className="rounded-full bg-[#ff385c] text-white shadow-[0_18px_38px_-24px_rgb(255_56_92/0.8)] hover:bg-[#e93152]"
              type="button"
            >
              <Bookmark className="size-4" />
              Board
            </Button>
          </div>
        </div>

        <div className="mt-6 flex gap-2 overflow-x-auto pb-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          {collections.map((collection) => (
            <button
              className={cn(
                "shrink-0 rounded-full border px-4 py-2 text-sm font-semibold transition",
                activeCollection === collection
                  ? "border-[#111] bg-[#111] text-white shadow-[0_16px_32px_-24px_rgb(0_0_0/0.75)]"
                  : "border-black/[0.07] bg-white text-[#444] hover:border-black/15 hover:bg-[#fbfbfa]",
              )}
              key={collection}
              onClick={() => setActiveCollection(collection)}
              type="button"
            >
              {collection}
            </button>
          ))}
        </div>
      </section>

      <section className="mx-auto w-full max-w-[1480px] px-4 pb-14 sm:px-6 lg:px-8">
        <div className="columns-1 gap-4 sm:columns-2 lg:columns-3 2xl:columns-4">
          {visiblePins.map((pin, index) => {
            const isSaved = savedPins.has(pin.id);

            return (
              <motion.article
                className="group mb-4 break-inside-avoid overflow-hidden rounded-[28px] border border-black/[0.05] bg-white shadow-[0_18px_48px_-38px_rgb(0_0_0/0.6)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_26px_70px_-42px_rgb(0_0_0/0.72)]"
                initial={{ opacity: 0, y: 18 }}
                key={pin.id}
                transition={{
                  duration: 0.34,
                  delay: Math.min(index * 0.035, 0.22),
                  ease: "easeOut",
                }}
                viewport={{ amount: 0.12, once: true }}
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
                    alt={pin.title}
                    className="h-full w-full object-cover transition duration-500 ease-out group-hover:scale-[1.035]"
                    id={pin.imageId}
                    sizes="(max-width: 640px) 92vw, (max-width: 1024px) 46vw, (max-width: 1536px) 31vw, 23vw"
                  />
                  <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0.02)_0%,rgb(0_0_0/0.02)_48%,rgb(0_0_0/0.44)_100%)] opacity-80 transition group-hover:opacity-100" />
                  <div className="absolute left-3 right-3 top-3 flex items-start justify-between gap-2 opacity-0 transition duration-200 group-hover:opacity-100">
                    <span className="rounded-full bg-white/86 px-3 py-1 text-xs font-semibold text-[#222] shadow-sm backdrop-blur-xl">
                      {pin.collection}
                    </span>
                    <Button
                      aria-label={isSaved ? "Remove save" : "Save"}
                      className={cn(
                        "rounded-full shadow-sm backdrop-blur-xl",
                        isSaved
                          ? "bg-[#ff385c] text-white hover:bg-[#e93152]"
                          : "bg-white/86 text-[#222] hover:bg-white",
                      )}
                      onClick={() => toggleSaved(pin.id)}
                      size="icon-sm"
                      type="button"
                      variant="ghost"
                    >
                      <Heart
                        className={cn("size-4", isSaved ? "fill-current" : "")}
                      />
                    </Button>
                  </div>
                  <div className="absolute inset-x-4 bottom-4 text-white">
                    <p className="text-xs font-semibold uppercase tracking-[0.16em] text-white/76">
                      {pin.city}
                    </p>
                    <h2 className="mt-1 text-2xl font-semibold tracking-normal">
                      {pin.title}
                    </h2>
                  </div>
                </div>

                <div className="flex items-center justify-between gap-3 p-4">
                  <div className="min-w-0">
                    <p className="truncate text-sm font-semibold text-[#202020]">
                      {pin.author}
                    </p>
                    <p className="mt-0.5 text-xs font-medium text-[#777]">
                      {pin.saves} saves
                    </p>
                  </div>
                  <Button
                    aria-label={isSaved ? "Saved" : "Save pin"}
                    className={cn(
                      "rounded-full",
                      isSaved
                        ? "border-[#ffb3c1] bg-[#fff1f4] text-[#cf2142]"
                        : "",
                    )}
                    onClick={() => toggleSaved(pin.id)}
                    size="icon-sm"
                    type="button"
                    variant="outline"
                  >
                    <Bookmark
                      className={cn("size-4", isSaved ? "fill-current" : "")}
                    />
                  </Button>
                </div>
              </motion.article>
            );
          })}
        </div>
      </section>
    </main>
  );
}
