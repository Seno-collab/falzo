"use client";

import { CalendarDays, Compass, MapPinned, Route, Wallet } from "lucide-react";
import Link from "next/link";
import { useEffect } from "react";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import { Button } from "@/components/ui/button";
import { useI18n } from "@/i18n/locale-provider";
import { ROUTES } from "@/lib/routes";

const copyByLocale = {
  vi: {
    documentTitle: "Lịch trình mẫu - Falzo.life",
    brand: "Falzo.life",
    subtitle: "Lịch trình mẫu",
    mobileTitle: "Itineraries",
    title: "Lịch trình mẫu",
    description:
      "Danh sách lịch trình đầy đủ sẽ được mở rộng ở sprint tiếp theo. Sprint 1 giữ route này để người dùng có đích rõ ràng từ trang chủ.",
    exploreCta: "Khám phá địa điểm",
    items: [
      {
        title: "Phú Yên 2 ngày",
        province: "Phú Yên",
        duration: "2 ngày",
        cost: "900k - 1.6tr",
        stops: "Hòn Yến, Gành Đá Đĩa, Mũi Điện",
      },
      {
        title: "Mù Cang Chải 3 ngày",
        province: "Yên Bái",
        duration: "3 ngày",
        cost: "1.8tr - 3tr",
        stops: "Khau Phạ, La Pán Tẩn, Tú Lệ",
      },
      {
        title: "Đà Nẵng - Hội An 1 ngày",
        province: "Đà Nẵng",
        duration: "1 ngày",
        cost: "650k - 1.2tr",
        stops: "Sơn Trà, Mỹ Khê, Hội An",
      },
    ],
  },
  en: {
    documentTitle: "Sample itineraries - Falzo.life",
    brand: "Falzo.life",
    subtitle: "Sample itineraries",
    mobileTitle: "Itineraries",
    title: "Sample itineraries",
    description:
      "The complete itinerary list will expand in the next sprint. Sprint 1 keeps this route available from the homepage.",
    exploreCta: "Explore places",
    items: [
      {
        title: "Phu Yen 2 days",
        province: "Phu Yen",
        duration: "2 days",
        cost: "900k - 1.6m VND",
        stops: "Hon Yen, Ganh Da Dia, Mui Dien",
      },
      {
        title: "Mu Cang Chai 3 days",
        province: "Yen Bai",
        duration: "3 days",
        cost: "1.8m - 3m VND",
        stops: "Khau Pha, La Pan Tan, Tu Le",
      },
      {
        title: "Da Nang - Hoi An 1 day",
        province: "Da Nang",
        duration: "1 day",
        cost: "650k - 1.2m VND",
        stops: "Son Tra, My Khe, Hoi An",
      },
    ],
  },
} as const;

export function ItinerariesListScreen() {
  const { locale } = useI18n();
  const copy = copyByLocale[locale];

  useEffect(() => {
    document.title = copy.documentTitle;
  }, [copy.documentTitle]);

  return (
    <PageShell
      contentClassName="space-y-6"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "locations",
              icon: <MapPinned className="size-4" />,
              label: copy.exploreCta,
              to: ROUTES.locations,
              variant: "outline",
            },
          ]}
          brand={copy.brand}
          brandIcon={<Route className="size-3.5" />}
          mobileMenuTitle={copy.mobileTitle}
          subtitle={copy.subtitle}
        />
      }
    >
      <section className="rounded-lg border border-[#d8e5e0] bg-white p-5 shadow-[0_18px_48px_-38px_rgb(22_44_40/0.72)] sm:p-7">
        <div className="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div className="max-w-2xl space-y-2">
            <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#397168]">
              Falzo.life
            </p>
            <h1 className="text-3xl font-semibold leading-tight text-[#152d28]">
              {copy.title}
            </h1>
            <p className="text-sm leading-6 text-[#5f716f]">
              {copy.description}
            </p>
          </div>
          <Button asChild variant="default">
            <Link href={ROUTES.locations}>
              <Compass className="size-4" />
              {copy.exploreCta}
            </Link>
          </Button>
        </div>
      </section>

      <section className="grid gap-4 lg:grid-cols-3">
        {copy.items.map((item) => (
          <article
            className="rounded-lg border border-[#d8e5e0] bg-white p-4 shadow-[0_16px_42px_-34px_rgb(22_44_40/0.72)]"
            key={item.title}
          >
            <div className="space-y-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#6d7c78]">
                  {item.province}
                </p>
                <h2 className="mt-1 text-lg font-semibold text-[#142c27]">
                  {item.title}
                </h2>
              </div>
              <div className="grid grid-cols-2 gap-2 text-xs font-semibold">
                <span className="inline-flex items-center gap-1.5 rounded-lg bg-[#f2f7f5] px-2.5 py-2 text-[#46645f]">
                  <CalendarDays className="size-3.5" />
                  {item.duration}
                </span>
                <span className="inline-flex items-center gap-1.5 rounded-lg bg-[#fff6e8] px-2.5 py-2 text-[#7a5a18]">
                  <Wallet className="size-3.5" />
                  {item.cost}
                </span>
              </div>
              <p className="text-sm leading-6 text-[#506865]">{item.stops}</p>
            </div>
          </article>
        ))}
      </section>
    </PageShell>
  );
}
