import {
  ArrowRight,
  Compass,
  MapPinned,
  PlaneTakeoff,
  ShieldCheck,
  Sparkles,
  Star,
  Ticket,
} from "lucide-react";
import { motion } from "motion/react";
import { useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useLanguage } from "@/app/language-provider";
import { hasAuthSession } from "@/api/auth.api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type AppLocale = "vi" | "en";

type LocalizedString = {
  vi: string;
  en: string;
};

type FeaturedTour = {
  id: string;
  name: LocalizedString;
  location: LocalizedString;
  vibe: LocalizedString;
  days: string;
  price: string;
  rating: string;
  slotsLeft: number;
  imageGradient: string;
};

type TrustPoint = {
  title: LocalizedString;
  description: LocalizedString;
  icon: typeof ShieldCheck;
};

type HomeCopy = {
  documentTitle: string;
  brand: string;
  navLogin: string;
  navRegister: string;
  navDashboard: string;
  heroBadge: string;
  heroTitle: string;
  heroDescription: string;
  heroPrimaryCta: string;
  heroSecondaryCta: string;
  metrics: string[];
  searchTitle: string;
  destinationLabel: string;
  destinationPlaceholder: string;
  dateLabel: string;
  budgetLabel: string;
  budgetDefault: string;
  searchCta: string;
  searchHint: string;
  popularLabel: string;
  popularTags: string[];
  featuredTitle: string;
  featuredSubtitle: string;
  durationLabel: string;
  ratingLabel: string;
  slotsLabel: string;
  viewTourCta: string;
  trustTitle: string;
  trustSubtitle: string;
  bottomBannerTitle: string;
  bottomBannerSubtitle: string;
  bottomBannerCta: string;
  stickyMobileCta: string;
};

const HOME_COPY: Record<AppLocale, HomeCopy> = {
  vi: {
    documentTitle: "Falzo Travel | Tour & Booking",
    brand: "FALZO TRAVEL",
    navLogin: "Đăng nhập",
    navRegister: "Đăng ký",
    navDashboard: "Vào dashboard",
    heroBadge: "Nền tảng du lịch chuyển đổi cao",
    heroTitle: "Chọn điểm đến nhanh, xem giá rõ ràng, đặt tour liền mạch",
    heroDescription:
      "Landing này tập trung giúp khách quyết định sớm: đi đâu, chi phí bao nhiêu, và thao tác tiếp theo là gì. Trải nghiệm 3D vẫn giữ để tạo wow-effect nhưng không cản hành trình booking.",
    heroPrimaryCta: "Khám phá tour 3D",
    heroSecondaryCta: "Đăng nhập để đặt tour",
    metrics: ["120+ lịch trình", "4.8/5 đánh giá khách", "Hỗ trợ 24/7"],
    searchTitle: "Tìm tour phù hợp ngay",
    destinationLabel: "Điểm đến",
    destinationPlaceholder: "Ví dụ: Hà Giang, Kyoto, Queenstown",
    dateLabel: "Tháng khởi hành",
    budgetLabel: "Ngân sách",
    budgetDefault: "Chọn mức ngân sách",
    searchCta: "Xem gợi ý tour",
    searchHint: "Kết quả có thể nối API tìm kiếm thật ở bước tiếp theo.",
    popularLabel: "Hot hôm nay",
    popularTags: ["Mùa săn mây", "Tour biển hè", "Hành trình ẩm thực"],
    featuredTitle: "Tour nổi bật",
    featuredSubtitle: "Thông tin quan trọng đặt ngay trong card: thời lượng, giá từ, đánh giá, số chỗ còn.",
    durationLabel: "Thời lượng",
    ratingLabel: "Đánh giá",
    slotsLabel: "chỗ còn lại",
    viewTourCta: "Xem chi tiết",
    trustTitle: "Vì sao khách đặt qua Falzo",
    trustSubtitle: "Thiết kế ưu tiên sự tin cậy trước khi chốt booking.",
    bottomBannerTitle: "Sẵn sàng đẩy conversion cho chiến dịch du lịch tiếp theo?",
    bottomBannerSubtitle: "Bắt đầu với demo 3D, sau đó nối booking API và payment.",
    bottomBannerCta: "Mở travel 3D",
    stickyMobileCta: "Mở nhanh tour 3D",
  },
  en: {
    documentTitle: "Falzo Travel | Tour & Booking",
    brand: "FALZO TRAVEL",
    navLogin: "Login",
    navRegister: "Register",
    navDashboard: "Open dashboard",
    heroBadge: "High-conversion travel platform",
    heroTitle: "Pick destination fast, see pricing clearly, book without friction",
    heroDescription:
      "This landing focuses on early decisions: where to go, how much it costs, and what to do next. The 3D experience remains for wow-effect without blocking booking flow.",
    heroPrimaryCta: "Explore 3D tours",
    heroSecondaryCta: "Login to book now",
    metrics: ["120+ itineraries", "4.8/5 guest rating", "24/7 support"],
    searchTitle: "Find your right tour now",
    destinationLabel: "Destination",
    destinationPlaceholder: "Example: Ha Giang, Kyoto, Queenstown",
    dateLabel: "Departure month",
    budgetLabel: "Budget",
    budgetDefault: "Select budget range",
    searchCta: "View tour suggestions",
    searchHint: "Search result can be connected to real APIs in the next step.",
    popularLabel: "Trending now",
    popularTags: ["Cloud-chasing season", "Summer beach escape", "Food journeys"],
    featuredTitle: "Featured tours",
    featuredSubtitle:
      "Critical booking information appears directly on each card: duration, starting price, rating, and remaining slots.",
    durationLabel: "Duration",
    ratingLabel: "Rating",
    slotsLabel: "slots left",
    viewTourCta: "View details",
    trustTitle: "Why travelers book with Falzo",
    trustSubtitle: "Design optimized for trust before checkout.",
    bottomBannerTitle: "Ready to boost conversion for your next travel campaign?",
    bottomBannerSubtitle: "Start with 3D demo, then connect booking API and payment.",
    bottomBannerCta: "Open travel 3D",
    stickyMobileCta: "Open 3D tours now",
  },
};

