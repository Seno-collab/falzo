"use client";

import { useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  CalendarClock,
  Compass,
  ImageIcon,
  Loader2,
  MapPinned,
  Route,
  ShieldAlert,
  Sparkles,
  Star,
  Users,
  Wallet,
} from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useMemo, type ReactNode } from "react";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import MapClient, { type MapPoint } from "@/components/map";
import { Button } from "@/components/ui/button";
import { getApiErrorMessage } from "@/features/auth/api";
import { getPlaceBySlugApi } from "@/features/locations/api";
import type { PlaceDetail } from "@/features/locations/types";
import { useI18n } from "@/i18n/locale-provider";
import { ROUTES } from "@/lib/routes";

const copyByLocale = {
  vi: {
    documentTitleFallback: "Chi tiết địa điểm - Falzo.life",
    brand: "Falzo.life",
    subtitle: "Chi tiết địa điểm",
    mobileTitle: "Place detail",
    backHome: "Trang chủ",
    explorePlaces: "Khám phá địa điểm",
    loadingTitle: "Đang tải địa điểm",
    loadingDescription: "Falzo đang lấy thông tin thực tế của địa điểm này.",
    errorTitle: "Không thể tải địa điểm",
    retry: "Thử lại",
    provinceLabel: "Tỉnh/thành",
    bestTime: "Thời điểm nên đi",
    estimatedCost: "Chi phí ước tính",
    travelStyles: "Style chuyến đi",
    suitableFor: "Phù hợp cho",
    warningNote: "Lưu ý thực tế",
    realityRating: "Độ thực tế",
    photoRating: "Ảnh đẹp",
    hiddenGem: "Hidden gem",
    mapTitle: "Bản đồ địa điểm",
    relatedTitle: "Lịch trình có địa điểm này",
    relatedDescription:
      "Các lịch trình gắn với địa điểm sẽ được mở rộng ở sprint itinerary tiếp theo.",
    viewSamples: "Xem lịch trình mẫu",
    noData: "Đang cập nhật",
    free: "Miễn phí",
    imageAltFallback: "Ảnh địa điểm du lịch",
  },
  en: {
    documentTitleFallback: "Place detail - Falzo.life",
    brand: "Falzo.life",
    subtitle: "Place detail",
    mobileTitle: "Place detail",
    backHome: "Home",
    explorePlaces: "Explore places",
    loadingTitle: "Loading place",
    loadingDescription: "Falzo is loading practical details for this place.",
    errorTitle: "Could not load this place",
    retry: "Try again",
    provinceLabel: "Province",
    bestTime: "Best time to visit",
    estimatedCost: "Estimated cost",
    travelStyles: "Travel styles",
    suitableFor: "Suitable for",
    warningNote: "Practical warning",
    realityRating: "Reality rating",
    photoRating: "Photo rating",
    hiddenGem: "Hidden gem",
    mapTitle: "Place map",
    relatedTitle: "Itineraries with this place",
    relatedDescription:
      "Itineraries linked to this place will expand in the next itinerary sprint.",
    viewSamples: "View sample itineraries",
    noData: "Updating",
    free: "Free",
    imageAltFallback: "Travel place photo",
  },
} as const;

function readSlugParam(value: string | string[] | undefined) {
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }

  return value ?? "";
}

function formatCost(place: PlaceDetail, freeLabel: string) {
  if (place.estimatedCostMin === 0 && place.estimatedCostMax === 0) {
    return freeLabel;
  }

  const formatter = new Intl.NumberFormat("vi-VN");
  return `${formatter.format(place.estimatedCostMin)} - ${formatter.format(
    place.estimatedCostMax,
  )} VND`;
}

function TagList({
  fallback,
  items,
}: Readonly<{
  fallback: string;
  items: string[];
}>) {
  if (items.length === 0) {
    return <p className="text-sm text-[#73827f]">{fallback}</p>;
  }

  return (
    <div className="flex flex-wrap gap-2">
      {items.map((item) => (
        <span
          className="rounded-full bg-[#eef7f3] px-3 py-1 text-xs font-semibold text-[#315d55]"
          key={item}
        >
          {item}
        </span>
      ))}
    </div>
  );
}

