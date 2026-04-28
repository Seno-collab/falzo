import { ShieldCheck, Sparkles } from "lucide-react";
import type { ReactNode } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { PageShell } from "@/components/layout/page-shell";

export function AuthShell({
  topbar,
  label,
  title,
  subtitle,
  points,
  note,
  children,
  footer,
}: Readonly<{
  topbar: ReactNode;
  label: string;
  title: string;
  subtitle: string;
  points: string[];
  note: string;
  children: ReactNode;
  footer?: ReactNode;
}>) {
  return (
    <PageShell contentClassName="pb-12 md:pb-16" topbar={topbar}>
      <div className="grid gap-5 lg:grid-cols-[1.02fr_0.98fr]">
        <section className="app-panel app-hover relative hidden overflow-hidden p-8 text-white shadow-[0_40px_90px_-48px_rgb(11_28_49/0.86)] lg:block">
          <div className="absolute inset-0 -z-10 bg-[linear-gradient(140deg,#0f2a49_0%,#1d4a7b_56%,#2d6daf_100%)]" />
          <div className="absolute -top-8 right-2 -z-10 h-32 w-32 rounded-full bg-white/15 blur-2xl" />
          <div className="absolute -bottom-12 left-8 -z-10 h-36 w-36 rounded-full bg-[#f3c782]/30 blur-3xl" />

          <p className="app-kicker text-[#d6e7f9]">{label}</p>
          <h1 className="falzo-display mt-2 text-4xl leading-tight font-semibold">
            {title}
          </h1>
          <p className="mt-3 max-w-lg text-sm leading-7 text-[#d8e8f8]">
            {subtitle}
          </p>

          <div className="mt-7 space-y-3">
            {points.map((point) => (
              <div
                className="flex items-start gap-2.5 rounded-xl border border-white/20 bg-white/10 px-3 py-2.5 backdrop-blur"
                key={point}
              >
                <span className="mt-0.5 inline-flex size-5 items-center justify-center rounded-full bg-white/16">
                  <ShieldCheck className="size-3.5" />
                </span>
                <p className="text-sm text-[#ecf4fd]">{point}</p>
              </div>
            ))}
          </div>

          <div className="mt-6 inline-flex items-center gap-2 rounded-full border border-white/24 bg-white/12 px-3 py-1.5 text-xs font-medium text-[#eef6ff]">
            <Sparkles className="size-3.5" />
            {note}
          </div>
        </section>

        <Card className="app-panel app-hover border-[#d7e6f7] bg-white/94 py-0">
          <CardContent className="space-y-6 p-6 sm:p-8">
            {children}
            {footer ? <div>{footer}</div> : null}
          </CardContent>
        </Card>
      </div>
    </PageShell>
  );
}
