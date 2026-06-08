"use client";

import { useQuery } from "@tanstack/react-query";
import {
  CalendarDays,
  Compass,
  Heart,
  Loader2,
  MapPinned,
  MessageCircle,
  Play,
  Route,
  Share2,
  TriangleAlert,
} from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import { Button } from "@/components/ui/button";
import { getApiErrorMessage } from "@/features/auth/api";
import { getItinerariesApi } from "@/features/itineraries/api/itineraries-api";
import { ItineraryCard } from "@/features/itineraries/components/itinerary-card";
import {
  ItineraryFilter,
  type ItineraryFilterValues,
} from "@/features/itineraries/components/itinerary-filter";
import type { ItineraryListParams } from "@/features/itineraries/types";
import { useI18n } from "@/i18n/locale-provider";
import { ROUTES } from "@/lib/routes";

const defaultLimit = 12;

const emptyFilters: ItineraryFilterValues = {
  province: "",
  durationDays: "",
  budgetMax: "",
  travelStyle: "",
};

function readFilters(searchParams: URLSearchParams): ItineraryFilterValues {
  return {
    province: searchParams.get("province") ?? "",
    durationDays: searchParams.get("durationDays") ?? "",
    budgetMax: searchParams.get("budgetMax") ?? "",
    travelStyle: searchParams.get("travelStyle") ?? "",
  };
}

function toNumber(value: string) {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined;
}

function toApiParams(values: ItineraryFilterValues): ItineraryListParams {
  return {
    province: values.province.trim() || undefined,
    durationDays: toNumber(values.durationDays),
    budgetMax: toNumber(values.budgetMax),
    travelStyle: values.travelStyle.trim() || undefined,
    page: 1,
    limit: defaultLimit,
  };
}

function buildQueryString(values: ItineraryFilterValues) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(values)) {
    const normalized = value.trim();
    if (normalized) {
      params.set(key, normalized);
    }
  }
  return params.toString();
}

