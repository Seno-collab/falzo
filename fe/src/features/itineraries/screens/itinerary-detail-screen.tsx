"use client";

import { useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  Bookmark,
  CalendarDays,
  Copy,
  Heart,
  Loader2,
  MapPinned,
  Route,
  Share2,
  Train,
  TriangleAlert,
  Wallet,
} from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useMemo } from "react";
import type { ReactNode } from "react";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import MapClient, { type MapPoint } from "@/components/map";
import { Button } from "@/components/ui/button";
import { getApiErrorMessage } from "@/features/auth/api";
import { getItineraryBySlugApi } from "@/features/itineraries/api/itineraries-api";
import { ItineraryDaySection } from "@/features/itineraries/components/itinerary-day-section";
import type {
  ItineraryDay,
  ItineraryDetail,
  ItineraryStop,
} from "@/features/itineraries/types";
import { useI18n } from "@/i18n/locale-provider";
import { ROUTES } from "@/lib/routes";
import { notifyError, notifySuccess } from "@/lib/toast";

function readSlugParam(value: string | string[] | undefined) {
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }
  return value ?? "";
}

function formatMoney(value: number) {
  return new Intl.NumberFormat("vi-VN").format(value);
}

function formatBudget(item: ItineraryDetail) {
  if (item.budgetMin === 0 && item.budgetMax === 0) {
    return "Miễn phí";
  }

  return `${formatMoney(item.budgetMin)} - ${formatMoney(item.budgetMax)}đ`;
}

function splitStyles(value: string) {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean)
    .slice(0, 5);
}

function formatStopForCopy(stop: ItineraryStop) {
  const time = stop.startTime || "Linh hoạt";
  const lines = [`${time} - ${stop.locationName}`];
  if (stop.note) {
    lines.push(`Ghi chú: ${stop.note}`);
  }
  lines.push(`Chi phí: ${formatMoney(stop.estimatedCost)}đ`);
  return lines.join("\n");
}

function formatItineraryForCopy(itinerary: ItineraryDetail) {
  const lines = [
    itinerary.title,
    "",
    `${itinerary.province} · ${itinerary.durationDays} ngày · ${formatBudget(
      itinerary,
    )}`,
  ];

  if (itinerary.transportation) {
    lines.push(`Di chuyển: ${itinerary.transportation}`);
  }
  if (itinerary.startLocation) {
    lines.push(`Điểm bắt đầu: ${itinerary.startLocation}`);
  }
  if (itinerary.description) {
    lines.push("", itinerary.description);
  }

  for (const day of itinerary.days) {
    lines.push("", `Ngày ${day.dayNumber}:`);
    for (const stop of day.stops) {
      lines.push(formatStopForCopy(stop), "");
    }
  }

  return lines.join("\n").trim();
}

async function copyText(value: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(value);
    return;
  }

  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  document.body.append(textarea);
  textarea.focus();
  textarea.select();

  try {
    const ok = document.execCommand("copy");
    if (!ok) {
      throw new Error("copy failed");
    }
  } finally {
    textarea.remove();
  }
}

function flattenStops(days: ItineraryDay[]) {
  return days.flatMap((day) => day.stops);
}

function toMapPoints(stops: ItineraryStop[]): MapPoint[] {
  return stops.map((stop) => ({
    id: stop.id,
    name: `${stop.stopOrder}. ${stop.locationName}`,
    address: stop.note,
    latitude: stop.latitude,
    longitude: stop.longitude,
  }));
}

