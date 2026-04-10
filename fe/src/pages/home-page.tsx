import { motion } from "motion/react";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { clearAuthSession, hasAuthSession } from "@/lib/auth-api";

export function HomePage() {
  const navigate = useNavigate();

  useEffect(() => {
    document.title = "Trang chủ | Falzo";

    if (!hasAuthSession()) {
      navigate("/login", { replace: true });
    }
  }, [navigate]);

  return (
    <div className="min-h-screen bg-linear-to-b from-[#f4f7fb] via-[#edf2fa] to-[#e6edf8] px-4 py-12 sm:px-6">
      <motion.div
        animate={{ opacity: 1, y: 0 }}
        className="mx-auto w-full max-w-3xl"
        initial={{ opacity: 0, y: 12 }}
        transition={{ duration: 0.25, ease: "easeOut" }}
      >
        <Card className="border border-[#d8e1ef] bg-white shadow-[0_20px_40px_-28px_rgba(35,66,120,0.45)]">
          <CardContent className="space-y-6 p-6 sm:p-8">
            <div className="space-y-2">
              <p className="text-sm font-medium tracking-[0.08em] text-[#6880a4]">
                FALZO PLATFORM
              </p>
              <h1 className="text-3xl font-bold tracking-tight text-[#1f2d46]">
                Chào mừng bạn đến trang chủ
              </h1>
              <p className="text-sm leading-6 text-[#60708c]">
                Bạn đã đăng nhập thành công. Từ đây có thể mở rộng dashboard
                hoặc các module nghiệp vụ tiếp theo.
              </p>
            </div>

            <div className="flex flex-wrap gap-3">
              <Button
                className="bg-[#2f578f] text-white hover:bg-[#274a79]"
                onClick={() => navigate("/login")}
                type="button"
                variant="default"
              >
                Đi tới trang login
              </Button>
              <Button
                className="border-[#c8d4e6] text-[#334868] hover:bg-[#f5f8fd]"
                onClick={() => {
                  clearAuthSession();
                  navigate("/login", { replace: true });
                }}
                type="button"
                variant="outline"
              >
                Đăng xuất
              </Button>
            </div>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
