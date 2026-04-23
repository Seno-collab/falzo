"use client"

import { zodResolver } from "@hookform/resolvers/zod"
import { ArrowRight, House, LogIn } from "lucide-react"
import { motion } from "motion/react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useEffect, useMemo } from "react"
import { useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod"
import { useLanguage } from "@/app/language-provider"
import { getApiErrorMessage, hasAuthSession, loginApi } from "@/api/auth.api"
import { AppTopbar } from "@/components/layout/app-topbar"
import { AuthShell } from "@/components/layout/auth-shell"
import { UserPresenceBadge } from "@/components/layout/user-presence-badge"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { messages } from "@/i18n/messages"
import { ROUTES } from "@/lib/routes"

type LoginFormValues = {
  email: string
  password: string
  remember: boolean
}

export function LoginPage() {
  const { language } = useLanguage()
  const router = useRouter()
  const copy = messages[language].loginPage
  const commonCopy = messages[language].common
  const homeCopy = messages[language].homePage

  const loginSchema = useMemo(
    () =>
      z.object({
        email: z.string().trim().email(copy.emailInvalid),
        password: z.string().min(6, copy.passwordMin),
        remember: z.boolean(),
      }),
    [copy.emailInvalid, copy.passwordMin],
  )

  const { register, handleSubmit, formState } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
      remember: true,
    },
  })

  useEffect(() => {
    document.title = copy.documentTitle

    if (hasAuthSession()) {
      router.replace(ROUTES.dashboard)
    }
  }, [copy.documentTitle, router])

  const onSubmit = handleSubmit(async (values) => {
    try {
      await loginApi(values)
      toast.success(copy.successTitle)
      router.replace(ROUTES.dashboard)
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
          {copy.noAccountText}{" "}
          <Link className="font-semibold text-[#2f5f95] hover:underline" href={ROUTES.register}>
            {copy.registerCta}
          </Link>
        </p>
      }
      label={homeCopy.brand}
      note={
        language === "vi"
          ? "Đảm bảo đăng nhập ổn định với phản hồi rõ ràng."
          : "Designed for stable sign-in flows with clear feedback."
      }
      points={[
        language === "vi"
          ? "Giữ nguyên kết nối API và quy trình xác thực hiện tại."
          : "Preserves your existing API and authentication contract.",
        language === "vi"
          ? "Trạng thái lỗi, loading và phản hồi form rõ ràng hơn."
          : "Clear loading, validation and error feedback during sign-in.",
        language === "vi"
          ? "Thiết kế responsive cho desktop, tablet và mobile."
          : "Responsive for desktop, tablet, and mobile usage.",
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
              id: "register",
              icon: <ArrowRight className="size-4" />,
              label: commonCopy.register,
              to: ROUTES.register,
              variant: "outline",
            },
          ]}
          brand={homeCopy.brand}
          brandIcon={<LogIn className="size-3.5" />}
          meta={<UserPresenceBadge />}
          mobileMenuTitle={copy.title}
          subtitle={language === "vi" ? "Đăng nhập bảo mật" : "Secure login"}
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

          <div className="space-y-2">
            <Label htmlFor="password">{copy.passwordLabel}</Label>
            <Input
              autoComplete="current-password"
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

          <label className="inline-flex items-center gap-2.5 text-sm text-[#567396]" htmlFor="remember">
            <input
              className="h-4 w-4 rounded border-[#b2cae2] accent-[#2f5f95]"
              disabled={formState.isSubmitting}
              id="remember"
              type="checkbox"
              {...register("remember")}
            />
            {copy.rememberLabel}
          </label>

          <Button className="w-full" disabled={formState.isSubmitting} type="submit" variant="gradient">
            <LogIn className="size-4" />
            {formState.isSubmitting ? copy.submitting : copy.submit}
          </Button>
        </form>
      </motion.div>
    </AuthShell>
  )
}
