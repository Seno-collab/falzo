"use client";

import {
  CalendarDays,
  Compass,
  Copy,
  MapPinned,
  Plus,
  Route,
  ShieldAlert,
  Wallet,
} from "lucide-react";
import Link from "next/link";
import { useEffect } from "react";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import { ScenicImage } from "@/components/scenic-image";
import { Button } from "@/components/ui/button";
import { useI18n } from "@/i18n/locale-provider";
import { ROUTES } from "@/lib/routes";

const copyByLocale = {
  vi: {
    documentTitle: "Falzo.life - Lên lịch trình du lịch Việt Nam",
    brand: "Falzo.life",
    navSubtitle: "Lịch trình du lịch Việt Nam",
    mobileTitle: "Falzo itinerary",
    headline: "Lên lịch trình du lịch Việt Nam trong vài phút",
    subHeadline:
      "Khám phá địa điểm thật, ảnh thật, chi phí thật và lịch trình có thể copy ngay cho chuyến đi 1-3 ngày.",
    sampleCta: "Xem lịch trình mẫu",
    createCta: "Tạo lịch trình",
    exploreCta: "Khám phá địa điểm",
    heroBadge: "Travel itinerary planner",
    proofItems: ["1-3 ngày", "Chi phí thật", "Route thật"],
    sectionLabel: "Lịch trình mẫu",
    sectionTitle: "Bắt đầu từ những chuyến đi đã được sắp sẵn",
    sectionDescription:
      "Mỗi lịch trình gom địa điểm, thời gian đẹp, chi phí ước tính và lưu ý thực tế để bạn copy nhanh.",
    detailLabel: "Có trong lịch trình",
    samples: [
      {
        title: "Phú Yên 2 ngày săn biển xanh",
        imageId: "ly-son-coast",
        province: "Phú Yên",
        duration: "2 ngày",
        cost: "900k - 1.6tr",
        stops: "Hòn Yến, Gành Đá Đĩa, Mũi Điện",
        note: "Dễ đi xe máy, nên tránh ngày biển động.",
      },
      {
        title: "Mù Cang Chải mùa lúa 3 ngày",
        imageId: "mu-cang-chai-dawn",
        province: "Yên Bái",
        duration: "3 ngày",
        cost: "1.8tr - 3tr",
        stops: "Đèo Khau Phạ, La Pán Tẩn, Tú Lệ",
        note: "Đẹp nhất sáng sớm, cần đặt homestay trước.",
      },
      {
        title: "Đà Nẵng - Hội An chill 1 ngày",
        imageId: "kyoto-lantern-night",
        province: "Đà Nẵng",
        duration: "1 ngày",
        cost: "650k - 1.2tr",
        stops: "Sơn Trà, Mỹ Khê, phố cổ Hội An",
        note: "Phù hợp nhóm nhỏ và cặp đôi.",
      },
    ],
  },
  en: {
    documentTitle: "Falzo.life - Plan Vietnam travel itineraries",
    brand: "Falzo.life",
    navSubtitle: "Vietnam travel itineraries",
    mobileTitle: "Falzo itinerary",
    headline: "Plan a Vietnam trip itinerary in minutes",
    subHeadline:
      "Explore real places, real photos, realistic costs, and copy-ready 1-3 day plans.",
    sampleCta: "View sample itineraries",
    createCta: "Create itinerary",
    exploreCta: "Explore places",
    heroBadge: "Travel itinerary planner",
    proofItems: ["1-3 days", "Real costs", "Real routes"],
    sectionLabel: "Sample itineraries",
    sectionTitle: "Start from trips that are already structured",
    sectionDescription:
      "Each itinerary groups places, best timing, estimated budget, and practical notes so you can copy faster.",
    detailLabel: "Included stops",
    samples: [
      {
        title: "Phu Yen blue-coast 2-day plan",
        imageId: "ly-son-coast",
        province: "Phu Yen",
        duration: "2 days",
        cost: "900k - 1.6m VND",
        stops: "Hon Yen, Ganh Da Dia, Mui Dien",
        note: "Easy by motorbike, check sea conditions first.",
      },
      {
        title: "Mu Cang Chai rice-season 3 days",
        imageId: "mu-cang-chai-dawn",
        province: "Yen Bai",
        duration: "3 days",
        cost: "1.8m - 3m VND",
        stops: "Khau Pha Pass, La Pan Tan, Tu Le",
        note: "Best at sunrise, book homestays early.",
      },
      {
        title: "Da Nang - Hoi An 1-day chill route",
        imageId: "kyoto-lantern-night",
        province: "Da Nang",
        duration: "1 day",
        cost: "650k - 1.2m VND",
        stops: "Son Tra, My Khe, Hoi An old town",
        note: "Good for small groups and couples.",
      },
    ],
  },
} as const;