function Fact({
  icon,
  label,
  value,
}: Readonly<{
  icon: ReactNode;
  label: string;
  value: string;
}>) {
  return (
    <div className="rounded-lg border border-[#dbe6e2] bg-white/94 p-4 shadow-[0_12px_30px_-26px_rgb(22_44_40/0.62)]">
      <div className="flex items-start gap-3">
        <span className="mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-full bg-[#eef7f3] text-[#397168]">
          {icon}
        </span>
        <div>
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

export function ItineraryDetailScreen() {
  const params = useParams();
  const { locale } = useI18n();
  const slug = readSlugParam(params.slug);

  const itineraryQuery = useQuery({
    enabled: slug.length > 0,
    queryKey: ["itineraries", "detail", slug],
    queryFn: ({ signal }) => getItineraryBySlugApi(slug, { signal }),
  });

  const itinerary = itineraryQuery.data;
  const styles = useMemo(
    () => splitStyles(itinerary?.travelStyle ?? ""),
    [itinerary?.travelStyle],
  );
  const stops = useMemo(
    () => flattenStops(itinerary?.days ?? []),
    [itinerary?.days],
  );
  const mapPoints = useMemo(() => toMapPoints(stops), [stops]);

  useEffect(() => {
    document.title = itinerary
      ? `${itinerary.title} - Falzo.life`
      : locale === "vi"
        ? "Chi tiết lịch trình - Falzo.life"
        : "Itinerary detail - Falzo.life";
  }, [itinerary, locale]);

  const handleCopy = async () => {
    if (!itinerary) {
      return;
    }

    try {
      await copyText(formatItineraryForCopy(itinerary));
      notifySuccess(locale === "vi" ? "Đã copy lịch trình" : "Itinerary copied");
    } catch {
      notifyError(
        locale === "vi"
          ? "Không thể copy lịch trình"
          : "Could not copy itinerary",
      );
    }
  };

  const handleShare = async () => {
    if (!itinerary) {
      return;
    }

    const url = globalThis.location?.href ?? ROUTES.itineraryDetail(itinerary.slug);

    try {
      if (navigator.share) {
        await navigator.share({
          title: itinerary.title,
          text: itinerary.description || itinerary.title,
          url,
        });
        notifySuccess(
          locale === "vi" ? "Đã mở chia sẻ lịch trình" : "Share sheet opened",
        );
        return;
      }

      await copyText(url);
      notifySuccess(
        locale === "vi"
          ? "Đã copy link lịch trình"
          : "Itinerary link copied",
      );
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") {
        return;
      }
      notifyError(
        locale === "vi"
          ? "Không thể chia sẻ lịch trình"
          : "Could not share itinerary",
      );
    }
  };

  return (
    <PageShell
      contentClassName="space-y-6"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "itineraries",
              icon: <ArrowLeft className="size-4" />,
              label: locale === "vi" ? "Danh sách" : "List",
              to: ROUTES.itineraries,
              variant: "outline",
            },
            {
              id: "locations",
              icon: <MapPinned className="size-4" />,
              label: locale === "vi" ? "Địa điểm" : "Places",
              to: ROUTES.locations,
              variant: "outline",
            },
          ]}
          brand="Falzo.life"
          brandIcon={<Route className="size-3.5" />}
          mobileMenuTitle="Itinerary"
          subtitle={locale === "vi" ? "Chi tiết lịch trình" : "Itinerary detail"}
        />
      }
    >
      {itineraryQuery.isLoading ? (
        <section className="rounded-lg border border-[#dbe6e2] bg-white p-8 text-center">
          <Loader2 className="mx-auto size-6 animate-spin text-[#397168]" />
          <p className="mt-3 text-sm font-semibold text-[#506865]">
            {locale === "vi" ? "Đang tải lịch trình..." : "Loading itinerary..."}
          </p>
        </section>
      ) : itineraryQuery.isError || !itinerary ? (
        <section className="rounded-lg border border-[#ead7d3] bg-white p-8 text-center">
          <TriangleAlert className="mx-auto size-7 text-[#a24b35]" />
          <h1 className="mt-3 text-xl font-semibold text-[#3b211a]">
            {locale === "vi"
              ? "Không tìm thấy lịch trình"
              : "Itinerary unavailable"}
          </h1>
          <p className="mt-2 text-sm text-[#7b5a52]">
            {itineraryQuery.error
              ? getApiErrorMessage(itineraryQuery.error)
              : locale === "vi"
                ? "Lịch trình này chưa public hoặc không tồn tại."
                : "This itinerary is not public or does not exist."}
          </p>
          <Button asChild className="mt-5" variant="default">
            <Link href={ROUTES.itineraries}>
              {locale === "vi" ? "Về danh sách" : "Back to list"}
            </Link>
          </Button>
        </section>
      ) : (
        <>
          <section className="overflow-hidden rounded-lg border border-[#d8e5e0] bg-white shadow-[0_18px_48px_-38px_rgb(22_44_40/0.72)]">
            <div className="grid lg:grid-cols-[minmax(0,1.05fr)_minmax(340px,0.95fr)]">
              <div className="relative min-h-[420px] bg-[#10231f]">
                {itinerary.coverImageUrl ? (
                  <img
                    alt={itinerary.title}
                    className="absolute inset-0 h-full w-full object-cover"
                    decoding="async"
                    src={itinerary.coverImageUrl}
                  />
                ) : (
                  <div className="absolute inset-0 bg-[linear-gradient(145deg,#17332e,#2f796d_55%,#f0aa52)]" />
                )}
                <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(6,18,15,0.16),rgba(6,18,15,0.82))]" />
                <div className="absolute bottom-0 left-0 right-0 space-y-4 p-5 text-white sm:p-7">
                  <div className="flex flex-wrap gap-2">
                    <span className="inline-flex w-fit items-center gap-1.5 rounded-full bg-white/16 px-3 py-1 text-xs font-semibold uppercase tracking-[0.14em] backdrop-blur">
                      <MapPinned className="size-3.5" />
                      {itinerary.province}
                    </span>
                    {styles.map((style) => (
                      <span
                        className="rounded-full bg-white/16 px-3 py-1 text-xs font-semibold backdrop-blur"
                        key={style}
                      >
                        #{style}
                      </span>
                    ))}
                  </div>
                  <div className="max-w-2xl space-y-3">
                    <h1 className="text-4xl font-semibold leading-tight sm:text-5xl">
                      {itinerary.title}
                    </h1>
                    <p className="text-base leading-7 text-white/86">
                      {itinerary.description ||
                        (locale === "vi"
                          ? "Lịch trình đang cập nhật mô tả."
                          : "Description is being updated.")}
                    </p>
                  </div>
                </div>
              </div>

              <div className="flex flex-col justify-between gap-5 border-t border-[#e6efeb] p-5 sm:p-6 lg:border-l lg:border-t-0">
                <div className="space-y-4">
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex min-w-0 items-center gap-3">
                      <span className="flex size-11 shrink-0 items-center justify-center rounded-full bg-[linear-gradient(145deg,#18342f,#2f796d)] text-white shadow-sm">
                        <Route className="size-5" />
                      </span>
                      <div className="min-w-0">
                        <p className="truncate text-sm font-semibold text-[#142c27]">
                          Falzo itinerary
                        </p>
                        <p className="truncate text-xs font-medium text-[#71817d]">
                          {stops.length} stops · {itinerary.province}
                        </p>
                      </div>
                    </div>
                    <div className="flex shrink-0 gap-1.5">
                      <span className="flex size-9 items-center justify-center rounded-full bg-[#f4f8f6] text-[#526762]">
                        <Heart className="size-4" />
                      </span>
                      <span className="flex size-9 items-center justify-center rounded-full bg-[#f4f8f6] text-[#526762]">
                        <Bookmark className="size-4" />
                      </span>
                    </div>
                  </div>

                  <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-1 xl:grid-cols-2">
                    <Fact
                      icon={<CalendarDays className="size-5" />}
                      label={locale === "vi" ? "Thời lượng" : "Duration"}
                      value={`${itinerary.durationDays} ngày`}
                    />
                    <Fact
                      icon={<Wallet className="size-5" />}
                      label={locale === "vi" ? "Ngân sách" : "Budget"}
                      value={formatBudget(itinerary)}
                    />
                    <Fact
                      icon={<Train className="size-5" />}
                      label={locale === "vi" ? "Di chuyển" : "Transportation"}
                      value={itinerary.transportation || "Đang cập nhật"}
                    />
                    <Fact
                      icon={<MapPinned className="size-5" />}
                      label={locale === "vi" ? "Điểm bắt đầu" : "Start"}
                      value={itinerary.startLocation || itinerary.province}
                    />
                  </div>
                </div>

                <div className="flex flex-col gap-2 sm:flex-row lg:flex-col xl:flex-row">
                  <Button
                    className="flex-1"
                    onClick={handleCopy}
                    type="button"
                    variant="default"
                  >
                    <Copy className="size-4" />
                    {locale === "vi" ? "Copy lịch trình" : "Copy itinerary"}
                  </Button>
                  <Button
                    className="flex-1"
                    onClick={handleShare}
                    type="button"
                    variant="outline"
                  >
                    <Share2 className="size-4" />
                    {locale === "vi" ? "Chia sẻ" : "Share"}
                  </Button>
                </div>
              </div>
            </div>
          </section>

          <section className="grid gap-5 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
            <div className="space-y-5">
              {itinerary.days.length > 0 ? (
                itinerary.days.map((day) => (
                  <ItineraryDaySection day={day} key={day.dayNumber} />
                ))
              ) : (
                <div className="rounded-lg border border-dashed border-[#c8d8d2] bg-white/82 p-6 text-center text-sm text-[#667b76]">
                  {locale === "vi"
                    ? "Lịch trình chưa có điểm dừng."
                    : "This itinerary has no stops yet."}
                </div>
              )}
            </div>

            <div className="space-y-3 xl:sticky xl:top-5 xl:self-start">
              <div className="rounded-lg border border-[#d8e5e0] bg-white/94 p-4 shadow-[0_16px_42px_-36px_rgb(22_44_40/0.62)]">
                <div className="flex items-center justify-between gap-3">
                  <div>
                    <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#397168]">
                      {locale === "vi" ? "Bản đồ route" : "Route map"}
                    </p>
                    <h2 className="mt-1 text-lg font-semibold text-[#17332e]">
                      {locale === "vi" ? "Các điểm trên hành trình" : "Stops on route"}
                    </h2>
                  </div>
                  <p className="rounded-full bg-[#eef7f3] px-3 py-1 text-xs font-semibold text-[#315d55]">
                    {stops.length} stops
                  </p>
                </div>
              </div>
              {mapPoints.length > 0 ? (
                <MapClient
                  height="large"
                  points={mapPoints}
                  selectedPointId={mapPoints[0]?.id}
                  zoom={11}
                />
              ) : (
                <div className="flex min-h-96 items-center justify-center rounded-lg border border-[#dbe6e2] bg-white text-sm font-semibold text-[#667b76]">
                  {locale === "vi"
                    ? "Chưa có tọa độ điểm dừng"
                    : "No stop coordinates yet"}
                </div>
              )}
            </div>
          </section>
        </>
      )}
    </PageShell>
  );
}