export function ItineraryListScreen() {
  const { locale } = useI18n();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [filters, setFilters] = useState<ItineraryFilterValues>(() =>
    readFilters(searchParams),
  );
  const apiParams = useMemo(() => toApiParams(filters), [filters]);

  useEffect(() => {
    document.title =
      locale === "vi"
        ? "Lịch trình du lịch - Falzo.life"
        : "Travel itineraries - Falzo.life";
  }, [locale]);

  useEffect(() => {
    setFilters(readFilters(searchParams));
  }, [searchParams]);

  const itinerariesQuery = useQuery({
    queryKey: ["itineraries", apiParams],
    queryFn: ({ signal }) => getItinerariesApi(apiParams, { signal }),
  });

  const submitFilters = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const queryString = buildQueryString(filters);
    router.replace(queryString ? `${pathname}?${queryString}` : pathname);
  };

  const resetFilters = () => {
    setFilters(emptyFilters);
    router.replace(pathname);
  };

  const items = itinerariesQuery.data?.items ?? [];

  return (
    <PageShell
      contentClassName="space-y-6"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "locations",
              icon: <MapPinned className="size-4" />,
              label: locale === "vi" ? "Khám phá địa điểm" : "Explore places",
              to: ROUTES.locations,
              variant: "outline",
            },
          ]}
          brand="Falzo.life"
          brandIcon={<Route className="size-3.5" />}
          mobileMenuTitle="Itineraries"
          subtitle={locale === "vi" ? "Lịch trình mẫu" : "Sample itineraries"}
        />
      }
    >
      <section className="overflow-hidden rounded-lg border border-[#d8e5e0] bg-white shadow-[0_18px_48px_-38px_rgb(22_44_40/0.72)]">
        <div className="grid gap-0 lg:grid-cols-[minmax(0,1.1fr)_minmax(340px,0.9fr)]">
          <div className="flex flex-col justify-between gap-8 p-5 sm:p-7">
            <div className="max-w-3xl space-y-3">
              <p className="inline-flex w-fit items-center gap-2 rounded-full bg-[#e8f6ef] px-3 py-1 text-xs font-semibold uppercase tracking-[0.14em] text-[#397168]">
                <Play className="size-3.5 fill-current" />
                Falzo itinerary
              </p>
              <h1 className="text-3xl font-semibold leading-tight text-[#152d28] sm:text-4xl">
                {locale === "vi"
                  ? "Khám phá lịch trình du lịch Việt Nam"
                  : "Explore Vietnam travel itineraries"}
              </h1>
              <p className="max-w-2xl text-sm leading-6 text-[#5f716f] sm:text-base">
                {locale === "vi"
                  ? "Lướt nhanh các route đẹp, xem chi phí và copy lịch trình phù hợp cho chuyến đi tiếp theo."
                  : "Browse visual routes, compare cost, and copy a plan for your next trip."}
              </p>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
              <Button asChild variant="default">
                <Link href={ROUTES.locations}>
                  <Compass className="size-4" />
                  {locale === "vi" ? "Khám phá địa điểm" : "Explore places"}
                </Link>
              </Button>
              <Button asChild variant="outline">
                <Link href={`${ROUTES.itineraries}?durationDays=2`}>
                  <CalendarDays className="size-4" />
                  {locale === "vi" ? "Route 2 ngày" : "2-day routes"}
                </Link>
              </Button>
            </div>
          </div>

          <div className="border-t border-[#e6efeb] bg-[#0f1f1b] p-4 text-white lg:border-l lg:border-t-0">
            <div className="mx-auto max-w-sm overflow-hidden rounded-lg border border-white/14 bg-white/8 shadow-[0_22px_48px_-30px_rgb(0_0_0/0.8)]">
              <div className="relative aspect-[4/5] bg-[linear-gradient(145deg,#294e44,#f0aa52)]">
                <div className="absolute inset-0 bg-[radial-gradient(circle_at_22%_18%,rgba(255,255,255,0.32),transparent_28%),linear-gradient(180deg,transparent_35%,rgba(0,0,0,0.68))]" />
                <div className="absolute right-3 top-3 flex flex-col gap-2">
                  {[Heart, MessageCircle, Share2].map((Icon) => (
                    <span
                      className="flex size-10 items-center justify-center rounded-full bg-black/34 backdrop-blur"
                      key={Icon.displayName ?? Icon.name}
                    >
                      <Icon className="size-4" />
                    </span>
                  ))}
                </div>
                <div className="absolute bottom-4 left-4 right-14 space-y-3">
                  <p className="text-xs font-semibold uppercase tracking-[0.14em] text-white/76">
                    {locale === "vi" ? "Đang nổi bật" : "Trending route"}
                  </p>
                  <h2 className="text-2xl font-semibold leading-tight">
                    {locale === "vi"
                      ? "Biển xanh, phố cổ, cafe sáng"
                      : "Blue coast, old town, morning coffee"}
                  </h2>
                  <div className="flex flex-wrap gap-2 text-xs font-semibold">
                    <span className="rounded-full bg-white/16 px-2.5 py-1 backdrop-blur">
                      2 ngày
                    </span>
                    <span className="rounded-full bg-white/16 px-2.5 py-1 backdrop-blur">
                      6 điểm dừng
                    </span>
                    <span className="rounded-full bg-white/16 px-2.5 py-1 backdrop-blur">
                      #chill
                    </span>
                  </div>
                </div>
              </div>
              <div className="flex items-center justify-between gap-3 px-4 py-3 text-sm">
                <span className="font-semibold text-white/88">
                  {locale === "vi" ? "Copy lịch trình mẫu" : "Copy sample plan"}
                </span>
                <span className="rounded-full bg-white px-3 py-1 text-xs font-semibold text-[#17332e]">
                  Falzo
                </span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <form onSubmit={submitFilters}>
        <ItineraryFilter
          onChange={setFilters}
          onReset={resetFilters}
          values={filters}
        />
      </form>

      {itinerariesQuery.isLoading ? (
        <section className="rounded-lg border border-[#dbe6e2] bg-white p-8 text-center">
          <Loader2 className="mx-auto size-6 animate-spin text-[#397168]" />
          <p className="mt-3 text-sm font-semibold text-[#506865]">
            {locale === "vi" ? "Đang tải lịch trình..." : "Loading itineraries..."}
          </p>
        </section>
      ) : itinerariesQuery.isError ? (
        <section className="rounded-lg border border-[#ead7d3] bg-white p-8 text-center">
          <TriangleAlert className="mx-auto size-7 text-[#a24b35]" />
          <h2 className="mt-3 text-lg font-semibold text-[#3b211a]">
            {locale === "vi"
              ? "Không thể tải lịch trình"
              : "Could not load itineraries"}
          </h2>
          <p className="mt-2 text-sm text-[#7b5a52]">
            {getApiErrorMessage(itinerariesQuery.error)}
          </p>
          <Button
            className="mt-5"
            onClick={() => {
              itinerariesQuery.refetch().catch(() => undefined);
            }}
            type="button"
            variant="default"
          >
            {locale === "vi" ? "Thử lại" : "Try again"}
          </Button>
        </section>
      ) : items.length === 0 ? (
        <section className="rounded-lg border border-dashed border-[#c8d8d2] bg-white/82 p-8 text-center">
          <Route className="mx-auto size-8 text-[#78908a]" />
          <h2 className="mt-3 text-lg font-semibold text-[#17332e]">
            {locale === "vi"
              ? "Chưa có lịch trình phù hợp"
              : "No matching itineraries"}
          </h2>
          <p className="mt-2 text-sm text-[#667b76]">
            {locale === "vi"
              ? "Hãy thử bỏ bớt filter hoặc quay lại sau khi dữ liệu mẫu được seed."
              : "Try fewer filters or come back after sample data is seeded."}
          </p>
        </section>
      ) : (
        <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {items.map((item) => (
            <ItineraryCard item={item} key={item.id} />
          ))}
        </section>
      )}
    </PageShell>
  );
}
