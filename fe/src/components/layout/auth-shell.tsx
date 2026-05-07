import { Camera, Sparkles } from "lucide-react";
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
    <main className="relative min-h-screen overflow-hidden bg-[#090908] text-white">
      <ScenicImage
        alt={title}
        className="absolute inset-0 h-full w-full scale-[1.02] object-cover"
        fetchPriority="high"
        id="patagonia-trail"
        loading="eager"
        sizes="100vw"
      />
      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgb(0_0_0/0.28)_0%,rgb(0_0_0/0.58)_52%,rgb(0_0_0/0.78)_100%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_42%,transparent_0%,rgb(0_0_0/0.36)_72%)]" />

      <div className="relative z-10 flex min-h-screen flex-col">
        <header className="px-4 py-4 sm:px-6 lg:px-8">{topbar}</header>

        <section className="flex flex-1 items-center justify-center px-4 pb-8 pt-4 sm:px-6 lg:px-8">
          <div className="w-full max-w-[440px]">
            <div className="mb-5 text-center">
              <div className="mx-auto mb-3 inline-flex size-11 items-center justify-center rounded-full border border-white/18 bg-white/12 shadow-[0_18px_42px_-28px_rgb(0_0_0/0.9)] backdrop-blur-2xl">
                <Camera className="size-5" />
              </div>
              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-white/58">
                {label}
              </p>
              <h1 className="mt-2 text-4xl font-semibold tracking-normal text-white">
                {title}
              </h1>
              <p className="mx-auto mt-2 max-w-sm text-sm leading-6 text-white/68">
                {subtitle}
              </p>
            </div>

            <div className="rounded-[32px] border border-white/14 bg-[#111111]/58 p-5 shadow-[0_34px_100px_-46px_rgb(0_0_0/0.96)] backdrop-blur-2xl sm:p-6">
              <div className="space-y-6">
                {children}
                {footer ? <div>{footer}</div> : null}
              </div>
            </div>

            <div className="mx-auto mt-5 flex w-fit items-center gap-2 rounded-full border border-white/12 bg-white/8 px-3 py-1.5 text-xs font-medium text-white/70 backdrop-blur-xl">
              <Sparkles className="size-3.5 text-white/78" />
              {note || "Explore first. Save when a place feels worth remembering."}
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
