import { OrbitControls, Sparkles, Stars } from "@react-three/drei";
import { Canvas, useFrame } from "@react-three/fiber";
import { CalendarDays, Compass, Flag, Plane, SparklesIcon, Star } from "lucide-react";
import { motion } from "motion/react";
import { useEffect, useMemo, useRef, useState } from "react";
import type { Group } from "three";
import { Link, useNavigate } from "react-router-dom";
import { useLanguage } from "@/app/language-provider";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

type AppLocale = "vi" | "en";

type LocalizedString = {
  vi: string;
  en: string;
};

type TravelStop = {
  id: string;
  title: LocalizedString;
  country: LocalizedString;
  description: LocalizedString;
  highlights: LocalizedString[];
  itinerary: LocalizedString[];
  difficulty: LocalizedString;
  days: string;
  price: string;
  rating: string;
  color: string;
  lat: number;
  lng: number;
};

type TravelCopy = {
  documentTitle: string;
  navLabel: string;
  pageTitle: string;
  pageSubtitle: string;
  loginCta: string;
  homeCta: string;
  stepTitle: string;
  steps: string[];
  globeBadge: string;
  globeTitle: string;
  globeHint: string;
  liteModeLabel: string;
  modePrefix: string;
  mode3D: string;
  modeLite: string;
  detailsBadge: string;
  detailsTitle: string;
  durationLabel: string;
  difficultyLabel: string;
  ratingLabel: string;
  highlightsTitle: string;
  itineraryTitle: string;
  primaryCta: string;
  secondaryCta: string;
  tourListTitle: string;
  tourListSubtitle: string;
  listCardCta: string;
  slotsLabel: string;
  selectedLabel: string;
  stickyCta: string;
};

const TRAVEL_COPY: Record<AppLocale, TravelCopy> = {
  vi: {
    documentTitle: "Du lich 3D | Falzo",
    navLabel: "FALZO 3D TRAVEL LAB",
    pageTitle: "Chon diem den tren globe 3D va dat tour ngay",
    pageSubtitle:
      "Page nay duoc thiet ke de giup khach ra quyet dinh nhanh: chon diem den, xem lich trinh, bam CTA dat tour.",
    loginCta: "Dang nhap",
    homeCta: "Ve trang chu",
    stepTitle: "Hanh trinh dat tour",
    steps: ["1. Chon diem den", "2. Xem lich trinh", "3. Dat tour"],
    globeBadge: "Ban do trai nghiem 3D",
    globeTitle: "Tuong tac voi globe de thay doi diem den",
    globeHint: "Cham vao marker sang de cap nhat panel thong tin ben phai.",
    liteModeLabel: "Che do nhe cho mobile / reduced motion",
    modePrefix: "Che do",
    mode3D: "3D WebGL",
    modeLite: "Lite",
    detailsBadge: "Chi tiet diem den",
    detailsTitle: "Thong tin tour theo diem den da chon",
    durationLabel: "Thoi luong",
    difficultyLabel: "Do kho",
    ratingLabel: "Danh gia",
    highlightsTitle: "Diem nhan",
    itineraryTitle: "Khung lich trinh",
    primaryCta: "Dat tu van tour nay",
    secondaryCta: "Dang nhap de dat tour",
    tourListTitle: "Tat ca goi tour",
    tourListSubtitle: "Card danh sach de quet nhanh va doi diem den tren panel 3D.",
    listCardCta: "Chon tour",
    slotsLabel: "slot con lai",
    selectedLabel: "Da chon",
    stickyCta: "Dat tour da chon",
  },
  en: {
    documentTitle: "3D Travel | Falzo",
    navLabel: "FALZO 3D TRAVEL LAB",
    pageTitle: "Pick destinations on 3D globe and book faster",
    pageSubtitle:
      "This page is optimized for quick decisions: choose destination, review itinerary, and hit booking CTA.",
    loginCta: "Login",
    homeCta: "Back home",
    stepTitle: "Booking flow",
    steps: ["1. Pick destination", "2. Review itinerary", "3. Book tour"],
    globeBadge: "3D experience map",
    globeTitle: "Interact with the globe to switch destination",
    globeHint: "Tap highlighted markers to update the details panel instantly.",
    liteModeLabel: "Lite mode for mobile / reduced motion",
    modePrefix: "Mode",
    mode3D: "3D WebGL",
    modeLite: "Lite",
    detailsBadge: "Destination details",
    detailsTitle: "Tour information based on selected destination",
    durationLabel: "Duration",
    difficultyLabel: "Difficulty",
    ratingLabel: "Rating",
    highlightsTitle: "Highlights",
    itineraryTitle: "Itinerary outline",
    primaryCta: "Request this tour",
    secondaryCta: "Login to book now",
    tourListTitle: "All tour packages",
    tourListSubtitle: "Scan cards quickly and switch destination on the 3D panel.",
    listCardCta: "Select tour",
    slotsLabel: "slots left",
    selectedLabel: "Selected",
    stickyCta: "Book selected tour",
  },
};

