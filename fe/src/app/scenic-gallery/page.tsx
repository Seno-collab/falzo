"use client"

import { ArrowRight, Compass, MapPinned, Sparkles } from "lucide-react"
import { motion } from "motion/react"
import { useRouter } from "next/navigation"
import { useEffect, useMemo, useRef, useState } from "react"
import { useLanguage } from "@/app/language-provider"
import { hasAuthSession } from "@/api/auth.api"
import { EmptyState } from "@/components/feedback/empty-state"
import { AppTopbar } from "@/components/layout/app-topbar"
import { PageShell } from "@/components/layout/page-shell"
import { SectionHeading } from "@/components/layout/section-heading"
import { UserPresenceBadge } from "@/components/layout/user-presence-badge"
import { ScenicImage } from "@/components/scenic-image"
import { ScenicFieldNote } from "@/components/scenic/scenic-field-note"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { messages } from "@/i18n/messages"
import { ROUTES, getDashboardOrRegisterRoute } from "@/lib/routes"

type Point = {
  x: number
  y: number
}

type ScenicFrame = {
  id: string
  title: string
  location: string
  description: string
  mood: string
  bestTime: string
  tag: string
  imageGradient: string
  baseX: number
  baseY: number
  depth: number
  rotate: number
}

function buildInitialOffsets(frames: readonly ScenicFrame[]) {
  return frames.reduce<Record<string, Point>>((accumulator, frame) => {
    accumulator[frame.id] = { x: 0, y: 0 }
    return accumulator
  }, {})
}

function Draggable3DCard({
  frame,
  offset,
  isActive,
  onFocus,
  onOffsetChange,
}: {
  frame: ScenicFrame
  offset: Point
  isActive: boolean
  onFocus: (id: string) => void
  onOffsetChange: (id: string, next: Point) => void
}) {
  const cardRef = useRef<HTMLDivElement>(null)
  const dragRef = useRef<{
    startPointerX: number
    startPointerY: number
    startOffsetX: number
    startOffsetY: number
  } | null>(null)
  const [dragging, setDragging] = useState(false)
  const [tilt, setTilt] = useState({ x: 0, y: 0 })

  const updateTilt = (clientX: number, clientY: number) => {
    if (!cardRef.current) {
      return
    }

    const rect = cardRef.current.getBoundingClientRect()
    const percentX = (clientX - rect.left) / rect.width - 0.5
    const percentY = (clientY - rect.top) / rect.height - 0.5

    setTilt({
      x: -percentY * 14,
      y: percentX * 16,
    })
  }

  const clearTilt = () => {
    setTilt({ x: 0, y: 0 })
  }

  return (
    <motion.div
      className="absolute touch-none select-none"
      style={{
        left: `${frame.baseX}%`,
        top: `${frame.baseY}%`,
        width: "clamp(164px, 25vw, 250px)",
        zIndex: dragging ? 90 : isActive ? 80 : 40 + Math.round(frame.depth / 4),
        transform: `translate3d(${offset.x}px, ${offset.y}px, ${frame.depth}px) rotateX(${tilt.x}deg) rotateY(${tilt.y}deg) rotateZ(${dragging ? 0 : frame.rotate}deg) scale(${dragging ? 1.03 : isActive ? 1.01 : 1})`,
        transformStyle: "preserve-3d",
        transition: dragging ? "none" : "transform 180ms ease",
      }}
      whileTap={{ scale: 1.02 }}
    >
      <div
        className={`falzo-hover-lift overflow-hidden rounded-2xl border ${
          isActive ? "border-[#8fb7df]" : "border-[#d7e5f4]"
        } bg-white/93 shadow-[0_20px_46px_-30px_rgba(21,49,85,0.74)]`}
        onPointerCancel={() => {
          dragRef.current = null
          setDragging(false)
          clearTilt()
        }}
        onPointerDown={(event) => {
          event.preventDefault()
          event.currentTarget.setPointerCapture(event.pointerId)
          dragRef.current = {
            startPointerX: event.clientX,
            startPointerY: event.clientY,
            startOffsetX: offset.x,
            startOffsetY: offset.y,
          }
          setDragging(true)
          onFocus(frame.id)
        }}
        onPointerLeave={() => {
          if (!dragging) {
            clearTilt()
          }
        }}
        onPointerMove={(event) => {
          const dragData = dragRef.current
          if (!dragData) {
            updateTilt(event.clientX, event.clientY)
            return
          }

          const deltaX = event.clientX - dragData.startPointerX
          const deltaY = event.clientY - dragData.startPointerY
          onOffsetChange(frame.id, {
            x: dragData.startOffsetX + deltaX,
            y: dragData.startOffsetY + deltaY,
          })
          updateTilt(event.clientX, event.clientY)
        }}
        onPointerUp={(event) => {
          if (event.currentTarget.hasPointerCapture(event.pointerId)) {
            event.currentTarget.releasePointerCapture(event.pointerId)
          }
          dragRef.current = null
          setDragging(false)
          clearTilt()
        }}
        ref={cardRef}
      >
        <div className={`relative aspect-[4/3] w-full ${frame.imageGradient}`}>
          <ScenicImage
            alt={frame.title}
            className="absolute inset-0 h-full w-full object-cover"
            id={frame.id}
          />
          <div className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(0,0,0,0.04)_0%,rgba(0,0,0,0.34)_100%)]" />
          <div className="absolute left-2.5 top-2.5 rounded-full bg-white/24 px-2 py-0.5 text-[11px] font-medium text-white backdrop-blur">
            {frame.tag}
          </div>
          <div className="absolute inset-x-2.5 bottom-2.5 text-[11px] text-white/90">
            <span className="inline-flex items-center gap-1">
              <MapPinned className="size-3.5" />
              {frame.location}
            </span>
          </div>
        </div>

        <div className="space-y-1 p-3">
          <p className="line-clamp-2 text-sm font-semibold text-[#183559]">{frame.title}</p>
          <p className="line-clamp-2 text-xs leading-5 text-[#527093]">{frame.description}</p>
        </div>
      </div>
    </motion.div>
  )
}