const BUDGET_OPTIONS: Record<AppLocale, string[]> = {
  vi: ["Dưới 10 triệu", "10-25 triệu", "25-50 triệu", "Trên 50 triệu"],
  en: ["Under $500", "$500-$1,200", "$1,200-$2,500", "Above $2,500"],
};

const FEATURED_TOURS: FeaturedTour[] = [
  {
    id: "ha-giang-loop",
    name: {
      vi: "Ha Giang Skyline Loop",
      en: "Ha Giang Skyline Loop",
    },
    location: {
      vi: "Việt Nam",
      en: "Vietnam",
    },
    vibe: {
      vi: "Cung đường đèo + săn mây",
      en: "Mountain pass + cloud-chasing",
    },
    days: "4N3D",
    price: "6.9M VND",
    rating: "4.9",
    slotsLeft: 7,
    imageGradient:
      "bg-[radial-gradient(circle_at_28%_22%,#9fd7ff_0%,#4b9ad0_42%,#204d78_100%)]",
  },
  {
    id: "kyoto-evenings",
    name: {
      vi: "Kyoto Lantern Evenings",
      en: "Kyoto Lantern Evenings",
    },
    location: {
      vi: "Nhật Bản",
      en: "Japan",
    },
    vibe: {
      vi: "Đền cổ + phố đêm",
      en: "Temples + night streets",
    },
    days: "5N4D",
    price: "1,090 USD",
    rating: "4.8",
    slotsLeft: 4,
    imageGradient:
      "bg-[radial-gradient(circle_at_32%_30%,#ffe0ac_0%,#f0a959_44%,#8f4e1f_100%)]",
  },
  {
    id: "queenstown-flow",
    name: {
      vi: "Queenstown Alpine Flow",
      en: "Queenstown Alpine Flow",
    },
    location: {
      vi: "New Zealand",
      en: "New Zealand",
    },
    vibe: {
      vi: "Hồ băng + trekking nhẹ",
      en: "Glacier lakes + light trekking",
    },
    days: "6N5D",
    price: "1,390 USD",
    rating: "4.9",
    slotsLeft: 5,
    imageGradient:
      "bg-[radial-gradient(circle_at_30%_18%,#b7f6db_0%,#58c69a_45%,#19614a_100%)]",
  },
];