const TRAVEL_STOPS: TravelStop[] = [
  {
    id: "ha-giang",
    title: {
      vi: "Ha Giang Skyline Loop",
      en: "Ha Giang Skyline Loop",
    },
    country: {
      vi: "Viet Nam",
      en: "Vietnam",
    },
    description: {
      vi: "Cung duong de cao, song cham tai ban dia va san may buoi sang tren Ma Pi Leng.",
      en: "Mountain passes, local village experiences, and cloud chasing above Ma Pi Leng.",
    },
    highlights: [
      {
        vi: "Chay xe qua Ma Pi Leng Pass",
        en: "Ride through Ma Pi Leng Pass",
      },
      {
        vi: "Check-in Lung Cu Flag Tower",
        en: "Check in at Lung Cu Flag Tower",
      },
      {
        vi: "Food tour am thuc ban dia",
        en: "Local food tasting route",
      },
    ],
    itinerary: [
      {
        vi: "Ngay 1: Ha Noi -> Ha Giang, orientation va cho dem.",
        en: "Day 1: Hanoi -> Ha Giang, orientation and local night market.",
      },
      {
        vi: "Ngay 2: Dong Van loop, viewpoints va village walk.",
        en: "Day 2: Dong Van loop, viewpoints and village walk.",
      },
      {
        vi: "Ngay 3: Ma Pi Leng canyon + boat on Nho Que.",
        en: "Day 3: Ma Pi Leng canyon + Nho Que boat trip.",
      },
      {
        vi: "Ngay 4: Relax brunch va quay lai Ha Noi.",
        en: "Day 4: Relaxed brunch and return to Hanoi.",
      },
    ],
    difficulty: {
      vi: "Trung binh",
      en: "Moderate",
    },
    days: "4N3D",
    price: "6.9M VND",
    rating: "4.9",
    color: "#62b8ff",
    lat: 22.8233,
    lng: 104.9836,
  },
  {
    id: "kyoto",
    title: {
      vi: "Kyoto Lantern Evenings",
      en: "Kyoto Lantern Evenings",
    },
    country: {
      vi: "Nhat Ban",
      en: "Japan",
    },
    description: {
      vi: "Can bang giua den co, trai nghiem tra dao va pho den buoi toi tai Gion.",
      en: "Balanced between temples, tea ceremony, and night-lantern district in Gion.",
    },
    highlights: [
      {
        vi: "Sunrise tai Fushimi Inari",
        en: "Fushimi Inari sunrise route",
      },
      {
        vi: "Am thuc pho dem Gion",
        en: "Gion street food and evening walk",
      },
      {
        vi: "Bamboo trail o Arashiyama",
        en: "Arashiyama bamboo trail",
      },
    ],
    itinerary: [
      {
        vi: "Ngay 1: Arrive Kyoto, old-town check-in, light city walk.",
        en: "Day 1: Arrive Kyoto, old-town check-in, light city walk.",
      },
      {
        vi: "Ngay 2: Temple trail + tea ceremony session.",
        en: "Day 2: Temple trail + tea ceremony session.",
      },
      {
        vi: "Ngay 3: Arashiyama + shopping + photo spots.",
        en: "Day 3: Arashiyama + shopping + photo spots.",
      },
      {
        vi: "Ngay 4-5: Free day and departure.",
        en: "Day 4-5: Free day and departure.",
      },
    ],
    difficulty: {
      vi: "De",
      en: "Easy",
    },
    days: "5N4D",
    price: "1,090 USD",
    rating: "4.8",
    color: "#ffd27e",
    lat: 35.0116,
    lng: 135.7681,
  },
  {
    id: "queenstown",
    title: {
      vi: "Queenstown Alpine Flow",
      en: "Queenstown Alpine Flow",
    },
    country: {
      vi: "New Zealand",
      en: "New Zealand",
    },
    description: {
      vi: "Hanh trinh thien nhien cao cap voi ho bang, trekking nhe va sky-view flight.",
      en: "Premium nature route with glacier lakes, light trekking, and scenic sky flight.",
    },
    highlights: [
      {
        vi: "Cruise Milford Sound",
        en: "Milford Sound cruise",
      },
      {
        vi: "Sunset tai Lake Wakatipu",
        en: "Lake Wakatipu sunset session",
      },
      {
        vi: "Skyline Gondola va luge",
        en: "Skyline Gondola and luge",
      },
    ],
    itinerary: [
      {
        vi: "Ngay 1: Arrival + central district warm-up.",
        en: "Day 1: Arrival + central district warm-up.",
      },
      {
        vi: "Ngay 2: Alpine trail + lake picnic.",
        en: "Day 2: Alpine trail + lake picnic.",
      },
      {
        vi: "Ngay 3: Milford Sound full-day excursion.",
        en: "Day 3: Milford Sound full-day excursion.",
      },
      {
        vi: "Ngay 4-6: Flexible adventure + departure.",
        en: "Day 4-6: Flexible adventure + departure.",
      },
    ],
    difficulty: {
      vi: "Trung binh",
      en: "Moderate",
    },
    days: "6N5D",
    price: "1,390 USD",
    rating: "4.9",
    color: "#8ce3b0",
    lat: -45.0312,
    lng: 168.6626,
  },
];

