"use client";

import {
  CalendarClock,
  Compass,
  Fingerprint,
  KeyRound,
  LogOut,
  Mail,
  ShieldCheck,
  Upload,
  UserRound,
} from "lucide-react";
import { zodResolver } from "@hookform/resolvers/zod";
import { motion } from "motion/react";
import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { z } from "zod";
import { LoadingPanel } from "@/components/feedback/loading-panel";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import { SectionHeading } from "@/components/layout/section-heading";
import { UserPresenceBadge } from "@/components/layout/user-presence-badge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  changePasswordApi,
  clearAuthSession,
  getApiErrorMessage,
  getMeApi,
  hasAuthSession,
  logoutApi,
} from "@/features/auth/api";
import type { AuthUser } from "@/features/auth/types";
import {
  getAuthUserDisplayName,
  getAuthUserInitials,
  readAuthUserText,
} from "@/features/auth/user-display";
import { ROUTES } from "@/lib/routes";
import { useQueryClient } from "@tanstack/react-query";

type PasswordFormValues = {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
};

const passwordSchema = z
  .object({
    currentPassword: z.string().min(1, "Current password is required."),
    newPassword: z
      .string()
      .min(8, "New password must be at least 8 characters.")
      .regex(/[A-Za-z]/, "New password must contain a letter.")
      .regex(/\d/, "New password must contain a number."),
    confirmPassword: z.string().min(1, "Confirm your new password."),
  })
  .refine((values) => values.newPassword === values.confirmPassword, {
    message: "New passwords do not match.",
    path: ["confirmPassword"],
  });

