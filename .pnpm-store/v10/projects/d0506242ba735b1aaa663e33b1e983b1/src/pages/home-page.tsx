import { Sparkles, ArrowRight, PanelTopOpen } from "lucide-react";
import { motion } from "motion/react";
import { toast } from "sonner";
import { Button } from "../components/ui/button";

export function HomePage() {
  return (
    <main className="relative overflow-hidden px-6 py-12 text-ink md:px-10">
      <div className="pointer-events-none absolute inset-x-0 top-0 h-72 bg-[radial-gradient(circle_at_top,rgba(255,212,126,0.28),transparent_58%)]" />
      <div className="mx-auto flex min-h-[calc(100vh-6rem)] max-w-6xl items-center">
        <motion.section
          className="grid w-full gap-8 rounded-[32px] border border-white/60 bg-white/70 p-8 shadow-[0_24px_80px_rgba(16,37,66,0.14)] backdrop-blur md:grid-cols-[1.25fr_0.9fr] md:p-12"
          initial={{ opacity: 0, y: 18 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.45, ease: "easeOut" }}
        >
          <div className="space-y-6">
            <div className="inline-flex items-center gap-2 rounded-full border border-ink/10 bg-white/80 px-4 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-ink/70">
              <Sparkles className="h-4 w-4" />
              Falzo FE setup
            </div>

            <div className="space-y-4">
              <h1 className="max-w-3xl text-4xl font-black leading-none sm:text-5xl md:text-6xl">
                Bộ thư viện nền đã sẵn sàng để mình bắt đầu làm UI đẹp.
              </h1>
              <p className="max-w-2xl text-base leading-7 text-ink/75 sm:text-lg">
                Dự án hiện đã có routing, React Query, form validation, toast,
                icon, motion và Tailwind để mình dựng giao diện nhanh hơn mà vẫn
                giữ code gọn.
              </p>
            </div>

            <div className="flex flex-wrap gap-3">
              <Button onClick={() => toast.success("Falzo FE is ready to build.")}>
                Start building
                <ArrowRight className="h-4 w-4" />
              </Button>
              <Button
                className="bg-transparent"
                variant="secondary"
                onClick={() =>
                  toast("Installed core stack", {
                    description:
                      "react-router-dom, react-query, axios, zod, react-hook-form, Tailwind and more.",
                  })
                }
              >
                View stack
              </Button>
            </div>
          </div>

          <div className="grid gap-4">
            {[
              "React Router for page structure",
              "TanStack Query for API caching",
              "React Hook Form + Zod for forms",
              "Tailwind + utility helpers for UI",
            ].map((item) => (
              <div
                key={item}
                className="rounded-[24px] border border-ink/8 bg-[linear-gradient(180deg,rgba(255,255,255,0.92),rgba(244,248,255,0.92))] p-5 shadow-[0_12px_30px_rgba(16,37,66,0.08)]"
              >
                <div className="mb-3 inline-flex rounded-2xl bg-ink px-3 py-3 text-white">
                  <PanelTopOpen className="h-5 w-5" />
                </div>
                <p className="m-0 text-sm font-medium leading-6 text-ink/80">
                  {item}
                </p>
              </div>
            ))}
          </div>
        </motion.section>
      </div>
    </main>
  );
}
