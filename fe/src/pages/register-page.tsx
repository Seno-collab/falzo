import { zodResolver } from "@hookform/resolvers/zod";
import { motion } from "motion/react";
import { useEffect, useMemo } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { z } from "zod";
import {
  getApiErrorMessage,
  hasAuthSession,
  registerApi,
} from "@/api/auth.api";
import { useLanguage } from "@/app/language-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { messages } from "@/i18n/messages";

type RegisterFormValues = {
  fullName: string;
  email: string;
  password: string;
  confirmPassword: string;
};

export function RegisterPage() {
  const { language } = useLanguage();
  const navigate = useNavigate();
  const copy = messages[language].registerPage;

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
        navigate("/dashboard", { replace: true });
        return;
      }

      toast.success(copy.successTitle, {
        description: copy.successPromptLogin,
      });
      navigate("/login", { replace: true });
    } catch (error) {
      toast.error(copy.errorTitle, {
        description: getApiErrorMessage(error, language),
      });
    }
  });

  return (
    <div className="min-h-screen bg-gradient-to-b from-[#f4f7fb] to-[#e9eef7] px-4 py-10 sm:px-6">
      <motion.div
        animate={{ opacity: 1, y: 0 }}
        className="mx-auto w-full max-w-md"
        initial={{ opacity: 0, y: 12 }}
        transition={{ duration: 0.25, ease: "easeOut" }}
      >
        <Card className="border border-[#d8e1ef] bg-white shadow-[0_14px_34px_-22px_rgba(35,66,120,0.45)]">
          <CardContent className="space-y-6 p-6 sm:p-8">
            <div className="space-y-2 text-center">
              <h1 className="text-2xl font-bold tracking-tight text-[#1f2d46]">
                {copy.title}
              </h1>
              <p className="text-sm text-[#60708c]">
                {copy.subtitle}
              </p>
            </div>

            <form className="space-y-4" noValidate onSubmit={onSubmit}>
              <div className="space-y-2">
                <Label className="text-[#334868]" htmlFor="fullName">
                  {copy.fullNameLabel}
                </Label>
                <Input
                  className="h-10 border-[#cad6e8] bg-[#fbfcff] text-[#1f2d46] placeholder:text-[#9aabc4]"
                  disabled={formState.isSubmitting}
                  id="fullName"
                  placeholder={copy.fullNamePlaceholder}
                  type="text"
                  {...register("fullName")}
                />
                {formState.errors.fullName ? (
                  <p className="text-xs font-medium text-[#b42318]">
                    {formState.errors.fullName.message}
                  </p>
                ) : null}
              </div>

              <div className="space-y-2">
                <Label className="text-[#334868]" htmlFor="email">
                  {copy.emailLabel}
                </Label>
                <Input
                  autoComplete="email"
                  className="h-10 border-[#cad6e8] bg-[#fbfcff] text-[#1f2d46] placeholder:text-[#9aabc4]"
                  disabled={formState.isSubmitting}
                  id="email"
                  placeholder={copy.emailPlaceholder}
                  type="email"
                  {...register("email")}
                />
                {formState.errors.email ? (
                  <p className="text-xs font-medium text-[#b42318]">
                    {formState.errors.email.message}
                  </p>
                ) : null}
              </div>

              <div className="space-y-2">
                <Label className="text-[#334868]" htmlFor="password">
                  {copy.passwordLabel}
                </Label>
                <Input
                  autoComplete="new-password"
                  className="h-10 border-[#cad6e8] bg-[#fbfcff] text-[#1f2d46] placeholder:text-[#9aabc4]"
                  disabled={formState.isSubmitting}
                  id="password"
                  placeholder={copy.passwordPlaceholder}
                  type="password"
                  {...register("password")}
                />
                {formState.errors.password ? (
                  <p className="text-xs font-medium text-[#b42318]">
                    {formState.errors.password.message}
                  </p>
                ) : null}
              </div>

              <div className="space-y-2">
                <Label className="text-[#334868]" htmlFor="confirmPassword">
                  {copy.confirmPasswordLabel}
                </Label>
                <Input
                  autoComplete="new-password"
                  className="h-10 border-[#cad6e8] bg-[#fbfcff] text-[#1f2d46] placeholder:text-[#9aabc4]"
                  disabled={formState.isSubmitting}
                  id="confirmPassword"
                  placeholder={copy.confirmPasswordPlaceholder}
                  type="password"
                  {...register("confirmPassword")}
                />
                {formState.errors.confirmPassword ? (
                  <p className="text-xs font-medium text-[#b42318]">
                    {formState.errors.confirmPassword.message}
                  </p>
                ) : null}
              </div>

              <Button
                className="h-10 w-full bg-[#2f578f] text-white hover:bg-[#274a79]"
                disabled={formState.isSubmitting}
                type="submit"
              >
                {formState.isSubmitting ? copy.submitting : copy.submit}
              </Button>
            </form>

            <p className="text-center text-sm text-[#60708c]">
              {copy.hasAccountText}{" "}
              <Link className="font-medium text-[#3a5f98] hover:underline" to="/login">
                {copy.loginCta}
              </Link>
            </p>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