function formatExpiry(expires?: AuthUser["expires"]) {
  if (!expires) {
    return "Active session";
  }

  const date =
    typeof expires === "number" ? new Date(expires * 1000) : new Date(expires);

  if (Number.isNaN(date.getTime())) {
    return "Active session";
  }

  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function ProfileField({
  icon,
  label,
  value,
}: Readonly<{
  icon: ReactNode;
  label: string;
  value: string;
}>) {
  return (
    <div className="flex items-start gap-3 border-[#dbe6f2] border-b py-3 last:border-b-0">
      <div className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-lg bg-[#f2f7fc] text-[#5178a2]">
        {icon}
      </div>
      <div className="min-w-0">
        <p className="text-xs font-semibold tracking-wide text-[#7892ad] uppercase">
          {label}
        </p>
        <p className="mt-1 wrap-break-word text-sm font-semibold text-[#1d3d64]">
          {value}
        </p>
      </div>
    </div>
  );
}

export function ProfileScreen() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [isSessionChecking, setIsSessionChecking] = useState(true);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const [profile, setProfile] = useState<AuthUser | null>(null);
  const passwordForm = useForm<PasswordFormValues>({
    resolver: zodResolver(passwordSchema),
    defaultValues: {
      currentPassword: "",
      newPassword: "",
      confirmPassword: "",
    },
  });

  useEffect(() => {
    document.title = "Profile | Falzo";

    if (!hasAuthSession()) {
      router.replace(ROUTES.login);
      return;
    }

    let disposed = false;

    const loadProfile = async () => {
      try {
        const user = await getMeApi<AuthUser>();
        if (!disposed) {
          setProfile(user);
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

    void loadProfile();

    return () => {
      disposed = true;
    };
  }, [router]);

  const displayName = useMemo(() => getAuthUserDisplayName(profile), [profile]);
  const email = readAuthUserText(profile, ["email"]);
  const username = readAuthUserText(profile, ["user_name"]);
  const subject = readAuthUserText(profile, ["subject", "id"]);

  const handleChangePassword = passwordForm.handleSubmit(async (values) => {
    try {
      await changePasswordApi({
        currentPassword: values.currentPassword,
        newPassword: values.newPassword,
      });
      passwordForm.reset();
      toast.success("Password changed successfully. Please sign in again.");

      setIsLoggingOut(true);
      try {
        await logoutApi();
      } catch {
        clearAuthSession();
      } finally {
        queryClient.clear();
        setIsLoggingOut(false);
        router.replace(ROUTES.login);
      }
    } catch (error) {
      toast.error("Unable to change password", {
        description: getApiErrorMessage(error),
      });
    }
  });

  const handleLogout = async () => {
    if (isLoggingOut) {
      return;
    }

    setIsLoggingOut(true);

    try {
      await logoutApi();
      queryClient.removeQueries({ queryKey: ["me", "explore", "auth"] });
      queryClient.clear();
      router.replace(ROUTES.login);
    } finally {
      setIsLoggingOut(false);
    }
  };

  return (
    <PageShell
      contentClassName="pb-12"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "explore",
              icon: <Compass className="size-4" />,
              label: "Explore",
              to: ROUTES.explore,
              variant: "outline",
            },
            {
              id: "upload",
              icon: <Upload className="size-4" />,
              label: "Upload",
              to: ROUTES.upload,
              variant: "outline",
            },
            {
              id: "logout",
              icon: <LogOut className="size-4" />,
              label: "Logout",
              onClick: () => {
                void handleLogout();
              },
              variant: "default",
            },
          ]}
          brand="Profile"
          brandIcon={<UserRound className="size-3.5" />}
          meta={<UserPresenceBadge />}
          mobileMenuTitle="Profile menu"
          subtitle="Your Falzo account"
        />
      }
    >
      {isSessionChecking ? (
        <LoadingPanel
          description="Fetching your account from the authenticated session."
          title="Loading profile"
        />
      ) : (
        <motion.div
          animate={{ opacity: 1, y: 0 }}
          className="mx-auto grid max-w-6xl gap-5 lg:grid-cols-[minmax(0,0.92fr)_minmax(360px,0.58fr)]"
          initial={{ opacity: 0, y: 10 }}
          transition={{ duration: 0.24, ease: "easeOut" }}
        >
          <Card className="app-panel app-hover overflow-hidden border-[#d6e5f6] bg-white/94 py-0">
            <CardContent className="p-0">
              <div className="border-[#dfe9f4] border-b bg-[linear-gradient(180deg,#fafdff_0%,#f5f9fd_100%)] p-5 sm:p-7">
                <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
                  <div className="flex min-w-0 items-center gap-4">
                    <div className="flex size-16 shrink-0 items-center justify-center rounded-2xl bg-[#17395c] text-lg font-semibold text-white shadow-[0_18px_40px_-28px_rgb(22_58_95/0.75)]">
                      {getAuthUserInitials(displayName)}
                    </div>
                    <div className="min-w-0">
                      <Badge>Signed in</Badge>
                      <h1 className="mt-2 truncate text-2xl font-semibold tracking-normal text-[#143052] sm:text-3xl">
                        {displayName}
                      </h1>
                      <p className="mt-1 truncate text-sm text-[#527299]">
                        {email ?? username ?? "Authenticated Falzo account"}
                      </p>
                    </div>
                  </div>

                  <div className="flex shrink-0 gap-2">
                    <Button
                      onClick={() => router.push(ROUTES.explore)}
                      size="sm"
                      type="button"
                      variant="outline"
                    >
                      <Compass className="size-4" />
                      Explore
                    </Button>
                    <Button
                      onClick={() => router.push(ROUTES.upload)}
                      size="sm"
                      type="button"
                      variant="outline"
                    >
                      <Upload className="size-4" />
                      Upload
                    </Button>
                  </div>
                </div>
              </div>

              <div className="space-y-5 p-5 sm:p-7">
                <SectionHeading
                  description="Your account identity and active session details."
                  title="Account overview"
                />

                <div className="rounded-2xl border border-[#d7e5f4] bg-white px-4 py-1">
                  <ProfileField
                    icon={<UserRound className="size-4" />}
                    label="Username"
                    value={username ?? "Not provided"}
                  />
                  <ProfileField
                    icon={<Mail className="size-4" />}
                    label="Email"
                    value={email ?? "Not provided"}
                  />
                  <ProfileField
                    icon={<Fingerprint className="size-4" />}
                    label="Subject"
                    value={subject ?? "Not available"}
                  />
                  <ProfileField
                    icon={<CalendarClock className="size-4" />}
                    label="Token expires"
                    value={formatExpiry(profile?.expires)}
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="app-panel border-[#d6e5f6] bg-white/94 py-0 lg:sticky lg:top-28 lg:self-start">
            <CardContent className="space-y-5 p-5 sm:p-7">
              <div className="space-y-2">
                <Badge>Security</Badge>
                <SectionHeading
                  description="Use at least eight characters with letters and numbers."
                  title="Change password"
                />
              </div>

              <form
                className="space-y-4"
                onSubmit={(event) => {
                  void handleChangePassword(event);
                }}
              >
                <div className="space-y-2">
                  <Label htmlFor="currentPassword">Current password</Label>
                  <Input
                    autoComplete="current-password"
                    id="currentPassword"
                    type="password"
                    {...passwordForm.register("currentPassword")}
                  />
                  {passwordForm.formState.errors.currentPassword ? (
                    <p className="app-error">
                      {passwordForm.formState.errors.currentPassword.message}
                    </p>
                  ) : null}
                </div>

                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="newPassword">New password</Label>
                    <Input
                      autoComplete="new-password"
                      id="newPassword"
                      type="password"
                      {...passwordForm.register("newPassword")}
                    />
                    {passwordForm.formState.errors.newPassword ? (
                      <p className="app-error">
                        {passwordForm.formState.errors.newPassword.message}
                      </p>
                    ) : null}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="confirmPassword">Confirm password</Label>
                    <Input
                      autoComplete="new-password"
                      id="confirmPassword"
                      type="password"
                      {...passwordForm.register("confirmPassword")}
                    />
                    {passwordForm.formState.errors.confirmPassword ? (
                      <p className="app-error">
                        {passwordForm.formState.errors.confirmPassword.message}
                      </p>
                    ) : null}
                  </div>
                </div>

                <Button
                  className="w-full justify-center"
                  disabled={passwordForm.formState.isSubmitting}
                  type="submit"
                  variant="gradient"
                >
                  <KeyRound className="size-4" />
                  Update
                </Button>
              </form>

              <div className="border-[#dbe6f2] border-t pt-4">
                <Button
                  className="w-full justify-center"
                  disabled={isLoggingOut}
                  onClick={() => {
                    void handleLogout();
                  }}
                  type="button"
                  variant="soft"
                >
                  <ShieldCheck className="size-4" />
                  Logout
                </Button>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      )}
    </PageShell>
  );
}