function latLngToCartesian(lat: number, lng: number, radius: number): [number, number, number] {
  const phi = ((90 - lat) * Math.PI) / 180;
  const theta = ((lng + 180) * Math.PI) / 180;
  const x = -radius * Math.sin(phi) * Math.cos(theta);
  const z = radius * Math.sin(phi) * Math.sin(theta);
  const y = radius * Math.cos(phi);
  return [x, y, z];
}

function canUseWebGL() {
  if (typeof window === "undefined") {
    return false;
  }

  try {
    const canvas = document.createElement("canvas");
    return Boolean(canvas.getContext("webgl") || canvas.getContext("experimental-webgl"));
  } catch {
    return false;
  }
}

function use3DReady() {
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const motionQuery = window.matchMedia("(prefers-reduced-motion: reduce)");

    const evaluate = () => {
      const isMobile = window.innerWidth < 768;
      const disable3D = motionQuery.matches || isMobile || !canUseWebGL();
      setReady(!disable3D);
    };

    evaluate();
    window.addEventListener("resize", evaluate);
    motionQuery.addEventListener("change", evaluate);

    return () => {
      window.removeEventListener("resize", evaluate);
      motionQuery.removeEventListener("change", evaluate);
    };
  }, []);

  return ready;
}

function Globe({
  stops,
  activeStopId,
  onSelectStop,
}: {
  stops: TravelStop[];
  activeStopId: string;
  onSelectStop: (id: string) => void;
}) {
  const globeRef = useRef<Group>(null);

  useFrame((state, delta) => {
    if (!globeRef.current) {
      return;
    }

    globeRef.current.rotation.y += delta * 0.14;
    globeRef.current.rotation.x = Math.sin(state.clock.elapsedTime * 0.24) * 0.08;
  });

  return (
    <group ref={globeRef}>
      <mesh>
        <sphereGeometry args={[2.05, 64, 64]} />
        <meshStandardMaterial
          color="#2f86c7"
          emissive="#0d3158"
          emissiveIntensity={0.2}
          metalness={0.16}
          roughness={0.7}
        />
      </mesh>
      <mesh>
        <sphereGeometry args={[2.12, 64, 64]} />
        <meshStandardMaterial
          color="#9ad7ff"
          depthWrite={false}
          opacity={0.12}
          roughness={0.32}
          transparent
        />
      </mesh>
      <mesh rotation={[Math.PI / 2.4, 0, 0]}>
        <torusGeometry args={[2.82, 0.05, 20, 120]} />
        <meshStandardMaterial
          color="#f7c67a"
          emissive="#f4a938"
          emissiveIntensity={0.3}
          metalness={0.7}
          roughness={0.35}
        />
      </mesh>
      {stops.map((stop) => {
        const position = latLngToCartesian(stop.lat, stop.lng, 2.2);
        const isActive = stop.id === activeStopId;

        return (
          <group key={stop.id}>
            {isActive ? (
              <mesh position={position}>
                <sphereGeometry args={[0.16, 18, 18]} />
                <meshStandardMaterial
                  color={stop.color}
                  depthWrite={false}
                  emissive={stop.color}
                  emissiveIntensity={0.5}
                  opacity={0.28}
                  transparent
                />
              </mesh>
            ) : null}
            <mesh
              onPointerDown={(event) => {
                event.stopPropagation();
                onSelectStop(stop.id);
              }}
              onPointerEnter={() => {
                document.body.style.cursor = "pointer";
              }}
              onPointerLeave={() => {
                document.body.style.cursor = "default";
              }}
              position={position}
            >
              <sphereGeometry args={[isActive ? 0.11 : 0.08, 16, 16]} />
              <meshStandardMaterial
                color={stop.color}
                emissive={stop.color}
                emissiveIntensity={isActive ? 1 : 0.78}
                metalness={0.5}
                roughness={0.2}
              />
            </mesh>
          </group>
        );
      })}
    </group>
  );
}

