"use client";

import {
  CalendarDays,
  Compass,
  MapPinned,
  ShieldAlert,
  Trophy,
  Wallet,
} from "lucide-react";
import { useEffect } from "react";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import { ScenicImage } from "@/components/scenic-image";
import { useI18n } from "@/i18n/locale-provider";
import { ROUTES } from "@/lib/routes";
import { TravelGameHero } from "@/features/travel-game/components/travel-game-hero";
import { travelGameCopyByLocale } from "@/features/travel-game/data/travel-game-copy";

export function HomeScreen() {
  const { locale } = useI18n();
  const copy = travelGameCopyByLocale[locale];

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
              id: "travel-game",
              icon: <Trophy className="size-4" />,
              label: copy.sampleCta,
              to: ROUTES.explore,
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
      <TravelGameHero copy={copy} />

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