export function HomeScreen() {
  const { locale } = useI18n();
  const copy = copyByLocale[locale];

  useEffect(() => {
    document.title = copy.documentTitle;
  }, [copy.documentTitle]);

  return (
    <PageShell
      contentClassName="space-y-8"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "itineraries",
              icon: <Route className="size-4" />,
              label: copy.sampleCta,
              to: ROUTES.itineraries,
              variant: "default",
            },
            {
              id: "locations",
              icon: <MapPinned className="size-4" />,
              label: copy.exploreCta,
              to: ROUTES.locations,
              variant: "outline",
            },
          ]}
          brand={copy.brand}
          brandIcon={<Compass className="size-3.5" />}
          mobileMenuTitle={copy.mobileTitle}
          subtitle={copy.navSubtitle}
        />
      }
    >
      <section className="relative min-h-[68svh] overflow-hidden rounded-lg bg-[#15251f] text-white shadow-[0_28px_72px_-48px_rgb(20_37_31/0.8)]">
        <ScenicImage
          alt=""
          className="absolute inset-0 h-full w-full object-cover opacity-72"
          fetchPriority="high"
          id="mu-cang-chai-dawn"
          loading="eager"
          sizes="100vw"
        />
        <div className="absolute inset-0 bg-[linear-gradient(90deg,rgba(7,20,16,0.86),rgba(7,20,16,0.48)_48%,rgba(7,20,16,0.2))]" />
        <div className="relative flex min-h-[68svh] max-w-3xl flex-col justify-end gap-6 px-5 py-8 sm:px-8 lg:px-10">
          <div className="space-y-4">
            <p className="inline-flex w-fit items-center gap-2 rounded-full border border-white/20 bg-white/12 px-3 py-1 text-xs font-semibold uppercase tracking-[0.14em] text-white/92 backdrop-blur">
              <Route className="size-3.5" />
              {copy.heroBadge}
            </p>
            <h1 className="max-w-3xl text-4xl font-semibold leading-[1.04] tracking-normal sm:text-5xl lg:text-6xl">
              {copy.headline}
            </h1>
            <p className="max-w-2xl text-base leading-7 text-white/84 sm:text-lg">
              {copy.subHeadline}
            </p>
          </div>

          <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
            <Button asChild className="sm:min-w-48" size="lg" variant="default">
              <Link href={ROUTES.itineraries}>
                <Copy className="size-4" />
                {copy.sampleCta}
              </Link>
            </Button>
            <Button asChild className="sm:min-w-44" size="lg" variant="outline">
              <Link href={`${ROUTES.itineraries}?intent=create`}>
                <Plus className="size-4" />
                {copy.createCta}
              </Link>
            </Button>
            <Button asChild className="sm:min-w-44" size="lg" variant="outline">
              <Link href={ROUTES.locations}>
                <MapPinned className="size-4" />
                {copy.exploreCta}
              </Link>
            </Button>
          </div>

          <div className="grid gap-2 text-sm font-semibold text-white/86 sm:grid-cols-3">
            {copy.proofItems.map((item) => (
              <div
                className="rounded-lg border border-white/14 bg-white/10 px-3 py-2 backdrop-blur"
                key={item}
              >
                {item}
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="space-y-4 pb-4">
        <div className="max-w-3xl space-y-2">
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#397168]">
            {copy.sectionLabel}
          </p>
          <h2 className="text-2xl font-semibold leading-tight text-[#162c28] sm:text-3xl">
            {copy.sectionTitle}
          </h2>
          <p className="text-sm leading-6 text-[#5f716f] sm:text-base">
            {copy.sectionDescription}
          </p>
        </div>

        <div className="grid gap-4 lg:grid-cols-3">
          {copy.samples.map((sample) => (
            <article
              className="overflow-hidden rounded-lg border border-[#d9e5df] bg-white shadow-[0_16px_42px_-34px_rgb(22_44_40/0.72)]"
              key={sample.title}
            >
              <div className="relative aspect-[16/10] bg-[#e7eee9]">
                <ScenicImage
                  alt={sample.title}
                  className="h-full w-full object-cover"
                  id={sample.imageId}
                  sizes="(max-width: 1024px) 100vw, 33vw"
                />
                <span className="absolute left-3 top-3 rounded-full bg-white/92 px-3 py-1 text-xs font-semibold text-[#16332e] shadow-sm">
                  {sample.province}
                </span>
              </div>
              <div className="space-y-4 p-4">
                <div className="space-y-2">
                  <h3 className="text-base font-semibold leading-6 text-[#10231f]">
                    {sample.title}
                  </h3>
                  <div className="grid grid-cols-2 gap-2 text-xs font-semibold text-[#46645f]">
                    <span className="inline-flex items-center gap-1.5 rounded-lg bg-[#f2f7f5] px-2.5 py-2">
                      <CalendarDays className="size-3.5" />
                      {sample.duration}
                    </span>
                    <span className="inline-flex items-center gap-1.5 rounded-lg bg-[#fff6e8] px-2.5 py-2 text-[#7a5a18]">
                      <Wallet className="size-3.5" />
                      {sample.cost}
                    </span>
                  </div>
                </div>

                <div className="space-y-2 text-sm leading-6 text-[#506865]">
                  <p>
                    <span className="font-semibold text-[#17332e]">
                      {copy.detailLabel}:{" "}
                    </span>
                    {sample.stops}
                  </p>
                  <p className="inline-flex gap-2 rounded-lg bg-[#f8f1ee] px-3 py-2 text-[#714436]">
                    <ShieldAlert className="mt-1 size-4 shrink-0" />
                    {sample.note}
                  </p>
                </div>
              </div>
            </article>
          ))}
        </div>
      </section>
    </PageShell>
  );
}