function GlobePanel({
  enabled,
  activeStopId,
  onSelectStop,
  liteLabel,
  modeLabel,
  modeHint,
  stops,
  language,
}: {
  enabled: boolean;
  activeStopId: string;
  onSelectStop: (id: string) => void;
  liteLabel: string;
  modeLabel: string;
  modeHint: string;
  stops: TravelStop[];
  language: AppLocale;
}) {
  if (!enabled) {
    return (
      <div className="relative h-full w-full overflow-hidden bg-[radial-gradient(circle_at_50%_32%,#5ca5de_0%,#14304d_42%,#081122_100%)]">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_72%_22%,rgba(245,197,109,0.24),transparent_42%)]" />
        <div className="absolute left-1/2 top-1/2 h-56 w-56 -translate-x-1/2 -translate-y-1/2 rounded-full border border-white/25 bg-[radial-gradient(circle_at_35%_30%,#7ec5f7,#245e8f_58%,#16344e)] shadow-[0_0_82px_rgba(61,146,212,0.35)]" />
        <div className="absolute left-1/2 top-1/2 h-72 w-72 -translate-x-1/2 -translate-y-1/2 rounded-full border border-[#f7c474]/40" />
        <div className="absolute bottom-4 left-4 right-4 space-y-2">
          <p className="rounded-full border border-white/25 bg-white/10 px-3 py-1 text-xs font-medium text-white/90 backdrop-blur">
            {liteLabel}
          </p>
          <div className="flex flex-wrap gap-2">
            {stops.map((stop) => (
              <button
                className={`rounded-full border px-2.5 py-1 text-xs ${
                  activeStopId === stop.id
                    ? "border-white/80 bg-white/85 text-[#16304c]"
                    : "border-white/25 bg-white/10 text-white/90"
                }`}
                key={stop.id}
                onClick={() => onSelectStop(stop.id)}
                type="button"
              >
                {stop.country[language]}
              </button>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <>
      <Canvas camera={{ fov: 42, position: [0, 0.15, 6.2] }} dpr={[1, 1.65]}>
        <color args={["#081122"]} attach="background" />
        <fog args={["#081122", 8, 20]} attach="fog" />
        <ambientLight intensity={0.7} />
        <directionalLight color="#ffd7a2" intensity={1.2} position={[5, 5, 4]} />
        <pointLight color="#6bc0ff" intensity={0.95} position={[-5, -2, -3]} />
        <Stars count={900} depth={46} factor={4} fade radius={40} speed={0.35} />
        <Sparkles
          color="#f7e1bd"
          count={60}
          noise={0.35}
          opacity={0.5}
          scale={11}
          size={2.6}
          speed={0.14}
        />
        <Globe activeStopId={activeStopId} onSelectStop={onSelectStop} stops={stops} />
        <OrbitControls
          autoRotate
          autoRotateSpeed={0.76}
          enablePan={false}
          enableZoom={false}
          maxPolarAngle={Math.PI * 0.62}
          minPolarAngle={Math.PI * 0.38}
        />
      </Canvas>
      <div className="pointer-events-none absolute inset-x-4 bottom-4 rounded-2xl border border-white/20 bg-[#0f2137]/55 px-4 py-3 text-sm text-[#eef4fb] backdrop-blur-md">
        <p className="font-medium">{modeLabel}</p>
        <p className="text-xs text-[#cadef7]">{modeHint}</p>
      </div>
    </>
  );
}

export function Travel3DPage() {
  const { language } = useLanguage();
  const navigate = useNavigate();
  const copy = TRAVEL_COPY[language];
  const enable3D = use3DReady();
  const [selectedStopId, setSelectedStopId] = useState(TRAVEL_STOPS[0].id);

  const t = (value: LocalizedString) => value[language];
  const selectedStop = useMemo(
    () => TRAVEL_STOPS.find((stop) => stop.id === selectedStopId) ?? TRAVEL_STOPS[0],
    [selectedStopId],
  );

  useEffect(() => {
    document.title = copy.documentTitle;
  }, [copy.documentTitle]);

  useEffect(() => {
    return () => {
      document.body.style.cursor = "default";
    };
  }, []);

  return (
    <div className="relative min-h-screen bg-[linear-gradient(160deg,#eff6ff_0%,#e4effb_34%,#fff4e5_100%)] pb-26 sm:pb-14">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_10%_8%,rgba(119,181,238,0.22),transparent_28%),radial-gradient(circle_at_92%_12%,rgba(247,186,106,0.2),transparent_30%)]" />
      <div className="relative mx-auto w-full max-w-6xl px-4 py-8 sm:px-6">
        <header className="mb-6 flex flex-wrap items-center justify-between gap-3">
          <div className="space-y-1">
            <p className="text-xs font-semibold tracking-[0.18em] text-[#587ca7]">
              {copy.navLabel}
            </p>
            <h1 className="text-2xl font-semibold tracking-tight text-[#162a47] sm:text-3xl">
              {copy.pageTitle}
            </h1>
            <p className="max-w-3xl text-sm text-[#4d678b]">{copy.pageSubtitle}</p>
          </div>
          <div className="flex items-center gap-2">
            <Button asChild className="bg-[#1f4f87] text-white hover:bg-[#193f6d]">
              <Link to="/login">{copy.loginCta}</Link>
            </Button>
            <Button asChild className="border-[#aec5df] text-[#2c517c]" variant="outline">
              <Link to="/">{copy.homeCta}</Link>
            </Button>
          </div>
        </header>

        <section className="mb-5 rounded-2xl border border-[#d7e6f6] bg-white/75 p-4 shadow-[0_24px_58px_-44px_rgba(22,52,90,0.75)] backdrop-blur-sm">
          <p className="text-xs font-semibold tracking-[0.16em] text-[#5a7da7]">{copy.stepTitle}</p>
          <div className="mt-2 grid gap-2 sm:grid-cols-3">
            {copy.steps.map((step) => (
              <div
                className="rounded-xl border border-[#d8e6f6] bg-[#f6faff] px-3 py-2 text-sm text-[#355a86]"
                key={step}
              >
                {step}
              </div>
            ))}
          </div>
        </section>

        <section className="grid gap-5 lg:grid-cols-[1.03fr_0.97fr]">
          <motion.div
            animate={{ opacity: 1, y: 0 }}
            className="space-y-3 rounded-3xl border border-[#d4e3f4] bg-white/78 p-4 shadow-[0_30px_64px_-46px_rgba(23,52,92,0.78)] backdrop-blur-sm sm:p-5"
            initial={{ opacity: 0, y: 14 }}
            transition={{ duration: 0.34, ease: "easeOut" }}
          >
            <div className="flex items-center justify-between gap-3">
              <Badge className="bg-[#e8f3ff] text-[#275d97]" variant="secondary">
                {copy.globeBadge}
              </Badge>
              <span className="text-xs text-[#5c789d]">{copy.globeHint}</span>
            </div>
            <h2 className="text-xl font-semibold tracking-tight text-[#16304e]">{copy.globeTitle}</h2>

            <div className="relative h-[430px] overflow-hidden rounded-2xl border border-[#d7e6f6] bg-[#081122] shadow-[0_34px_70px_-46px_rgba(7,19,35,0.95)]">
              <GlobePanel
                activeStopId={selectedStop.id}
                enabled={enable3D}
                liteLabel={copy.liteModeLabel}
                modeHint={copy.globeHint}
                modeLabel={`${copy.modePrefix}: ${enable3D ? copy.mode3D : copy.modeLite}`}
                onSelectStop={setSelectedStopId}
                stops={TRAVEL_STOPS}
                language={language}
              />
            </div>
          </motion.div>

          <motion.div
            animate={{ opacity: 1, y: 0 }}
            className="space-y-4 rounded-3xl border border-[#d4e3f4] bg-white/82 p-5 shadow-[0_30px_64px_-46px_rgba(23,52,92,0.78)] backdrop-blur-sm sm:p-6"
            initial={{ opacity: 0, y: 14 }}
            transition={{ duration: 0.4, delay: 0.05, ease: "easeOut" }}
          >
            <div className="space-y-2">
              <Badge className="bg-[#e8f3ff] text-[#275d97]" variant="secondary">
                {copy.detailsBadge}
              </Badge>
              <h2 className="text-xl font-semibold tracking-tight text-[#15304d]">{copy.detailsTitle}</h2>
            </div>

            <Card className="border border-[#d9e6f5] bg-white py-0">
              <CardContent className="space-y-4 p-5">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-xs font-semibold tracking-[0.12em] text-[#6787ad] uppercase">
                      {t(selectedStop.country)}
                    </p>
                    <h3 className="mt-1 text-2xl font-semibold tracking-tight text-[#173457]">
                      {t(selectedStop.title)}
                    </h3>
                  </div>
                  <span
                    aria-hidden
                    className="mt-2 inline-block h-4 w-4 rounded-full shadow-[0_0_0_5px_rgba(255,255,255,0.85)]"
                    style={{ backgroundColor: selectedStop.color }}
                  />
                </div>

                <p className="text-sm leading-7 text-[#4a668b]">{t(selectedStop.description)}</p>

                <div className="grid gap-2 sm:grid-cols-3">
                  <div className="rounded-xl border border-[#d6e5f4] bg-[#f7fbff] px-3 py-2">
                    <p className="text-[11px] text-[#6280a8] uppercase">{copy.durationLabel}</p>
                    <p className="mt-1 text-sm font-semibold text-[#214a7a]">{selectedStop.days}</p>
                  </div>
                  <div className="rounded-xl border border-[#d6e5f4] bg-[#f7fbff] px-3 py-2">
                    <p className="text-[11px] text-[#6280a8] uppercase">{copy.difficultyLabel}</p>
                    <p className="mt-1 text-sm font-semibold text-[#214a7a]">{t(selectedStop.difficulty)}</p>
                  </div>
                  <div className="rounded-xl border border-[#d6e5f4] bg-[#f7fbff] px-3 py-2">
                    <p className="text-[11px] text-[#6280a8] uppercase">{copy.ratingLabel}</p>
                    <p className="mt-1 inline-flex items-center gap-1 text-sm font-semibold text-[#214a7a]">
                      <Star className="size-3.5 fill-[#f4ad42] text-[#f4ad42]" />
                      {selectedStop.rating}
                    </p>
                  </div>
                </div>

                <div className="space-y-2">
                  <p className="text-sm font-semibold text-[#1d3c62]">{copy.highlightsTitle}</p>
                  <ul className="space-y-1.5 text-sm text-[#436183]">
                    {selectedStop.highlights.map((item) => (
                      <li className="flex items-start gap-2" key={`${selectedStop.id}-${item.en}`}>
                        <Flag className="mt-0.5 size-4 text-[#2c5c94]" />
                        {t(item)}
                      </li>
                    ))}
                  </ul>
                </div>

                <div className="space-y-2">
                  <p className="text-sm font-semibold text-[#1d3c62]">{copy.itineraryTitle}</p>
                  <ul className="space-y-2">
                    {selectedStop.itinerary.map((item, index) => (
                      <li className="flex gap-2.5 text-sm text-[#436183]" key={`${selectedStop.id}-itinerary-${index + 1}`}>
                        <span className="mt-0.5 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-[#e7f2ff] text-[11px] font-semibold text-[#275d97]">
                          {index + 1}
                        </span>
                        <span>{t(item)}</span>
                      </li>
                    ))}
                  </ul>
                </div>

                <div className="flex flex-wrap gap-2">
                  <Button className="h-10 bg-[#1f4f87] text-white hover:bg-[#193f6d]" type="button">
                    <Compass />
                    {copy.primaryCta}
                  </Button>
                  <Button
                    className="h-10 border-[#b9cde4] text-[#2d527d]"
                    onClick={() => navigate("/login")}
                    type="button"
                    variant="outline"
                  >
                    <Plane />
                    {copy.secondaryCta}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </motion.div>
        </section>

        <section className="mt-8 space-y-4">
          <div className="space-y-1">
            <h3 className="text-2xl font-semibold tracking-tight text-[#173457]">{copy.tourListTitle}</h3>
            <p className="text-sm text-[#4a658b]">{copy.tourListSubtitle}</p>
          </div>
          <div className="grid gap-4 md:grid-cols-3">
            {TRAVEL_STOPS.map((stop) => (
              <Card
                className={`border py-0 shadow-[0_20px_40px_-34px_rgba(22,51,89,0.72)] ${
                  selectedStop.id === stop.id
                    ? "border-[#7fb0e2] bg-[#f2f8ff]"
                    : "border-[#d5e3f2] bg-white/90"
                }`}
                key={stop.id}
              >
                <CardContent className="space-y-3 p-4">
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <p className="text-xs font-semibold tracking-[0.1em] text-[#6786ad] uppercase">
                        {t(stop.country)}
                      </p>
                      <h4 className="mt-1 text-lg font-semibold text-[#173457]">{t(stop.title)}</h4>
                    </div>
                    <span
                      aria-hidden
                      className="inline-block h-3.5 w-3.5 rounded-full"
                      style={{ backgroundColor: stop.color }}
                    />
                  </div>
                  <div className="flex flex-wrap items-center gap-2 text-xs">
                    <Badge className="bg-[#eaf4ff] text-[#23578f]" variant="secondary">
                      <CalendarDays className="size-3.5" />
                      {stop.days}
                    </Badge>
                    <Badge className="bg-[#fff2de] text-[#86581c]" variant="secondary">
                      {stop.price}
                    </Badge>
                    <Badge className="bg-[#edf7f1] text-[#2a6b4d]" variant="secondary">
                      {stop.id === selectedStop.id
                        ? copy.selectedLabel
                        : `${Math.max(3, 8 - stop.id.length)} ${copy.slotsLabel}`}
                    </Badge>
                  </div>
                  <Button
                    className="h-9 w-full bg-[#20518a] text-white hover:bg-[#19406f]"
                    onClick={() => setSelectedStopId(stop.id)}
                    type="button"
                  >
                    {copy.listCardCta}
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>
      </div>

      <div className="fixed inset-x-4 bottom-4 z-40 md:hidden">
        <Button
          className="h-11 w-full bg-[#1f4f87] text-white shadow-[0_18px_34px_-16px_rgba(15,37,64,0.85)] hover:bg-[#1a406f]"
          onClick={() => navigate("/login")}
          type="button"
        >
          <SparklesIcon />
          {copy.stickyCta}
        </Button>
      </div>
    </div>
  );
}
