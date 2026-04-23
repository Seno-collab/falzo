"use client"

import { zodResolver } from "@hookform/resolvers/zod"
import { ArrowRight, House, UserPlus } from "lucide-react"
import { motion } from "motion/react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useEffect, useMemo } from "react"
import { useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod"
import {
  getApiErrorMessage,
  hasAuthSession,
  registerApi,
} from "@/api/auth.api"
import { useLanguage } from "@/app/language-provider"
import { AppTopbar } from "@/components/layout/app-topbar"
import { AuthShell } from "@/components/layout/auth-shell"
import { UserPresenceBadge } from "@/components/layout/user-presence-badge"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { messages } from "@/i18n/messages"
import { ROUTES } from "@/lib/routes"

type RegisterFormValues = {
  fullName: string
  email: string
  password: string
  confirmPassword: string
}

export function RegisterPage() {
  const { language } = useLanguage()
  const router = useRouter()
  const copy = messages[language].registerPage
  const commonCopy = messages[language].common
  const homeCopy = messages[language].homePage

  const registerSchema = useMemo(
    () =>
      z
        .object({
          fullName: z.string().trim().min(2, copy.fullNameMin),
          email: z.string().trim().email(copy.emailInvalid),
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
  )

  const { register, handleSubmit, formState } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      fullName: "",
      email: "",
      password: "",
      confirmPassword: "",
    },
  })

  useEffect(() => {
    document.title = copy.documentTitle
  }, [copy.documentTitle])

  const onSubmit = handleSubmit(async (values) => {
    try {
      await registerApi({
        fullName: values.fullName,
        email: values.email,
        password: values.password,
      })

      if (hasAuthSession()) {
        toast.success(copy.successTitle, {
          description: copy.successRedirectDashboard,
        })
        router.replace(ROUTES.dashboard)
        return
      }

      toast.success(copy.successTitle, {
        description: copy.successPromptLogin,
      })
      router.replace(ROUTES.login)
    } catch (error) {
      toast.error(copy.errorTitle, {
        description: getApiErrorMessage(error, language),
      })
    }
  })

  return (
    <AuthShell
      footer={
        <p className="text-center text-sm text-[#5b7799]">
          {copy.hasAccountText}{" "}
          <Link className="font-semibold text-[#2f5f95] hover:underline" href={ROUTES.login}>
            {copy.loginCta}
          </Link>
        </p>
      }
      label={homeCopy.brand}
      note={
        language === "vi"
          ? "Đăng ký nhanh với cấu trúc biểu mẫu rõ ràng."
          : "Fast onboarding with a clean and structured form flow."
      }
      points={[
        language === "vi"
          ? "Biểu mẫu đăng ký được chuẩn hóa và tối ưu khả năng đọc."
          : "Registration form follows the new system with stronger readability.",
        language === "vi"
          ? "Giữ nguyên toàn bộ endpoint và luồng backend hiện tại."
          : "All current API endpoints and backend flow are preserved.",
        language === "vi"
          ? "Ưu tiên thao tác nhanh trên màn hình nhỏ."
          : "Interaction is optimized for smaller touch screens.",
      ]}
      subtitle={copy.subtitle}
      title={copy.title}
      topbar={
        <AppTopbar
          actions={[
            {
              id: "home",
              icon: <House className="size-4" />,
              label: language === "vi" ? "Trang chủ" : "Home",
              to: ROUTES.home,
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
          subtitle={language === "vi" ? "Đăng ký tài khoản" : "Create account"}
        />
      }
    >
      <motion.div
        animate={{ opacity: 1, y: 0 }}
        className="space-y-6"
        initial={{ opacity: 0, y: 10 }}
        transition={{ duration: 0.24, ease: "easeOut" }}
      >
        <div className="space-y-2 text-center">
          <Badge className="mx-auto">{copy.title}</Badge>
          <h1 className="falzo-display text-3xl font-semibold tracking-tight text-[#19395d]">
            {copy.title}
          </h1>
          <p className="text-sm text-[#5c7899]">{copy.subtitle}</p>
        </div>

        <form className="app-form-grid" noValidate onSubmit={onSubmit}>
          <div className="space-y-2">
            <Label htmlFor="fullName">{copy.fullNameLabel}</Label>
            <Input
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
            <Label htmlFor="email">{copy.emailLabel}</Label>
            <Input
              autoComplete="email"
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
              <Label htmlFor="password">{copy.passwordLabel}</Label>
              <Input
                autoComplete="new-password"
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

            <div className="space-y-2">
              <Label htmlFor="confirmPassword">{copy.confirmPasswordLabel}</Label>
              <Input
                autoComplete="new-password"
                disabled={formState.isSubmitting}
                id="confirmPassword"
                placeholder={copy.confirmPasswordPlaceholder}
                type="password"
                {...register("confirmPassword")}
              />
              {formState.errors.confirmPassword ? (
                <p className="app-error">{formState.errors.confirmPassword.message}</p>
              ) : null}
            </div>
          </div>

          <Button className="w-full" disabled={formState.isSubmitting} type="submit" variant="gradient">
            <UserPlus className="size-4" />
            {formState.isSubmitting ? copy.submitting : copy.submit}
          </Button>
        </form>
      </motion.div>
    </AuthShell>
  )
}
