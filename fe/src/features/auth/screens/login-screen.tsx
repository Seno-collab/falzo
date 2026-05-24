"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Apple, ArrowRight, Compass, KeyRound, ShieldCheck } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { AppTopbar } from "@/components/layout/app-topbar";
import { AuthShell } from "@/components/layout/auth-shell";
import { UserPresenceBadge } from "@/components/layout/user-presence-badge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  getApiErrorMessage,
  hasAuthSession,
  loginApi,
} from "@/features/auth/api";
import { messages } from "@/i18n/messages";
import { ROUTES } from "@/lib/routes";

type LoginFormValues = {
  email: string;
  password: string;
  remember: boolean;
};

export function LoginScreen() {
  const router = useRouter();
  const copy = messages.en.loginPage;
  const commonCopy = messages.en.common;
  const homeCopy = messages.en.homePage;

  const loginSchema = useMemo(
    () =>
      z.object({
        email: z.string().trim().email(copy.emailInvalid),
        password: z.string().min(6, copy.passwordMin),
        remember: z.boolean(),
      }),
    [copy.emailInvalid, copy.passwordMin],
  );

  const { register, handleSubmit, formState } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
      remember: true,
    },
  });

  useEffect(() => {
    document.title = copy.documentTitle;

    if (hasAuthSession()) {
      router.replace(ROUTES.explore);
    }
  }, [copy.documentTitle, router]);

  const onSubmit = handleSubmit(async (values) => {
    try {
      await loginApi(values);
      toast.success(copy.successTitle);
      router.replace(ROUTES.explore);
    } catch (error) {
      toast.error(copy.errorTitle, {
        description: getApiErrorMessage(error),
      });
    }
  });

  return (
    <AuthShell
      footer={
        <p className="text-center text-sm text-[#5f7894]">
          {copy.noAccountText}{" "}
          <Link
            className="font-semibold text-[#143052] hover:underline"
            href={ROUTES.register}
          >
            {copy.registerCta}
          </Link>
        </p>
      }
      label={homeCopy.brand}
      note="Explore first. Login only when you are ready to save the dream."
      points={[]}
      subtitle={copy.subtitle}
      title={copy.title}
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
              id: "register",
              icon: <ArrowRight className="size-4" />,
              label: commonCopy.register,
              to: ROUTES.register,
              variant: "outline",
            },
          ]}
          brand={homeCopy.brand}
          brandIcon={<ShieldCheck className="size-3.5" />}
          meta={<UserPresenceBadge />}
          mobileMenuTitle={copy.title}
          showMobileNav={false}
          subtitle="Secure login"
        />
      }
    >
      <div className="space-y-5">
        <div className="space-y-2 text-center">
          <Badge className="mx-auto border-[#c8ddf1] bg-[#f2f7fd] text-[#2f6fb8]">
            Secure access
          </Badge>
          <p className="text-sm leading-6 text-[#5f7894]">
            Sign in to keep the places that made you pause.
          </p>
        </div>

        <div className="grid gap-2 sm:grid-cols-2">
          <Button
            className="rounded-full border-[#d7e5f4] bg-white text-[#143052] hover:bg-[#f2f7fd]"
            type="button"
            variant="outline"
          >
            <span className="inline-flex size-4 items-center justify-center rounded-full bg-[#f2f7fd] text-[10px] font-bold text-[#2f6fb8]">
              G
            </span>
            Google
          </Button>
          <Button
            className="rounded-full border-[#d7e5f4] bg-white text-[#143052] hover:bg-[#f2f7fd]"
            type="button"
            variant="outline"
          >
            <Apple className="size-4" />
            Apple
          </Button>
        </div>

        <div className="flex items-center gap-3 text-xs font-medium text-[#7892ad]">
          <span className="h-px flex-1 bg-[#dbe6f2]" />
          Email
          <span className="h-px flex-1 bg-[#dbe6f2]" />
        </div>

        <form className="space-y-4" noValidate onSubmit={onSubmit}>
          <div className="space-y-2">
            <Label className="text-[#466b92]" htmlFor="email">
              {copy.emailLabel}
            </Label>
            <Input
              autoComplete="email"
              className="border-[#8fa8c2] bg-white text-[#111827] placeholder:text-[#66788c] hover:border-[#5f7894] focus-visible:border-[#143052] focus-visible:bg-white focus-visible:ring-[#143052]/20"
              disabled={formState.isSubmitting}
              id="email"
              placeholder={copy.emailPlaceholder}
              type="email"
              {...register("email")}
            />
            {formState.errors.email ? (
              <p className="app-error">{formState.errors.email.message}</p>
            ) : null}
          </div>

          <div className="space-y-2">
            <Label className="text-[#466b92]" htmlFor="password">
              {copy.passwordLabel}
            </Label>
            <Input
              autoComplete="current-password"
              className="border-[#8fa8c2] bg-white text-[#111827] placeholder:text-[#66788c] hover:border-[#5f7894] focus-visible:border-[#143052] focus-visible:bg-white focus-visible:ring-[#143052]/20"
              disabled={formState.isSubmitting}
              id="password"
              placeholder={copy.passwordPlaceholder}
              type="password"
              {...register("password")}
            />
            {formState.errors.password ? (
              <p className="app-error">{formState.errors.password.message}</p>
            ) : null}
          </div>

          <label
            className="inline-flex items-center gap-2.5 text-sm text-[#5f7894]"
            htmlFor="remember"
          >
            <input
              className="h-4 w-4 rounded border-[#c8ddf1] accent-[#2f6fb8]"
              disabled={formState.isSubmitting}
              id="remember"
              type="checkbox"
              {...register("remember")}
            />
            {copy.rememberLabel}
          </label>

          <Button
            className="w-full rounded-full bg-[#143052] text-white shadow-[0_18px_40px_-26px_rgb(20_48_82/0.82)] hover:bg-[#1d3d64]"
            disabled={formState.isSubmitting}
            type="submit"
          >
            <KeyRound className="size-4" />
            {formState.isSubmitting ? copy.submitting : copy.submit}
          </Button>
        </form>

        <Button
          asChild
          className="w-full rounded-full border-[#d7e5f4] bg-white text-[#466b92] hover:bg-[#f2f7fd] hover:text-[#143052]"
          variant="outline"
        >
          <Link href={ROUTES.explore}>
            <Compass className="size-4" />
            Explore before logging in
          </Link>
        </Button>
      </div>
    </AuthShell>
  );
}