function InfoBlock({
  icon,
  label,
  value,
}: Readonly<{
  icon: ReactNode;
  label: string;
  value: string;
}>) {
  return (
    <div className="rounded-lg border border-[#dbe6e2] bg-white p-4">
      <div className="flex items-start gap-3">
        <span className="mt-0.5 text-[#397168]">{icon}</span>
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#6d7c78]">
            {label}
          </p>
          <p className="mt-1 text-sm font-semibold leading-6 text-[#172e29]">
            {value}
          </p>
        </div>
      </div>
    </div>
  );
}

export function PlaceDetailScreen() {
  const params = useParams();
  const { locale } = useI18n();
  const copy = copyByLocale[locale];
  const slug = readSlugParam(params.slug);

  const placeQuery = useQuery({
    enabled: slug.length > 0,
    queryKey: ["places", slug],
    queryFn: ({ signal }) => getPlaceBySlugApi(slug, { signal }),
    retry: 1,
  });

  const place = placeQuery.data;
  const mapPoints = useMemo<MapPoint[]>(() => {
    if (!place) {
      return [];
    }

    return [
      {
        id: place.id,
        name: place.name,
        address: place.address || place.province,
        latitude: place.latitude,
        longitude: place.longitude,
        imageUrl: place.imageUrl,
      },
    ];
  }, [place]);

  useEffect(() => {
    document.title = place
      ? `${place.name} - Falzo.life`
      : copy.documentTitleFallback;
  }, [copy.documentTitleFallback, place]);

  return (
    <PageShell
      contentClassName="space-y-6"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "home",
              icon: <ArrowLeft className="size-4" />,
              label: copy.backHome,
              to: ROUTES.home,
              variant: "outline",
            },
            {
              id: "locations",
              icon: <MapPinned className="size-4" />,
              label: copy.explorePlaces,
              to: ROUTES.locations,
              variant: "default",
            },
          ]}
          brand={copy.brand}
          brandIcon={<Compass className="size-3.5" />}
          mobileMenuTitle={copy.mobileTitle}
          subtitle={copy.subtitle}
        />
      }
    >
      {placeQuery.isLoading ? (
        <section className="rounded-lg border border-[#dbe6e2] bg-white p-6 text-center shadow-[0_18px_48px_-38px_rgb(22_44_40/0.72)]">
          <Loader2 className="mx-auto size-6 animate-spin text-[#397168]" />
          <h1 className="mt-4 text-xl font-semibold text-[#172e29]">
            {copy.loadingTitle}
          </h1>
          <p className="mt-2 text-sm text-[#667b76]">
            {copy.loadingDescription}
          </p>
        </section>
      ) : placeQuery.isError || !place ? (
        <section className="rounded-lg border border-[#ead7d3] bg-white p-6 text-center shadow-[0_18px_48px_-38px_rgb(84_34_24/0.48)]">
          <ShieldAlert className="mx-auto size-7 text-[#a24b35]" />
          <h1 className="mt-4 text-xl font-semibold text-[#3b211a]">
            {copy.errorTitle}
          </h1>
          <p className="mt-2 text-sm leading-6 text-[#7b5a52]">
            {placeQuery.error
              ? getApiErrorMessage(placeQuery.error)
              : copy.errorTitle}
          </p>
          <Button
            className="mt-5"
            onClick={() => {
              placeQuery.refetch().catch(() => undefined);
            }}
            type="button"
            variant="default"
          >
            {copy.retry}
          </Button>
        </section>
      ) : (
        <>
          <section className="grid gap-5 lg:grid-cols-[1.15fr_0.85fr]">
            <div className="space-y-5">
              <div className="space-y-3">
                <div className="flex flex-wrap items-center gap-2">
                  {place.isHiddenGem ? (
                    <span className="inline-flex items-center gap-1.5 rounded-full bg-[#e8f6ef] px-3 py-1 text-xs font-semibold text-[#285f4d]">
                      <Sparkles className="size-3.5" />
                      {copy.hiddenGem}
                    </span>
                  ) : null}
                  <span className="rounded-full bg-[#f2f7f5] px-3 py-1 text-xs font-semibold text-[#44635d]">
                    {place.province || copy.noData}
                  </span>
                </div>
                <h1 className="text-4xl font-semibold leading-tight text-[#122b26] sm:text-5xl">
                  {place.name}
                </h1>
                <p className="text-base leading-7 text-[#5b716d]">
                  {place.description || place.address || copy.noData}
                </p>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                <InfoBlock
                  icon={<MapPinned className="size-5" />}
                  label={copy.provinceLabel}
                  value={[place.district, place.province]
                    .filter(Boolean)
                    .join(", ") || copy.noData}
                />
                <InfoBlock
                  icon={<CalendarClock className="size-5" />}
                  label={copy.bestTime}
                  value={place.bestTimeToVisit || copy.noData}
                />
                <InfoBlock
                  icon={<Wallet className="size-5" />}
                  label={copy.estimatedCost}
                  value={formatCost(place, copy.free)}
                />
                <InfoBlock
                  icon={<Star className="size-5" />}
                  label={`${copy.realityRating} / ${copy.photoRating}`}
                  value={`${place.ratingReality ?? copy.noData} / ${
                    place.ratingPhoto ?? copy.noData
                  }`}
                />
              </div>
            </div>

            <div className="overflow-hidden rounded-lg border border-[#dbe6e2] bg-[#eef3ef]">
              {place.imageUrl ? (
                <img
                  alt={place.name}
                  className="aspect-[4/3] h-full min-h-72 w-full object-cover"
                  decoding="async"
                  fetchPriority="high"
                  src={place.imageUrl}
                />
              ) : (
                <div className="flex aspect-[4/3] min-h-72 flex-col items-center justify-center gap-3 text-[#647671]">
                  <ImageIcon className="size-8" />
                  <p className="text-sm font-semibold">
                    {copy.imageAltFallback}
                  </p>
                </div>
              )}
            </div>
          </section>

          <section className="grid gap-4 lg:grid-cols-[0.8fr_1.2fr]">
            <div className="space-y-4">
              <div className="rounded-lg border border-[#dbe6e2] bg-white p-4">
                <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-[#17332e]">
                  <Route className="size-4 text-[#397168]" />
                  {copy.travelStyles}
                </div>
                <TagList fallback={copy.noData} items={place.travelStyles} />
              </div>

              <div className="rounded-lg border border-[#dbe6e2] bg-white p-4">
                <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-[#17332e]">
                  <Users className="size-4 text-[#397168]" />
                  {copy.suitableFor}
                </div>
                <TagList fallback={copy.noData} items={place.suitableFor} />
              </div>

              {place.warningNote ? (
                <div className="rounded-lg border border-[#eadbd5] bg-[#fff8f5] p-4 text-[#714436]">
                  <div className="mb-2 flex items-center gap-2 text-sm font-semibold">
                    <ShieldAlert className="size-4" />
                    {copy.warningNote}
                  </div>
                  <p className="text-sm leading-6">{place.warningNote}</p>
                </div>
              ) : null}
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between gap-3">
                <h2 className="text-lg font-semibold text-[#17332e]">
                  {copy.mapTitle}
                </h2>
                <p className="text-xs font-semibold text-[#667b76]">
                  {place.latitude.toFixed(5)}, {place.longitude.toFixed(5)}
                </p>
              </div>
              <MapClient
                height="large"
                points={mapPoints}
                selectedPointId={place.id}
                zoom={13}
              />
            </div>
          </section>

          <section className="rounded-lg border border-[#dbe6e2] bg-white p-5 shadow-[0_16px_42px_-34px_rgb(22_44_40/0.72)] sm:p-6">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div className="max-w-2xl">
                <h2 className="text-xl font-semibold text-[#17332e]">
                  {copy.relatedTitle}
                </h2>
                <p className="mt-2 text-sm leading-6 text-[#667b76]">
                  {copy.relatedDescription}
                </p>
              </div>
              <Button asChild variant="default">
                <Link href={ROUTES.itineraries}>
                  <Route className="size-4" />
                  {copy.viewSamples}
                </Link>
              </Button>
            </div>
          </section>
        </>
      )}
    </PageShell>
  );
}
