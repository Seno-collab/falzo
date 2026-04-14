import {
  ArrowRight,
  Camera,
  Compass,
  MapPinned,
  Mountain,
  Sparkles,
  Sun,
} from "lucide-react";
import { motion } from "motion/react";
import { useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useLanguage } from "@/app/language-provider";
import { hasAuthSession } from "@/api/auth.api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { ScenicImage } from "@/components/scenic-image";
import { messages } from "@/i18n/messages";

const THEME_ICONS = {
  "alpine-lines": Mountain,
  "golden-coast": Sun,
  "heritage-streets": Compass,
} as const;

const HIGHLIGHT_ICONS = {
  "visual-first": Camera,
  "region-story": MapPinned,
  "promo-ready": Sparkles,
} as const;

export function HomePage() {
  const { language } = useLanguage();
  const navigate = useNavigate();
  const authenticated = hasAuthSession();
  const copy = messages[language].homePage;

  useEffect(() => {
    document.title = copy.documentTitle;
  }, [copy.documentTitle]);

  return (
    <div className="relative min-h-screen overflow-hidden bg-[linear-gradient(150deg,#ebf6ff_0%,#f1f7ff_44%,#fff3de_100%)] pb-28 sm:pb-16">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_6%_8%,rgba(85,150,214,0.23),transparent_27%),radial-gradient(circle_at_91%_9%,rgba(240,173,84,0.22),transparent_32%),radial-gradient(circle_at_52%_95%,rgba(78,166,155,0.16),transparent_26%)]" />
      <div className="falzo-float pointer-events-none absolute -left-14 top-24 h-60 w-60 rounded-full bg-[#77b5e5]/30 blur-3xl" />
      <div className="falzo-float-delayed pointer-events-none absolute -right-16 top-28 h-72 w-72 rounded-full bg-[#f5c77e]/24 blur-3xl" />

      <div className="relative mx-auto w-full max-w-7xl px-4 py-8 sm:px-6">
        <header className="sticky top-4 z-30 mb-8">
          <div className="falzo-glass flex flex-wrap items-center justify-between gap-3 rounded-2xl px-4 py-3 sm:px-5">
            <div className="flex items-center gap-2.5">
              <span className="flex h-7 w-7 items-center justify-center rounded-full bg-[#d8eaff] text-[#2267a8]">
                <Camera className="size-3.5" />
              </span>
              <p className="falzo-display text-xs font-semibold tracking-[0.18em] text-[#3c638d]">
                {copy.brand}
              </p>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Button asChild className="bg-[#1f4f87] text-white hover:bg-[#1a406f]">
                <Link to="/login">{copy.navLogin}</Link>
              </Button>
              <Button asChild className="border-[#b6cae1] bg-white/85 text-[#2d527d]" variant="outline">
                <Link to="/register">{copy.navRegister}</Link>
              </Button>
              {authenticated ? (
                <Button
                  className="border-[#b6cae1] bg-white/85 text-[#2d527d]"
                  onClick={() => navigate("/dashboard")}
                  type="button"
                  variant="outline"
                >
                  {copy.navDashboard}
                </Button>
              ) : null}
            </div>
          </div>
        </header>

        <section className="grid gap-5 lg:grid-cols-[1.06fr_0.94fr]">
          <motion.div
            animate={{ opacity: 1, y: 0 }}
            className="falzo-glass falzo-hover-lift space-y-6 rounded-[2rem] p-6 sm:p-8"
            initial={{ opacity: 0, y: 14 }}
            transition={{ duration: 0.32, ease: "easeOut" }}
          >
            <Badge className="w-fit bg-[#e7f2ff] text-[#275e98]" variant="secondary">
              {copy.heroBadge}
            </Badge>
            <h1 className="falzo-display text-4xl leading-[1.05] font-semibold tracking-tight text-[#132842] sm:text-5xl">
              {copy.heroTitle}
            </h1>
            <p className="max-w-2xl text-sm leading-7 text-[#48658a] sm:text-base">
              {copy.heroDescription}
            </p>

            <div className="flex flex-wrap gap-3">
              <Button
                className="h-10 bg-[#1f4f87] px-5 text-white hover:bg-[#1a406f]"
                onClick={() => navigate("/scenic-gallery")}
                type="button"
              >
                <Sparkles />
                {copy.heroPrimaryCta}
              </Button>
              <Button
                className="h-10 border-[#b6cae1] text-[#2d527d]"
                onClick={() => navigate("/scenic-gallery")}
                type="button"
                variant="outline"
              >
                {copy.heroSecondaryCta}
                <ArrowRight />
              </Button>
            </div>

            <ul className="flex flex-wrap gap-2.5 pt-1 text-xs text-[#456789] sm:text-sm">
              {copy.heroStats.map((stat) => (
                <li
                  className="rounded-full border border-[#cfe0f1] bg-[#f7fbff] px-3 py-1.5 font-medium"
                  key={stat}
                >
                  {stat}
                </li>
              ))}
            </ul>

            <div className="space-y-2">
              <p className="text-xs font-semibold tracking-[0.14em] text-[#5a79a0] uppercase">
                {copy.heroRegionLabel}
              </p>
              <div className="flex flex-wrap gap-2">
                {copy.regionTags.slice(0, 6).map((tag) => (
                  <span
                    className="rounded-full border border-[#c9dbef] bg-[#f2f8ff] px-2.5 py-1 text-xs font-medium text-[#365f8f]"
                    key={`hero-region-${tag}`}
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          </motion.div>

          <motion.div
            animate={{ opacity: 1, y: 0 }}
            className="falzo-hover-lift space-y-4 rounded-[2rem] border border-[#cddff2] bg-[linear-gradient(155deg,#0f2946_0%,#163557_56%,#1d3f63_100%)] p-6 text-[#eaf3fc] shadow-[0_34px_72px_-48px_rgba(9,22,38,0.95)] sm:p-7"
            initial={{ opacity: 0, y: 14 }}
            transition={{ duration: 0.36, delay: 0.05, ease: "easeOut" }}
          >
            <div className="space-y-1">
              <h2 className="falzo-display text-2xl font-semibold tracking-tight">{copy.heroSideTitle}</h2>
              <p className="text-xs leading-6 text-[#c7d9ee]">{copy.heroSideSubtitle}</p>
            </div>

            <div className={`relative h-44 overflow-hidden rounded-2xl ${copy.scenicGallery[0].imageGradient}`}>
              <ScenicImage
                alt={copy.scenicGallery[0].title}
                className="absolute inset-0 h-full w-full object-cover"
                id={copy.scenicGallery[0].id}
              />
              <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(0,0,0,0.02)_0%,rgba(0,0,0,0.3)_100%)]" />
              <div className="absolute left-3 top-3 rounded-full bg-white/20 px-2.5 py-1 text-xs font-medium text-white">
                {copy.scenicGallery[0].tag}
              </div>
              <div className="absolute inset-x-3 bottom-3 text-white">
                <p className="text-sm font-semibold">{copy.scenicGallery[0].title}</p>
                <p className="mt-1 inline-flex items-center gap-1 text-xs text-white/90">
                  <MapPinned className="size-3.5" />
                  {copy.scenicGallery[0].location}
                </p>
              </div>
            </div>

            <div className="space-y-2">
              {copy.scenicGallery.slice(1, 5).map((item, index) => (
                <div
                  className="flex items-center justify-between rounded-xl border border-[#2f506f] bg-[#173452] px-3 py-2"
                  key={item.id}
                >
                  <div className="space-y-0.5">
                    <p className="text-sm font-medium text-white">
                      {index + 2}. {item.title}
                    </p>
                    <p className="text-xs text-[#b7cfe8]">{item.location}</p>
                  </div>
                  <span className="text-xs text-[#d5e5f6]">{item.bestTime}</span>
                </div>
              ))}
            </div>

            <Button
              className="h-10 w-full bg-[#2f6ba8] text-white hover:bg-[#285f95]"
              onClick={() => navigate("/scenic-gallery")}
              type="button"
            >
              {copy.heroSideCta}
            </Button>
          </motion.div>
        </section>

        <section className="mt-12 space-y-4">
          <div className="space-y-1">
            <h3 className="falzo-display text-2xl font-semibold tracking-tight text-[#173457] sm:text-3xl">
              {copy.showcaseTitle}
            </h3>
            <p className="text-sm text-[#4c678b]">{copy.showcaseSubtitle}</p>
          </div>

          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {copy.scenicGallery.map((item, index) => (
              <motion.div
                initial={{ opacity: 0, y: 16 }}
                key={item.id}
                transition={{ duration: 0.3, delay: Math.min(index * 0.05, 0.24), ease: "easeOut" }}
                viewport={{ amount: 0.15, once: true }}
                whileInView={{ opacity: 1, y: 0 }}
              >
                <Card className="falzo-hover-lift overflow-hidden border border-[#d5e3f2] bg-white/90 py-0 shadow-[0_20px_44px_-36px_rgba(25,54,95,0.75)]">
                  <div className={`relative aspect-[4/3] w-full ${item.imageGradient}`}>
                    <ScenicImage
                      alt={item.title}
                      className="absolute inset-0 h-full w-full object-cover"
                      id={item.id}
                    />
                    <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(0,0,0,0.04)_0%,rgba(0,0,0,0.34)_100%)]" />
                    <div className="absolute left-3 top-3 rounded-full bg-white/18 px-2.5 py-1 text-xs font-medium text-white backdrop-blur">
                      {item.tag}
                    </div>
                    <div className="absolute inset-x-3 bottom-3 flex items-center justify-between text-xs text-white/90">
                      <span className="inline-flex items-center gap-1">
                        <MapPinned className="size-3.5" />
                        {item.location}
                      </span>
                      <span className="rounded-full bg-black/20 px-2 py-0.5">HDR</span>
                    </div>
                  </div>
                  <CardContent className="space-y-3 p-4">
                    <h4 className="text-lg font-semibold text-[#173354]">{item.title}</h4>
                    <div className="space-y-1.5 text-xs text-[#537193]">
                      <p>
                        {copy.bestTimeLabel}: <strong className="text-[#2a507e]">{item.bestTime}</strong>
                      </p>
                      <p>
                        {copy.moodLabel}: <strong className="text-[#2a507e]">{item.mood}</strong>
                      </p>
                    </div>
                    <Button
                      className="h-9 w-full border-[#bed3ea] text-[#1f4f87]"
                      onClick={() => navigate("/scenic-gallery")}
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

        <section className="mt-12 space-y-4">
          <div className="space-y-1">
            <h3 className="falzo-display text-2xl font-semibold tracking-tight text-[#173457] sm:text-3xl">
              {copy.themeTitle}
            </h3>
            <p className="text-sm text-[#4c678b]">{copy.themeSubtitle}</p>
          </div>

          <div className="grid gap-4 md:grid-cols-3">
            {copy.scenicThemes.map((theme, index) => {
              const Icon = THEME_ICONS[theme.id as keyof typeof THEME_ICONS] ?? Compass;

              return (
                <motion.div
                  initial={{ opacity: 0, y: 16 }}
                  key={theme.id}
                  transition={{ duration: 0.3, delay: Math.min(index * 0.06, 0.22), ease: "easeOut" }}
                  viewport={{ amount: 0.2, once: true }}
                  whileInView={{ opacity: 1, y: 0 }}
                >
                  <Card className="falzo-hover-lift border border-[#d7e4f3] bg-white/90 py-0 shadow-[0_20px_44px_-38px_rgba(25,54,95,0.65)]">
                    <CardContent className="space-y-3 p-5">
                      <Icon className="size-5 text-[#2b5e96]" />
                      <h4 className="text-lg font-semibold text-[#183559]">{theme.title}</h4>
                      <p className="text-sm leading-6 text-[#4b6689]">{theme.description}</p>
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
              );
            })}
          </div>
        </section>

        <section className="mt-12 grid gap-4 lg:grid-cols-[1.1fr_0.9fr]">
          <Card className="falzo-hover-lift border border-[#d7e4f3] bg-white/90 py-0 shadow-[0_20px_44px_-38px_rgba(25,54,95,0.65)]">
            <CardContent className="space-y-3 p-5 sm:p-6">
              <div className="space-y-1">
                <h3 className="falzo-display text-2xl font-semibold tracking-tight text-[#173457]">
                  {copy.highlightsTitle}
                </h3>
                <p className="text-sm text-[#4c678b]">{copy.highlightsSubtitle}</p>
              </div>

              <div className="grid gap-3 sm:grid-cols-3">
                {copy.brandHighlights.map((item) => {
                  const Icon = HIGHLIGHT_ICONS[item.id as keyof typeof HIGHLIGHT_ICONS] ?? Sparkles;

                  return (
                    <div
                      className="falzo-hover-lift rounded-2xl border border-[#d8e5f4] bg-[#f9fcff] p-4"
                      key={item.id}
                    >
                      <Icon className="mb-2 size-4 text-[#2a5c94]" />
                      <p className="text-sm font-semibold text-[#183559]">{item.title}</p>
                      <p className="mt-1.5 text-xs leading-5 text-[#4b6689]">{item.description}</p>
                    </div>
                  );
                })}
              </div>
            </CardContent>
          </Card>

          <Card className="falzo-hover-lift border border-[#d7e4f3] bg-white/90 py-0 shadow-[0_20px_44px_-38px_rgba(25,54,95,0.65)]">
            <CardContent className="space-y-3 p-5 sm:p-6">
              <h3 className="falzo-display text-xl font-semibold tracking-tight text-[#173457]">
                {copy.regionTitle}
              </h3>
              <p className="text-sm text-[#567396]">{copy.regionHint}</p>
              <div className="flex flex-wrap gap-2">
                {copy.regionTags.map((tag) => (
                  <span
                    className="rounded-full border border-[#cfe0f3] bg-[#f5faff] px-3 py-1.5 text-xs font-medium text-[#315f90]"
                    key={tag}
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </CardContent>
          </Card>
        </section>

        <section className="falzo-hover-lift mt-10 rounded-3xl border border-[#d4e3f5] bg-[linear-gradient(135deg,#123056_0%,#1f4f87_48%,#2a6cab_100%)] p-6 text-white shadow-[0_34px_76px_-52px_rgba(9,22,38,0.98)] sm:p-7">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="max-w-2xl space-y-1.5">
              <h3 className="falzo-display text-2xl font-semibold tracking-tight">{copy.bottomBannerTitle}</h3>
              <p className="text-sm text-[#d6e6f8]">{copy.bottomBannerSubtitle}</p>
            </div>
            <Button
              className="h-10 bg-white text-[#1f4f87] hover:bg-[#e9f2fb]"
              onClick={() => navigate(authenticated ? "/dashboard" : "/register")}
              type="button"
            >
              <Sparkles />
              {copy.bottomBannerCta}
            </Button>
          </div>
        </section>
      </div>

      <div className="fixed inset-x-4 bottom-4 z-40 md:hidden">
        <Button
          className="h-11 w-full bg-[#1f4f87] text-white shadow-[0_18px_34px_-16px_rgba(15,37,64,0.85)] backdrop-blur hover:bg-[#1a406f]"
          onClick={() => navigate("/scenic-gallery")}
          type="button"
        >
          <MapPinned />
          {copy.stickyMobileCta}
        </Button>
      </div>
    </div>
  );
}
