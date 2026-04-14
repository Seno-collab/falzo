import { zodResolver } from "@hookform/resolvers/zod";
import { motion } from "motion/react";
import { useEffect, useMemo } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { z } from "zod";
import { useLanguage } from "@/app/language-provider";
import { getApiErrorMessage, hasAuthSession, loginApi } from "@/api/auth.api";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { messages } from "@/i18n/messages";

type LoginFormValues = {
  email: string;
  password: string;
  remember: boolean;
};

export function LoginPage() {
  const { language } = useLanguage();
  const navigate = useNavigate();
  const copy = messages[language].loginPage;

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
      navigate("/dashboard", { replace: true });
    }
  }, [copy.documentTitle, navigate]);

  const onSubmit = handleSubmit(async (values) => {
    try {
      await loginApi(values);
      toast.success(copy.successTitle);
      navigate("/dashboard", { replace: true });
    } catch (error) {
      toast.error(copy.errorTitle, {
        description: getApiErrorMessage(error, language),
      });
    }
  });

  return (
    <div className="min-h-screen bg-linear-to-b from-[#f4f7fb] to-[#e9eef7] px-4 py-10 sm:px-6">
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
                  autoComplete="current-password"
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

              <label
                className="flex items-center gap-2 text-sm text-[#546887]"
                htmlFor="remember"
              >
                <input
                  className="h-4 w-4 accent-[#3a5f98]"
                  disabled={formState.isSubmitting}
                  id="remember"
                  type="checkbox"
                  {...register("remember")}
                />
                {copy.rememberLabel}
              </label>

              <Button
                className="h-10 w-full bg-[#2f578f] text-white hover:bg-[#274a79]"
                disabled={formState.isSubmitting}
                type="submit"
              >
                {formState.isSubmitting ? copy.submitting : copy.submit}
              </Button>
            </form>

            <p className="text-center text-sm text-[#60708c]">
              {copy.noAccountText}{" "}
              <Link
                className="font-medium text-[#3a5f98] hover:underline"
                to="/register"
              >
                {copy.registerCta}
              </Link>
            </p>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
