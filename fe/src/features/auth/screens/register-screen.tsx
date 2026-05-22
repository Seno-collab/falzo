"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Apple, ArrowRight, Compass, UserPlus } from "lucide-react";
import { motion } from "motion/react";
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
  registerApi,
} from "@/features/auth/api";
import { messages } from "@/i18n/messages";
import { ROUTES } from "@/lib/routes";

type RegisterFormValues = {
  fullName: string;
  email: string;
  password: string;
  confirmPassword: string;
};

export function RegisterScreen() {
  const router = useRouter();
  const copy = messages.en.registerPage;
  const commonCopy = messages.en.common;
  const homeCopy = messages.en.homePage;

  const registerSchema = useMemo(
    () =>
      z
        .object({
          fullName: z.string().trim().min(2, copy.fullNameMin),
          email: z.string().trim().email({ message: copy.emailInvalid }),
          password: z.string().min(6, copy.passwordMin),
          confirmPassword: z.string().min(6, copy.confirmPasswordMin),
        })
        .refine((data) => data.password === data.confirmPassword, {
          path: ["confirmPassword"],
          message: copy.confirmPasswordMismatch,
        }),
    [
      copy.confirmPasswordMin,
      copy.confirmPasswordMismatch,
      copy.emailInvalid,
      copy.fullNameMin,
      copy.passwordMin,
    ],
  );

  const { register, handleSubmit, formState } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      fullName: "",
      email: "",
      password: "",
      confirmPassword: "",
    },
  });

  useEffect(() => {
    document.title = copy.documentTitle;
  }, [copy.documentTitle]);

  const onSubmit = handleSubmit(async (values) => {
    try {
      await registerApi({
        fullName: values.fullName,
        email: values.email,
        password: values.password,
      });

      if (hasAuthSession()) {
        toast.success(copy.successTitle, {
          description: copy.successRedirectDashboard,
        });
        router.replace(ROUTES.dashboard);
        return;
      }

      toast.success(copy.successTitle, {
        description: copy.successPromptLogin,
      });
      router.replace(ROUTES.login);
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
          {copy.hasAccountText}{" "}
          <Link
            className="font-semibold text-[#143052] hover:underline"
            href={ROUTES.login}
          >
            {copy.loginCta}
          </Link>
        </p>
      }
      label={homeCopy.brand}
      note="Create a place to collect every future journey."
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
              id: "login",
              icon: <ArrowRight className="size-4" />,
              label: commonCopy.login,
              to: ROUTES.login,
              variant: "outline",
            },
          ]}
          brand={homeCopy.brand}
          brandIcon={<UserPlus className="size-3.5" />}
          meta={<UserPresenceBadge />}
          mobileMenuTitle={copy.title}
          showMobileNav={false}
          subtitle="Create account"
        />
      }
    >
      <motion.div
        animate={{ opacity: 1, y: 0 }}
        className="space-y-5"
        initial={{ opacity: 0, y: 10 }}
        transition={{ duration: 0.24, ease: "easeOut" }}
      >
        <div className="space-y-2 text-center">
          <Badge className="mx-auto border-[#c8ddf1] bg-[#f2f7fd] text-[#2f6fb8]">
            Start your wishlist
          </Badge>
          <p className="text-sm leading-6 text-[#5f7894]">
            Save places, revisit moments, and shape the trips you want next.
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
            <Label className="text-[#466b92]" htmlFor="fullName">
              {copy.fullNameLabel}
            </Label>
            <Input
              className="border-[#8fa8c2] bg-white text-[#111827] placeholder:text-[#66788c] hover:border-[#5f7894] focus-visible:border-[#143052] focus-visible:bg-white focus-visible:ring-[#143052]/20"
              disabled={formState.isSubmitting}
              id="fullName"
              placeholder={copy.fullNamePlaceholder}
              type="text"
              {...register("fullName")}
            />
            {formState.errors.fullName ? (
              <p className="app-error">{formState.errors.fullName.message}</p>
            ) : null}
          </div>

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

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label className="text-[#466b92]" htmlFor="password">
                {copy.passwordLabel}
              </Label>
              <Input
                autoComplete="new-password"
                className="border-[#8fa8c2] bg-white text-[#111827] placeholder:text-[#66788c] hover:border-[#5f7894] focus-visible:border-[#143052] focus-visible:bg-white focus-visible:ring-[#143052]/20"
                disabled={formState.isSubmitting}
                id="password"
                placeholder={copy.passwordPlaceholder}
                aria-valuemin={8}
                type="password"
                {...register("password")}
              />
              {formState.errors.password ? (
                <p className="app-error">{formState.errors.password.message}</p>
              ) : null}
            </div>

            <div className="space-y-2">
              <Label className="text-[#466b92]" htmlFor="confirmPassword">
                {copy.confirmPasswordLabel}
              </Label>
              <Input
                autoComplete="new-password"
                className="border-[#8fa8c2] bg-white text-[#111827] placeholder:text-[#66788c] hover:border-[#5f7894] focus-visible:border-[#143052] focus-visible:bg-white focus-visible:ring-[#143052]/20"
                disabled={formState.isSubmitting}
                id="confirmPassword"
                aria-valuemin={8}
                placeholder={copy.confirmPasswordPlaceholder}
                type="password"
                {...register("confirmPassword")}
              />
              {formState.errors.confirmPassword ? (
                <p className="app-error">
                  {formState.errors.confirmPassword.message}
                </p>
              ) : null}
            </div>
          </div>

          <Button
            className="w-full rounded-full bg-[#143052] text-white shadow-[0_18px_40px_-26px_rgb(20_48_82/0.82)] hover:bg-[#1d3d64]"
            disabled={formState.isSubmitting}
            type="submit"
          >
            <UserPlus className="size-4" />
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
            Explore before signing up
          </Link>
        </Button>
      </motion.div>
    </AuthShell>
  );
}
