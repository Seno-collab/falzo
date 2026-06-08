"use client";

import { useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  CalendarDays,
  Copy,
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
    <div className="rounded-lg border border-[#dbe6e2] bg-white p-4">
      <div className="flex items-start gap-3">
        <span className="mt-0.5 text-[#397168]">{icon}</span>
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
          <section className="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
            <div className="space-y-5">
              <div className="space-y-3">
                <p className="inline-flex w-fit items-center rounded-full bg-[#e8f6ef] px-3 py-1 text-xs font-semibold uppercase tracking-[0.14em] text-[#285f4d]">
                  {itinerary.province}
                </p>
                <h1 className="text-4xl font-semibold leading-tight text-[#122b26] sm:text-5xl">
                  {itinerary.title}
                </h1>
                <p className="text-base leading-7 text-[#5b716d]">
                  {itinerary.description ||
                    (locale === "vi"
                      ? "Lịch trình đang cập nhật mô tả."
                      : "Description is being updated.")}
                </p>
              </div>

              <div className="flex flex-wrap gap-2">
                <Button onClick={handleCopy} type="button" variant="default">
                  <Copy className="size-4" />
                  {locale === "vi" ? "Copy lịch trình" : "Copy itinerary"}
                </Button>
                <Button onClick={handleShare} type="button" variant="outline">
                  <Share2 className="size-4" />
                  {locale === "vi" ? "Chia sẻ" : "Share"}
                </Button>
              </div>
            </div>

            <div className="grid gap-3 sm:grid-cols-2">
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
              <div className="flex items-center justify-between gap-3">
                <h2 className="text-lg font-semibold text-[#17332e]">
                  {locale === "vi" ? "Bản đồ route" : "Route map"}
                </h2>
                <p className="text-xs font-semibold text-[#667b76]">
                  {stops.length} stops
                </p>
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
