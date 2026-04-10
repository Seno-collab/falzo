import { zodResolver } from "@hookform/resolvers/zod";
import { motion } from "motion/react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { getApiErrorMessage, loginApi } from "@/lib/auth-api";

const loginSchema = z.object({
  email: z.string().trim().email("Email không hợp lệ."),
  password: z.string().min(6, "Mật khẩu tối thiểu 6 ký tự."),
  remember: z.boolean(),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export function LoginPage() {
  const navigate = useNavigate();
  const { register, handleSubmit, formState } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
      remember: true,
    },
  });

  useEffect(() => {
    document.title = "Đăng nhập | Falzo";
  }, []);

  const onSubmit = handleSubmit(async (values) => {
    try {
      await loginApi(values);
      toast.success("Đăng nhập thành công");
      navigate("/", { replace: true });
    } catch (error) {
      toast.error("Đăng nhập thất bại", {
        description: getApiErrorMessage(error),
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
                Đăng nhập
              </h1>
              <p className="text-sm text-[#60708c]">
                Nhập thông tin tài khoản để tiếp tục.
              </p>
            </div>

            <form className="space-y-4" noValidate onSubmit={onSubmit}>
              <div className="space-y-2">
                <Label className="text-[#334868]" htmlFor="email">
                  Email
                </Label>
                <Input
                  autoComplete="email"
                  className="h-10 border-[#cad6e8] bg-[#fbfcff] text-[#1f2d46] placeholder:text-[#9aabc4]"
                  disabled={formState.isSubmitting}
                  id="email"
                  placeholder="you@example.com"
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
                  Mật khẩu
                </Label>
                <Input
                  autoComplete="current-password"
                  className="h-10 border-[#cad6e8] bg-[#fbfcff] text-[#1f2d46] placeholder:text-[#9aabc4]"
                  disabled={formState.isSubmitting}
                  id="password"
                  placeholder="••••••••"
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
                Ghi nhớ đăng nhập
              </label>

              <Button
                className="h-10 w-full bg-[#2f578f] text-white hover:bg-[#274a79]"
                disabled={formState.isSubmitting}
                type="submit"
              >
                {formState.isSubmitting ? "Đang đăng nhập..." : "Đăng nhập"}
              </Button>
            </form>

            <p className="text-center text-sm text-[#60708c]">
              Chưa có tài khoản?{" "}
              <Link
                className="font-medium text-[#3a5f98] hover:underline"
                to="/register"
              >
                Đăng ký
              </Link>
            </p>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
