"use client"

import {
  ArrowRight,
  Camera,
  Compass,
  MapPinned,
  Mountain,
  Sparkles,
  Sun,
} from "lucide-react"
import { motion } from "motion/react"
import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { useLanguage } from "@/app/language-provider"
import { hasAuthSession } from "@/api/auth.api"
import { AppTopbar } from "@/components/layout/app-topbar"
import { PageShell } from "@/components/layout/page-shell"
import { SectionHeading } from "@/components/layout/section-heading"
import { UserPresenceBadge } from "@/components/layout/user-presence-badge"
import { ScenicImage } from "@/components/scenic-image"
import { ScenicFieldNote } from "@/components/scenic/scenic-field-note"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { messages } from "@/i18n/messages"
import { ROUTES, getDashboardOrRegisterRoute } from "@/lib/routes"

const THEME_ICONS = {
  "alpine-lines": Mountain,
  "golden-coast": Sun,
  "heritage-streets": Compass,
} as const

const HIGHLIGHT_ICONS = {
  "visual-first": Camera,
  "region-story": MapPinned,
  "promo-ready": Sparkles,
} as const

export default function RootPage() {
  const { language } = useLanguage()
  const router = useRouter()
  const [authenticated, setAuthenticated] = useState(false)
  const copy = messages[language].homePage
  const commonCopy = messages[language].common

  useEffect(() => {
    document.title = copy.documentTitle
  }, [copy.documentTitle])

  useEffect(() => {
    setAuthenticated(hasAuthSession())
  }, [])

  return (
    <PageShell
      contentClassName="pb-24 sm:pb-16"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "login",
              label: copy.navLogin,
              to: ROUTES.login,
              variant: "outline",
            },
            {
              id: "register",
              label: copy.navRegister,
              to: ROUTES.register,
              variant: "default",
            },
            ...(authenticated
              ? [
                  {
                    id: "dashboard",
                    label: copy.navDashboard,
                    to: ROUTES.dashboard,
                    variant: "soft" as const,
                  },
                ]
              : []),
          ]}
          brand={copy.brand}
          brandIcon={<Camera className="size-3.5" />}
          meta={<UserPresenceBadge />}
          mobileMenuTitle={commonCopy.appName}
          subtitle={copy.heroDescription}
        />
      }
    >
      <section className="grid gap-5 lg:grid-cols-[1.05fr_0.95fr]">
        <motion.div
          animate={{ opacity: 1, y: 0 }}
          className="app-panel app-hover falzo-topography space-y-6 p-6 sm:p-8"
          initial={{ opacity: 0, y: 14 }}
          transition={{ duration: 0.3, ease: "easeOut" }}
        >
          <div className="space-y-3">
            <Badge>{copy.heroBadge}</Badge>
            <h1 className="app-title max-w-3xl text-4xl sm:text-5xl">{copy.heroTitle}</h1>
            <p className="app-subtitle max-w-3xl text-sm sm:text-base">{copy.heroDescription}</p>
          </div>

          <div className="flex flex-wrap gap-2.5">
            <Button
              onClick={() => router.push(ROUTES.scenicGallery)}
              type="button"
              variant="gradient"
            >
              <Sparkles className="size-4" />
              {copy.heroPrimaryCta}
            </Button>
            <Button
              onClick={() => router.push(ROUTES.scenicGallery)}
              type="button"
              variant="outline"
            >
              {copy.heroSecondaryCta}
              <ArrowRight className="size-4" />
            </Button>
          </div>

          <ul className="flex flex-wrap gap-2.5">
            {copy.heroStats.map((stat) => (
              <li className="app-chip" key={stat}>
                {stat}
              </li>
            ))}
          </ul>

          <ScenicFieldNote
            bestTime={copy.scenicGallery[0].bestTime}
            language={language}
            location={copy.scenicGallery[0].location}
            mood={copy.scenicGallery[0].mood}
            tag={copy.scenicGallery[0].tag}
          />

          <div className="space-y-2">
            <p className="app-kicker">{copy.heroRegionLabel}</p>
            <div className="flex flex-wrap gap-2">
              {copy.regionTags.slice(0, 6).map((tag) => (
                <span className="app-chip" key={`hero-region-${tag}`}>
                  {tag}
                </span>
              ))}
            </div>
          </div>
        </motion.div>

        <motion.div
          animate={{ opacity: 1, y: 0 }}
          className="app-panel app-hover falzo-topography space-y-4 border-[#d4e4f6] bg-[linear-gradient(145deg,#0f2b4a_0%,#1a3f66_55%,#234f7d_100%)] p-6 text-[#eef5fd] sm:p-7"
          initial={{ opacity: 0, y: 14 }}
          transition={{ duration: 0.34, delay: 0.05, ease: "easeOut" }}
        >
          <div className="space-y-1.5">
            <h2 className="falzo-display text-2xl font-semibold tracking-tight sm:text-3xl">
              {copy.heroSideTitle}
            </h2>
            <p className="text-xs leading-6 text-[#c9dcee] sm:text-sm">{copy.heroSideSubtitle}</p>
          </div>

          <div className={`relative h-48 overflow-hidden rounded-2xl ${copy.scenicGallery[0].imageGradient}`}>
            <ScenicImage
              alt={copy.scenicGallery[0].title}
              className="absolute inset-0 h-full w-full object-cover"
              id={copy.scenicGallery[0].id}
            />
            <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(0,0,0,0.03)_0%,rgba(0,0,0,0.34)_100%)]" />
            <div className="absolute left-3 top-3 rounded-full bg-white/24 px-2.5 py-1 text-xs font-medium text-white backdrop-blur">
              {copy.scenicGallery[0].tag}
            </div>
            <div className="absolute inset-x-3 bottom-3 text-white">
              <p className="text-sm font-semibold">{copy.scenicGallery[0].title}</p>
              <p className="mt-1 inline-flex items-center gap-1 text-xs text-white/88">
                <MapPinned className="size-3.5" />
                {copy.scenicGallery[0].location}
              </p>
            </div>
          </div>

          <div className="space-y-2">
            {copy.scenicGallery.slice(1, 5).map((item, index) => (
              <div
                className="flex items-center justify-between rounded-xl border border-white/16 bg-white/8 px-3 py-2.5"
                key={item.id}
              >
                <div className="space-y-0.5">
                  <p className="text-sm font-medium text-white">
                    {index + 2}. {item.title}
                  </p>
                  <p className="text-xs text-[#bed3ea]">{item.location}</p>
                </div>
                <span className="text-xs text-[#d7e6f7]">{item.bestTime}</span>
              </div>
            ))}
          </div>

          <Button
            className="w-full"
            onClick={() => router.push(ROUTES.scenicGallery)}
            type="button"
            variant="gradient"
          >
            {copy.heroSideCta}
          </Button>
        </motion.div>
      </section>

      <section className="mt-11 app-section">
        <SectionHeading description={copy.showcaseSubtitle} title={copy.showcaseTitle} />

        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {copy.scenicGallery.map((item, index) => (
            <motion.div
              initial={{ opacity: 0, y: 14 }}
              key={item.id}
              transition={{ duration: 0.28, delay: Math.min(index * 0.04, 0.2), ease: "easeOut" }}
              viewport={{ amount: 0.2, once: true }}
              whileInView={{ opacity: 1, y: 0 }}
            >
              <Card className="app-hover overflow-hidden border-[#d6e4f4] bg-white/90 py-0">
                <div className={`relative aspect-[4/3] w-full ${item.imageGradient}`}>
                  <ScenicImage
                    alt={item.title}
                    className="absolute inset-0 h-full w-full object-cover"
                    id={item.id}
                  />
                  <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(0,0,0,0.04)_0%,rgba(0,0,0,0.33)_100%)]" />
                  <div className="absolute left-3 top-3 rounded-full bg-white/24 px-2.5 py-1 text-xs font-medium text-white backdrop-blur">
                    {item.tag}
                  </div>
                  <div className="absolute inset-x-3 bottom-3 flex items-center justify-between text-xs text-white/90">
                    <span className="inline-flex items-center gap-1">
                      <MapPinned className="size-3.5" />
                      {item.location}
                    </span>
                    <span className="rounded-full bg-black/22 px-2 py-0.5">HDR</span>
                  </div>
                </div>

                <CardContent className="space-y-3 p-4">
                  <h3 className="text-lg font-semibold text-[#173a61]">{item.title}</h3>
                  <div className="space-y-1.5 text-xs text-[#567496]">
                    <p>
                      {copy.bestTimeLabel}: <strong className="text-[#284f7e]">{item.bestTime}</strong>
                    </p>
                    <p>
                      {copy.moodLabel}: <strong className="text-[#284f7e]">{item.mood}</strong>
                    </p>
                  </div>
                  <Button
                    className="w-full"
                    onClick={() => router.push(ROUTES.scenicGallery)}
                    type="button"
                    variant="outline"
                  >
                    {copy.viewCollectionCta}
                  </Button>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>
      </section>

      <section className="mt-11 app-section">
        <SectionHeading description={copy.themeSubtitle} title={copy.themeTitle} />

        <div className="grid gap-4 md:grid-cols-3">
          {copy.scenicThemes.map((theme, index) => {
            const Icon = THEME_ICONS[theme.id as keyof typeof THEME_ICONS] ?? Compass

            return (
              <motion.div
                initial={{ opacity: 0, y: 14 }}
                key={theme.id}
                transition={{ duration: 0.28, delay: Math.min(index * 0.05, 0.2), ease: "easeOut" }}
                viewport={{ amount: 0.2, once: true }}
                whileInView={{ opacity: 1, y: 0 }}
              >
                <Card className="app-hover border-[#d7e4f3] bg-white/90 py-0">
                  <CardContent className="space-y-3 p-5">
                    <Icon className="size-5 text-[#2e639c]" />
                    <h3 className="text-lg font-semibold text-[#193a60]">{theme.title}</h3>
                    <p className="text-sm leading-6 text-[#4f6d8f]">{theme.description}</p>
                    <div className="space-y-1.5 text-xs text-[#567395]">
                      {theme.points.map((point) => (
                        <p className="inline-flex items-center gap-1.5" key={point}>
                          <span className="inline-block h-1.5 w-1.5 rounded-full bg-[#4f80b4]" />
                          {point}
                        </p>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            )
          })}
        </div>
      </section>

      <section className="mt-11 grid gap-4 lg:grid-cols-[1.08fr_0.92fr]">
        <Card className="app-hover border-[#d7e4f3] bg-white/90 py-0">
          <CardContent className="space-y-4 p-5 sm:p-6">
            <SectionHeading description={copy.highlightsSubtitle} title={copy.highlightsTitle} />
            <div className="grid gap-3 sm:grid-cols-3">
              {copy.brandHighlights.map((item) => {
                const Icon = HIGHLIGHT_ICONS[item.id as keyof typeof HIGHLIGHT_ICONS] ?? Sparkles

                return (
                  <div
                    className="app-hover rounded-2xl border border-[#d7e5f4] bg-[#f8fbff] p-4"
                    key={item.id}
                  >
                    <Icon className="mb-2 size-4 text-[#2e639c]" />
                    <p className="text-sm font-semibold text-[#193a60]">{item.title}</p>
                    <p className="mt-1.5 text-xs leading-5 text-[#4f6d8f]">{item.description}</p>
                  </div>
                )
              })}
            </div>
          </CardContent>
        </Card>

        <Card className="app-hover border-[#d7e4f3] bg-white/90 py-0">
          <CardContent className="space-y-3 p-5 sm:p-6">
            <SectionHeading description={copy.regionHint} title={copy.regionTitle} />
            <div className="flex flex-wrap gap-2">
              {copy.regionTags.map((tag) => (
                <span className="app-chip" key={tag}>
                  {tag}
                </span>
              ))}
            </div>
          </CardContent>
        </Card>
      </section>

      <section className="app-panel app-hover mt-10 overflow-hidden border-[#d3e3f4] bg-[linear-gradient(135deg,#123056_0%,#1f4f87_48%,#2a6cab_100%)] p-6 text-white shadow-[0_34px_74px_-46px_rgb(10_24_40/0.88)] sm:p-7">
        <div className="absolute -left-16 -bottom-16 h-52 w-52 rounded-full bg-white/12 blur-3xl" />
        <div className="absolute -right-12 -top-14 h-44 w-44 rounded-full bg-[#f5ca86]/30 blur-3xl" />
        <div className="relative flex flex-wrap items-center justify-between gap-4">
          <div className="max-w-2xl space-y-1.5">
            <h3 className="falzo-display text-2xl font-semibold tracking-tight">{copy.bottomBannerTitle}</h3>
            <p className="text-sm text-[#d9e8f8]">{copy.bottomBannerSubtitle}</p>
          </div>
          <Button
            className="min-w-36"
            onClick={() => router.push(getDashboardOrRegisterRoute(authenticated))}
            type="button"
            variant="outline"
          >
            <Sparkles className="size-4" />
            {copy.bottomBannerCta}
          </Button>
        </div>
      </section>

      <div className="fixed inset-x-4 bottom-4 z-40 md:hidden">
        <Button
          className="h-11 w-full"
          onClick={() => router.push(ROUTES.scenicGallery)}
          type="button"
          variant="gradient"
        >
          <MapPinned className="size-4" />
          {copy.stickyMobileCta}
        </Button>
      </div>
    </PageShell>
  )
}
