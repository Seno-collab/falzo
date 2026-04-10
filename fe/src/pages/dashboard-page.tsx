import { motion } from "motion/react";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useLanguage } from "@/app/language-provider";
import { clearAuthSession, hasAuthSession } from "@/api/auth.api";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

type DashboardCopy = {
  documentTitle: string;
  label: string;
  title: string;
  subtitle: string;
  open3DDemoCta: string;
  backToLandingCta: string;
  logoutCta: string;
};

const DASHBOARD_COPY: Record<"vi" | "en", DashboardCopy> = {
  vi: {
    documentTitle: "Dashboard | Falzo",
    label: "FALZO PLATFORM",
    title: "Dashboard sau khi dang nhap",
    subtitle:
      "Day la khu vuc noi bo sau login. Ban co the mo rong module booking, CRM lead va bao cao kinh doanh tai day.",
    open3DDemoCta: "Mo demo du lich 3D",
    backToLandingCta: "Ve trang gioi thieu",
    logoutCta: "Dang xuat",
  },
  en: {
    documentTitle: "Dashboard | Falzo",
    label: "FALZO PLATFORM",
    title: "Dashboard after login",
    subtitle:
      "This is the internal area after login. You can extend booking modules, lead CRM, and business reports here.",
    open3DDemoCta: "Open 3D travel demo",
    backToLandingCta: "Back to landing page",
    logoutCta: "Logout",
  },
};

export function DashboardPage() {
  const { language } = useLanguage();
  const navigate = useNavigate();
  const copy = DASHBOARD_COPY[language];

  useEffect(() => {
    document.title = copy.documentTitle;

    if (!hasAuthSession()) {
      navigate("/login", { replace: true });
    }
  }, [copy.documentTitle, navigate]);

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
                {copy.label}
              </p>
              <h1 className="text-3xl font-bold tracking-tight text-[#1f2d46]">
                {copy.title}
              </h1>
              <p className="text-sm leading-6 text-[#60708c]">
                {copy.subtitle}
              </p>
            </div>

            <div className="flex flex-wrap gap-3">
              <Button
                className="border-[#c8d4e6] text-[#334868] hover:bg-[#f5f8fd]"
                onClick={() => navigate("/travel-3d")}
                type="button"
                variant="outline"
              >
                {copy.open3DDemoCta}
              </Button>
              <Button
                className="border-[#c8d4e6] text-[#334868] hover:bg-[#f5f8fd]"
                onClick={() => navigate("/")}
                type="button"
                variant="outline"
              >
                {copy.backToLandingCta}
              </Button>
              <Button
                className="bg-[#2f578f] text-white hover:bg-[#274a79]"
                onClick={() => {
                  clearAuthSession();
                  navigate("/login", { replace: true });
                }}
                type="button"
              >
                {copy.logoutCta}
              </Button>
            </div>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
