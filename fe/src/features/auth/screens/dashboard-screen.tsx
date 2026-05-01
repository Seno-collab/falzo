"use client";

import {
  BarChart3,
  Compass,
  House,
  LogOut,
  Sparkles,
  User,
} from "lucide-react";
import { motion } from "motion/react";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { LoadingPanel } from "@/components/feedback/loading-panel";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import { SectionHeading } from "@/components/layout/section-heading";
import { UserPresenceBadge } from "@/components/layout/user-presence-badge";
import { ScenicFieldNote } from "@/components/scenic/scenic-field-note";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  clearAuthSession,
  getMeApi,
  hasAuthSession,
  logoutApi,
} from "@/features/auth/api";
import { messages } from "@/i18n/messages";
import { ROUTES } from "@/lib/routes";

function resolveUserDisplayName(payload: unknown): string | null {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    return null;
  }

  const data = payload as Record<string, unknown>;
  for (const key of ["fullName", "name", "displayName", "email"]) {
    const value = data[key];
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }

  return null;
}

export function DashboardScreen() {
  const router = useRouter();
  const copy = messages.en.dashboardPage;
  const featuredFrame = messages.en.homePage.scenicGallery[0];

  const [isSessionChecking, setIsSessionChecking] = useState(true);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const [userDisplayName, setUserDisplayName] = useState<string | null>(null);

  useEffect(() => {
    document.title = copy.documentTitle;

    if (!hasAuthSession()) {
      router.replace(ROUTES.login);
      return;
    }

    let disposed = false;

    const validateSession = async () => {
      try {
        const profile = await getMeApi<Record<string, unknown>>();
        if (!disposed) {
          setUserDisplayName(resolveUserDisplayName(profile));
          setIsSessionChecking(false);
        }
      } catch {
        if (disposed) {
          return;
        }

        clearAuthSession();
        router.replace(ROUTES.login);
      }
    };

    void validateSession();

    return () => {
      disposed = true;
    };
  }, [copy.documentTitle, router]);

  const handleLogout = async () => {
    if (isLoggingOut) {
      return;
    }

    setIsLoggingOut(true);

    try {
      await logoutApi();
      router.replace(ROUTES.login);
    } finally {
      setIsLoggingOut(false);
    }
  };

  return (
    <PageShell
      topbar={
        <AppTopbar
          actions={[
            {
              id: "gallery",
              icon: <Compass className="size-4" />,
              label: copy.open3DDemoCta,
              to: ROUTES.explore,
              variant: "outline",
            },
            {
              id: "home",
              icon: <House className="size-4" />,
              label: copy.backToLandingCta,
              to: ROUTES.home,
              variant: "outline",
            },
            {
              id: "logout",
              icon: <LogOut className="size-4" />,
              label: copy.logoutCta,
              onClick: () => {
                void handleLogout();
              },
              variant: "default",
            },
          ]}
          brand={copy.label}
          brandIcon={<BarChart3 className="size-3.5" />}
          meta={<UserPresenceBadge />}
          mobileMenuTitle={copy.title}
          subtitle={copy.subtitle}
        />
      }
    >
      {isSessionChecking ? (
        <LoadingPanel
          description={copy.subtitle}
          title="Verifying your session"
        />
      ) : (
        <motion.div
          animate={{ opacity: 1, y: 0 }}
          className="space-y-4"
          initial={{ opacity: 0, y: 10 }}
          transition={{ duration: 0.24, ease: "easeOut" }}
        >
          <Card className="app-panel app-hover border-[#d6e5f6] bg-white/92 py-0">
            <CardContent className="space-y-5 p-6 sm:p-8">
              <div className="space-y-2">
                <Badge>{copy.label}</Badge>
                <SectionHeading
                  description={copy.subtitle}
                  title={copy.title}
                />
                {userDisplayName ? (
                  <p className="text-sm text-[#527299]">
                    Welcome, {userDisplayName}
                  </p>
                ) : null}
              </div>

              <div className="grid gap-3 sm:grid-cols-3">
                {[
                  copy.open3DDemoCta,
                  copy.backToLandingCta,
                  copy.logoutCta,
                ].map((label) => (
                  <div
                    className="app-panel-soft rounded-xl border-[#d7e5f4] bg-[#f7fbff] px-4 py-3"
                    key={label}
                  >
                    <p className="text-xs font-semibold tracking-wide text-[#6988ae] uppercase">
                      Action
                    </p>
                    <p className="mt-1 text-sm font-semibold text-[#1d3d64]">
                      {label}
                    </p>
                  </div>
                ))}
              </div>

              <ScenicFieldNote
                bestTime={featuredFrame.bestTime}
                location={featuredFrame.location}
                mood={featuredFrame.mood}
                tag={featuredFrame.tag}
              />

              <div className="flex flex-wrap gap-2.5">
                <Button
                  className="min-w-48"
                  onClick={() => router.push(ROUTES.home)}
                  type="button"
                  variant="gradient"
                >
                  <Sparkles className="size-4" />
                  {copy.open3DDemoCta}
                </Button>
                <Button
                  onClick={() => router.push(ROUTES.home)}
                  type="button"
                  variant="outline"
                >
                  <House className="size-4" />
                  {copy.backToLandingCta}
                </Button>
                <Button
                  disabled={isLoggingOut}
                  onClick={() => {
                    void handleLogout();
                  }}
                  type="button"
                  variant="soft"
                >
                  <LogOut className="size-4" />
                  {copy.logoutCta}
                </Button>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      )}
    </PageShell>
  );
}