export default function ScenicGalleryRoutePage() {
  const { language } = useLanguage()
  const router = useRouter()
  const [authenticated, setAuthenticated] = useState(false)
  const copy = messages[language].scenicGalleryPage
  const homeCopy = messages[language].homePage
  const frames = copy.frames as readonly ScenicFrame[]

  const [activeFrameId, setActiveFrameId] = useState<string>(() => frames[0]?.id ?? "")
  const [offsets, setOffsets] = useState<Record<string, Point>>(() => buildInitialOffsets(frames))
  const [viewerFrameId, setViewerFrameId] = useState<string | null>(null)

  useEffect(() => {
    document.title = copy.documentTitle
  }, [copy.documentTitle])

  useEffect(() => {
    setAuthenticated(hasAuthSession())
  }, [])

  useEffect(() => {
    if (!frames.length) {
      return
    }

    setActiveFrameId((previous) => {
      if (frames.some((frame) => frame.id === previous)) {
        return previous
      }

      return frames[0].id
    })
    setOffsets(buildInitialOffsets(frames))
  }, [frames])

  const activeFrame = useMemo(
    () => frames.find((frame) => frame.id === activeFrameId) ?? frames[0],
    [activeFrameId, frames],
  )
  const viewerFrame = useMemo(
    () => frames.find((frame) => frame.id === viewerFrameId) ?? null,
    [frames, viewerFrameId],
  )

  if (!activeFrame) {
    return (
      <PageShell
        topbar={
          <AppTopbar
            actions={[
              {
                id: "home",
                label: copy.homeCta,
                to: ROUTES.home,
                variant: "outline",
              },
            ]}
            brand={homeCopy.brand}
            brandIcon={<Sparkles className="size-3.5" />}
            meta={<UserPresenceBadge />}
            mobileMenuTitle={copy.pageTitle}
          />
        }
      >
        <EmptyState
          description={language === "vi" ? "Hiện chưa có dữ liệu hình ảnh để hiển thị." : "No image data is currently available."}
          title={language === "vi" ? "Không có khung ảnh" : "No frames available"}
        />
      </PageShell>
    )
  }

  return (
    <PageShell
      contentClassName="pb-24 sm:pb-14"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "login",
              label: copy.loginCta,
              to: ROUTES.login,
              variant: "outline",
            },
            {
              id: "home",
              label: copy.homeCta,
              to: ROUTES.home,
              variant: "soft",
            },
          ]}
          brand={copy.navLabel}
          brandIcon={<Compass className="size-3.5" />}
          meta={<UserPresenceBadge />}
          mobileMenuTitle={copy.pageTitle}
          subtitle={copy.pageSubtitle}
        />
      }
    >
      <section className="app-section space-y-1">
        <SectionHeading description={copy.pageSubtitle} kicker={copy.navLabel} title={copy.pageTitle} />
      </section>

      <section className="mt-5 grid gap-5 lg:grid-cols-[1.15fr_0.85fr]">
        <motion.div
          animate={{ opacity: 1, y: 0 }}
          className="app-panel falzo-topography p-5"
          initial={{ opacity: 0, y: 14 }}
          transition={{ duration: 0.32, ease: "easeOut" }}
        >
          <div className="flex flex-wrap items-center justify-between gap-2">
            <Badge>{copy.boardBadge}</Badge>
            <Button
              onClick={() => setOffsets(buildInitialOffsets(frames))}
              size="sm"
              type="button"
              variant="outline"
            >
              {copy.resetCta}
            </Button>
          </div>

          <h2 className="falzo-display mt-3 text-2xl font-semibold tracking-tight text-[#173456]">
            {copy.boardTitle}
          </h2>
          <p className="mt-1 text-sm text-[#4b678b]">{copy.boardHint}</p>

          <div
            className="relative mt-4 h-[620px] overflow-hidden rounded-[1.75rem] border border-[#d1e2f2] bg-[radial-gradient(circle_at_35%_18%,#f7fcff_0%,#eff5ff_50%,#e5eef9_100%)] shadow-[0_30px_70px_-50px_rgba(20,52,90,0.85)]"
            style={{ perspective: "1400px" }}
          >
            <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_78%_14%,rgba(98,175,238,0.22),transparent_35%),radial-gradient(circle_at_20%_86%,rgba(246,188,106,0.22),transparent_36%)]" />
            {frames.map((frame) => (
              <Draggable3DCard
                frame={frame}
                isActive={activeFrameId === frame.id}
                key={frame.id}
                offset={offsets[frame.id] ?? { x: 0, y: 0 }}
                onFocus={setActiveFrameId}
                onOffsetChange={(id, next) => {
                  setOffsets((previous) => ({
                    ...previous,
                    [id]: next,
                  }))
                }}
              />
            ))}
          </div>
        </motion.div>

        <motion.div
          animate={{ opacity: 1, y: 0 }}
          className="app-panel space-y-4 border-[#d5e4f4] bg-white/90 p-5 sm:p-6"
          initial={{ opacity: 0, y: 14 }}
          transition={{ duration: 0.36, delay: 0.04, ease: "easeOut" }}
        >
          <Badge>{copy.detailsBadge}</Badge>
          <h2 className="falzo-display text-2xl font-semibold tracking-tight text-[#15304d]">
            {copy.detailsTitle}
          </h2>

          <Card className="border-[#d9e6f5] bg-white py-0">
            <CardContent className="space-y-3 p-5">
              <button
                className={`relative block aspect-[16/10] w-full overflow-hidden rounded-xl ${activeFrame.imageGradient}`}
                onClick={() => setViewerFrameId(activeFrame.id)}
                type="button"
              >
                <ScenicImage
                  alt={activeFrame.title}
                  className="absolute inset-0 h-full w-full object-cover"
                  id={activeFrame.id}
                />
                <span className="pointer-events-none absolute inset-0 bg-[linear-gradient(180deg,rgba(0,0,0,0.04)_0%,rgba(0,0,0,0.22)_100%)]" />
              </button>
              <h3 className="text-xl font-semibold text-[#173457]">{activeFrame.title}</h3>
              <p className="text-sm leading-6 text-[#4a668b]">{activeFrame.description}</p>

              <ScenicFieldNote
                bestTime={activeFrame.bestTime}
                language={language}
                location={activeFrame.location}
                mood={activeFrame.mood}
                tag={activeFrame.tag}
              />

              <Button
                onClick={() => setViewerFrameId(activeFrame.id)}
                type="button"
                variant="outline"
              >
                {copy.openViewerCta}
              </Button>
            </CardContent>
          </Card>

          <div className="rounded-2xl border border-[#d8e5f3] bg-[#f7fbff] p-4">
            <p className="text-sm font-semibold text-[#1d3c62]">{copy.highlightsTitle}</p>
            <ul className="mt-2 space-y-1.5 text-sm text-[#436183]">
              {copy.highlights.map((item) => (
                <li className="inline-flex items-center gap-2" key={item}>
                  <span className="inline-block h-1.5 w-1.5 rounded-full bg-[#2f6ba8]" />
                  {item}
                </li>
              ))}
            </ul>
          </div>

          <div className="flex flex-wrap gap-2">
            <Button onClick={() => router.push(ROUTES.home)} type="button" variant="gradient">
              <Compass className="size-4" />
              {copy.primaryCta}
            </Button>
            <Button
              onClick={() => router.push(getDashboardOrRegisterRoute(authenticated))}
              type="button"
              variant="outline"
            >
              <ArrowRight className="size-4" />
              {copy.secondaryCta}
            </Button>
          </div>
        </motion.div>
      </section>

      <Dialog
        onOpenChange={(open) => {
          if (!open) {
            setViewerFrameId(null)
          }
        }}
        open={Boolean(viewerFrame)}
      >
        <DialogContent showCloseButton={false}>
          {viewerFrame ? (
            <>
              <DialogHeader className="border-b border-white/14 pb-3">
                <DialogTitle>{viewerFrame.title}</DialogTitle>
                <DialogDescription>{viewerFrame.location}</DialogDescription>
              </DialogHeader>

              <div
                className={`relative flex h-[min(74vh,760px)] items-center justify-center rounded-xl ${viewerFrame.imageGradient}`}
              >
                <ScenicImage
                  alt={viewerFrame.title}
                  className="max-h-full max-w-full rounded-lg object-contain"
                  fetchPriority="high"
                  id={viewerFrame.id}
                  loading="eager"
                  sizes="100vw"
                />
              </div>

              <div className="flex flex-wrap items-center justify-between gap-2 border-t border-white/14 pt-3">
                <p className="text-xs text-[#d1def0]">{copy.viewerHint}</p>
                <Button
                  onClick={() => setViewerFrameId(null)}
                  size="sm"
                  type="button"
                  variant="outline"
                >
                  {copy.closeViewerCta}
                </Button>
              </div>
            </>
          ) : null}
        </DialogContent>
      </Dialog>

      <div className="fixed inset-x-4 bottom-4 z-40 md:hidden">
        <Button
          className="h-11 w-full"
          onClick={() => router.push(ROUTES.home)}
          type="button"
          variant="gradient"
        >
          <Sparkles className="size-4" />
          {copy.stickyCta}
        </Button>
      </div>
    </PageShell>
  )
}