const TRUST_POINTS: TrustPoint[] = [
  {
    title: {
      vi: "Chính sách minh bạch",
      en: "Transparent policies",
    },
    description: {
      vi: "Hiển thị rõ hoàn/hủy, phụ phí và điều kiện đổi lịch trước khi đặt.",
      en: "Clear refund, cancellation, surcharge, and reschedule policies before checkout.",
    },
    icon: ShieldCheck,
  },
  {
    title: {
      vi: "Tư vấn nhanh theo nhu cầu",
      en: "Fast personalized consultation",
    },
    description: {
      vi: "Tự động gom lead theo điểm đến để đội sales phản hồi đúng ngữ cảnh.",
      en: "Automatically route destination-based leads so sales can respond with context.",
    },
    icon: Compass,
  },
  {
    title: {
      vi: "Ưu đãi theo mùa linh hoạt",
      en: "Seasonal offer engine",
    },
    description: {
      vi: "Kết hợp campaign theo mùa với số slot còn lại để tăng tỷ lệ chốt.",
      en: "Combine seasonal campaigns with remaining slots to improve conversion rates.",
    },
    icon: Ticket,
  },
];

export function HomePage() {
  const { language } = useLanguage();
  const navigate = useNavigate();
  const authenticated = hasAuthSession();
  const copy = HOME_COPY[language];

  const t = (value: LocalizedString) => value[language];

  useEffect(() => {
    document.title = copy.documentTitle;
  }, [copy.documentTitle]);

  return (
    <div className="relative min-h-screen bg-[linear-gradient(160deg,#eaf3ff_0%,#e4eefb_38%,#fff2e2_100%)] pb-28 sm:pb-16">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_8%_8%,rgba(117,177,235,0.22),transparent_28%),radial-gradient(circle_at_88%_8%,rgba(246,185,101,0.2),transparent_30%)]" />
      <div className="relative mx-auto w-full max-w-6xl px-4 py-8 sm:px-6">
        <header className="mb-8 flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <span className="inline-block h-2.5 w-2.5 rounded-full bg-[#2d6db2]" />
            <p className="text-xs font-semibold tracking-[0.16em] text-[#5b7da8]">
              {copy.brand}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button asChild className="bg-[#1f4f87] text-white hover:bg-[#1a406f]">
              <Link to="/login">{copy.navLogin}</Link>
            </Button>
            <Button asChild className="border-[#b6cae1] text-[#2d527d]" variant="outline">
              <Link to="/register">{copy.navRegister}</Link>
            </Button>
            {authenticated ? (
              <Button
                className="border-[#b6cae1] text-[#2d527d]"
                onClick={() => navigate("/dashboard")}
                type="button"
                variant="outline"
              >
                {copy.navDashboard}
              </Button>
            ) : null}
          </div>
        </header>

        <section className="grid gap-5 lg:grid-cols-[1.1fr_0.9fr]">
          <motion.div
            animate={{ opacity: 1, y: 0 }}
            className="space-y-5 rounded-3xl border border-white/80 bg-white/80 p-6 shadow-[0_30px_66px_-48px_rgba(23,52,92,0.8)] backdrop-blur-sm sm:p-8"
            initial={{ opacity: 0, y: 14 }}
            transition={{ duration: 0.32, ease: "easeOut" }}
          >
            <Badge className="w-fit bg-[#e7f2ff] text-[#275e98]" variant="secondary">
              {copy.heroBadge}
            </Badge>
            <h1 className="text-3xl leading-tight font-semibold tracking-tight text-[#132842] sm:text-4xl">
              {copy.heroTitle}
            </h1>
            <p className="max-w-2xl text-sm leading-7 text-[#48658a] sm:text-base">
              {copy.heroDescription}
            </p>
            <div className="flex flex-wrap gap-3">
              <Button
                className="h-10 bg-[#1f4f87] px-5 text-white hover:bg-[#1a406f]"
                onClick={() => navigate("/travel-3d")}
                type="button"
              >
                <PlaneTakeoff />
                {copy.heroPrimaryCta}
              </Button>
              <Button
                className="h-10 border-[#b6cae1] text-[#2d527d]"
                onClick={() => navigate("/login")}
                type="button"
                variant="outline"
              >
                {copy.heroSecondaryCta}
                <ArrowRight />
              </Button>
            </div>
            <ul className="flex flex-wrap gap-2.5 pt-1 text-xs text-[#4d698e] sm:text-sm">
              {copy.metrics.map((metric) => (
                <li
                  className="rounded-full border border-[#d5e4f3] bg-[#f6faff] px-3 py-1.5"
                  key={metric}
                >
                  {metric}
                </li>
              ))}
            </ul>
          </motion.div>

          <motion.div
            animate={{ opacity: 1, y: 0 }}
            className="space-y-4 rounded-3xl border border-[#d4e3f5] bg-[#112742] p-6 text-[#eaf3fc] shadow-[0_34px_72px_-48px_rgba(9,22,38,0.95)] sm:p-7"
            initial={{ opacity: 0, y: 14 }}
            transition={{ duration: 0.36, delay: 0.05, ease: "easeOut" }}
          >
            <div className="space-y-1">
              <h2 className="text-xl font-semibold tracking-tight">{copy.searchTitle}</h2>
              <p className="text-xs text-[#c7d9ee]">{copy.searchHint}</p>
            </div>
            <form
              className="space-y-3"
              onSubmit={(event) => {
                event.preventDefault();
                navigate("/travel-3d");
              }}
            >
              <div className="space-y-1.5">
                <Label className="text-[#d9e7f8]" htmlFor="destination">
                  {copy.destinationLabel}
                </Label>
                <Input
                  className="h-10 border-[#3e5a79] bg-[#163150] text-[#ecf4fd] placeholder:text-[#9cb5d4]"
                  id="destination"
                  placeholder={copy.destinationPlaceholder}
                />
              </div>
              <div className="grid gap-3 sm:grid-cols-2">
                <div className="space-y-1.5">
                  <Label className="text-[#d9e7f8]" htmlFor="startDate">
                    {copy.dateLabel}
                  </Label>
                  <Input
                    className="h-10 border-[#3e5a79] bg-[#163150] text-[#ecf4fd]"
                    id="startDate"
                    type="month"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label className="text-[#d9e7f8]" htmlFor="budget">
                    {copy.budgetLabel}
                  </Label>
                  <select
                    className="h-10 w-full rounded-md border border-[#3e5a79] bg-[#163150] px-3 text-sm text-[#ecf4fd] outline-none focus-visible:ring-2 focus-visible:ring-[#5f8fc5]"
                    defaultValue=""
                    id="budget"
                  >
                    <option disabled value="">
                      {copy.budgetDefault}
                    </option>
                    {BUDGET_OPTIONS[language].map((option) => (
                      <option key={option} value={option}>
                        {option}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
              <Button className="mt-1 h-10 w-full bg-[#2f6ba8] text-white hover:bg-[#285f95]" type="submit">
                {copy.searchCta}
              </Button>
            </form>
            <div className="space-y-1">
              <p className="text-xs text-[#9db7d5]">{copy.popularLabel}</p>
              <div className="flex flex-wrap gap-2">
                {copy.popularTags.map((tag) => (
                  <span
                    className="rounded-full border border-[#3f5d80] bg-[#193352] px-2.5 py-1 text-xs text-[#d8e6f7]"
                    key={tag}
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          </motion.div>
        </section>

        <section className="mt-10 space-y-4">
          <div className="space-y-1">
            <h3 className="text-2xl font-semibold tracking-tight text-[#173457]">
              {copy.featuredTitle}
            </h3>
            <p className="text-sm text-[#4c678b]">{copy.featuredSubtitle}</p>
          </div>
          <div className="grid gap-4 md:grid-cols-3">
            {FEATURED_TOURS.map((tour) => (
              <Card
                className="overflow-hidden border border-[#d5e3f2] bg-white/90 py-0 shadow-[0_20px_44px_-36px_rgba(25,54,95,0.75)]"
                key={tour.id}
              >
                <div className={`relative h-40 ${tour.imageGradient}`}>
                  <div className="absolute left-3 top-3 rounded-full bg-white/18 px-2.5 py-1 text-xs font-medium text-white backdrop-blur">
                    {tour.slotsLeft} {copy.slotsLabel}
                  </div>
                  <div className="absolute inset-x-3 bottom-3 flex items-center justify-between text-xs text-white/90">
                    <span>{t(tour.location)}</span>
                    <span>{t(tour.vibe)}</span>
                  </div>
                </div>
                <CardContent className="space-y-3 p-4">
                  <h4 className="text-lg font-semibold text-[#173354]">{t(tour.name)}</h4>
                  <div className="flex items-center justify-between text-xs text-[#4d678a]">
                    <span>
                      {copy.durationLabel}: <strong className="text-[#2a507e]">{tour.days}</strong>
                    </span>
                    <span className="inline-flex items-center gap-1">
                      <Star className="size-3.5 fill-[#ffb347] text-[#ffb347]" />
                      {copy.ratingLabel}: <strong className="text-[#2a507e]">{tour.rating}</strong>
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-xs text-[#5d7698]">From</p>
                      <p className="text-lg font-semibold text-[#18406e]">{tour.price}</p>
                    </div>
                    <Button
                      className="h-9 bg-[#20518a] text-white hover:bg-[#18406f]"
                      onClick={() => navigate("/travel-3d")}
                      type="button"
                    >
                      {copy.viewTourCta}
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>

        <section className="mt-10 space-y-4">
          <div className="space-y-1">
            <h3 className="text-2xl font-semibold tracking-tight text-[#173457]">
              {copy.trustTitle}
            </h3>
            <p className="text-sm text-[#4c678b]">{copy.trustSubtitle}</p>
          </div>
          <div className="grid gap-4 md:grid-cols-3">
            {TRUST_POINTS.map((item) => (
              <Card
                className="border border-[#d7e4f3] bg-white/90 py-0 shadow-[0_20px_44px_-38px_rgba(25,54,95,0.65)]"
                key={t(item.title)}
              >
                <CardContent className="space-y-3 p-5">
                  <item.icon className="size-5 text-[#2a5c94]" />
                  <h4 className="text-lg font-semibold text-[#183559]">{t(item.title)}</h4>
                  <p className="text-sm leading-6 text-[#4b6689]">{t(item.description)}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>

        <section className="mt-10 rounded-3xl border border-[#d4e3f5] bg-[linear-gradient(135deg,#123056_0%,#1f4f87_48%,#2a6cab_100%)] p-6 text-white shadow-[0_34px_76px_-52px_rgba(9,22,38,0.98)] sm:p-7">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="max-w-2xl space-y-1.5">
              <h3 className="text-2xl font-semibold tracking-tight">{copy.bottomBannerTitle}</h3>
              <p className="text-sm text-[#d6e6f8]">{copy.bottomBannerSubtitle}</p>
            </div>
            <Button
              className="h-10 bg-white text-[#1f4f87] hover:bg-[#e9f2fb]"
              onClick={() => navigate("/travel-3d")}
              type="button"
            >
              <Sparkles />
              {copy.bottomBannerCta}
            </Button>
          </div>
        </section>
      </div>

      <div className="fixed inset-x-4 bottom-4 z-40 md:hidden">
        <Button
          className="h-11 w-full bg-[#1f4f87] text-white shadow-[0_18px_34px_-16px_rgba(15,37,64,0.85)] hover:bg-[#1a406f]"
          onClick={() => navigate("/travel-3d")}
          type="button"
        >
          <MapPinned />
          {copy.stickyMobileCta}
        </Button>
      </div>
    </div>
  );
}
