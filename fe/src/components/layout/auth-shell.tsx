import { BadgeCheck, ShieldCheck } from "lucide-react";
import type { ReactNode } from "react";
import { ScenicImage } from "@/components/scenic-image";

export function AuthShell({
  topbar,
  label,
  title,
  subtitle,
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
    <main className="relative min-h-svh overflow-x-hidden bg-[#f7f7f5] text-[#1f1f1f]">
      <ScenicImage
        alt={title}
        className="absolute inset-0 h-full w-full scale-[1.03] object-cover opacity-[0.24]"
        fetchPriority="high"
        id="patagonia-trail"
        loading="eager"
        sizes="100vw"
      />
      <div className="absolute inset-0 bg-[linear-gradient(115deg,rgb(247_247_245/0.98)_0%,rgb(247_247_245/0.94)_42%,rgb(239_246_252/0.82)_72%,rgb(247_247_245/0.72)_100%)]" />
      <div className="absolute inset-0 bg-[repeating-linear-gradient(135deg,rgb(255_255_255/0.34)_0px,rgb(255_255_255/0.34)_1px,transparent_1px,transparent_18px)] opacity-55" />
      <div className="absolute inset-x-0 bottom-0 h-2/5 bg-[linear-gradient(0deg,rgb(215_229_244/0.68)_0%,rgb(247_247_245/0)_100%)]" />
      <div className="absolute right-0 top-0 hidden h-full w-[42vw] border-[#d6e5f6]/70 border-l bg-white/22 backdrop-blur-[1px] lg:block" />

      <div className="relative z-10 flex min-h-svh flex-col">
        <header className="px-4 py-3 sm:px-6 sm:py-4 lg:px-8">{topbar}</header>

        <section className="flex flex-1 items-start justify-center px-4 pb-[calc(env(safe-area-inset-bottom)+2rem)] pt-1 sm:items-center sm:px-6 sm:pb-8 sm:pt-4 lg:px-8">
          <div className="w-full max-w-[440px]">
            <div className="mb-5 text-center">
              <div className="mx-auto mb-3 inline-flex size-11 items-center justify-center rounded-full border border-[#c8ddf1] bg-white text-[#2f6fb8] shadow-[0_18px_38px_-28px_rgb(32_72_116/0.7)]">
                <ShieldCheck className="size-5" />
              </div>
              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[#7892ad]">
                {label}
              </p>
              <h1 className="mt-2 text-4xl font-semibold tracking-normal text-[#143052]">
                {title}
              </h1>
              <p className="mx-auto mt-2 max-w-sm text-sm leading-6 text-[#5f7894]">
                {subtitle}
              </p>
            </div>

            <div className="rounded-3xl border border-[#d6e5f6] bg-white/90 p-5 shadow-[0_30px_86px_-46px_rgb(32_72_116/0.82)] backdrop-blur-2xl sm:p-6">
              <div className="space-y-6">
                {children}
                {footer ? <div>{footer}</div> : null}
              </div>
            </div>

            <div className="mx-auto mt-5 flex w-fit items-center gap-2 rounded-full border border-[#d6e5f6] bg-white/88 px-3 py-1.5 text-xs font-medium text-[#5f7894] shadow-[0_14px_34px_-28px_rgb(32_72_116/0.62)] backdrop-blur-xl">
              <BadgeCheck className="size-3.5 text-[#2f6fb8]" />
              {note || "Explore first. Save when a place feels worth remembering."}
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
