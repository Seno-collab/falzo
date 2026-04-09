import { motion } from "motion/react";
import {
  ArrowRight,
  Layers2,
  MousePointer2,
  Palette,
  Play,
  Shapes,
  Sparkles,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

const featureCards = [
  {
    title: "Bold 2D direction",
    description:
      "Flat color blocks, playful geometry, and crisp composition instead of glossy depth.",
    icon: Palette,
    color: "bg-[#ff8f6b]",
  },
  {
    title: "Interactive motion",
    description:
      "Small animated transitions keep the page lively without drifting into heavy 3D effects.",
    icon: MousePointer2,
    color: "bg-[#ffcb57]",
  },
  {
    title: "Flexible sections",
    description:
      "The layout is ready for product, studio, agency, or portfolio content with minimal edits.",
    icon: Layers2,
    color: "bg-[#5ec7a4]",
  },
];

const processCards = [
  {
    step: "01",
    title: "Sketch the mood",
    description:
      "Start with clean visual blocks, a simple grid, and one clear personality for the brand.",
  },
  {
    step: "02",
    title: "Build the scene",
    description:
      "Compose 2D cards, shapes, and illustrations so each screen feels intentional and readable.",
  },
  {
    step: "03",
    title: "Ship the story",
    description:
      "Turn the visual system into a homepage, product page, or campaign page without overcomplicating it.",
  },
];

export function HomePage() {
  return (
    <div className="relative overflow-hidden pb-16">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(255,214,116,0.35),transparent_22%),radial-gradient(circle_at_bottom_right,rgba(94,199,164,0.22),transparent_24%)]" />
      <div className="pointer-events-none absolute left-6 top-28 h-20 w-20 rounded-[28px] border-2 border-[#2a3556] bg-[#ffcb57] md:left-14 md:top-36" />
      <div className="pointer-events-none absolute right-4 top-44 h-14 w-14 rotate-12 rounded-full border-2 border-[#2a3556] bg-[#ff8f6b] md:right-16 md:top-52" />

      <header className="relative z-10 mx-auto flex max-w-7xl items-center justify-between px-6 py-6 lg:px-10">
        <div className="inline-flex items-center gap-3 rounded-full border-2 border-[#2a3556] bg-white px-4 py-2 text-sm font-black text-[#2a3556]">
          <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-[#2a3556] text-white">
            2D
          </span>
          Flat Studio
        </div>

        <nav className="hidden items-center gap-8 text-sm font-semibold text-[#334064] md:flex">
          <a href="#about">About</a>
          <a href="#features">Features</a>
          <a href="#process">Process</a>
        </nav>

        <Button className="rounded-full border-2 border-[#2a3556] bg-[#2a3556] px-5 text-white shadow-none hover:bg-[#334064]">
          Start project
        </Button>
      </header>

      <main className="relative z-10 mx-auto max-w-7xl space-y-20 px-6 lg:px-10">
        <motion.section
          animate={{ opacity: 1, y: 0 }}
          className="grid min-h-[calc(100vh-120px)] items-center gap-10 lg:grid-cols-[0.9fr_1.1fr]"
          initial={{ opacity: 0, y: 18 }}
          transition={{ duration: 0.45, ease: "easeOut" }}
        >
          <div className="space-y-8">
            <Badge className="rounded-full border-2 border-[#2a3556] bg-white px-4 py-2 text-[#2a3556]" variant="secondary">
              <Sparkles className="mr-2 size-3.5" />
              2D website concept
            </Badge>

            <div className="space-y-4">
              <h1 className="max-w-xl text-5xl font-black leading-none tracking-tight text-[#22304f] md:text-6xl">
                Build a bright 2D website with playful shapes and clear sections.
              </h1>
              <p className="max-w-xl text-base leading-7 text-[#52627f] md:text-lg">
                The page now leans into a flat visual language with geometric
                illustrations, strong outlines, and friendly motion instead of a
                login-focused layout.
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-3">
              <Button className="rounded-full border-2 border-[#2a3556] bg-[#2a3556] px-6 py-6 text-white shadow-none hover:bg-[#334064]" size="lg">
                Explore layout
                <ArrowRight className="size-4" />
              </Button>
              <Button
                className="rounded-full border-2 border-[#2a3556] bg-white px-6 py-6 text-[#2a3556] shadow-none hover:bg-[#fff3cf]"
                size="lg"
                variant="outline"
              >
                <Play className="size-4 fill-current" />
                Watch style
              </Button>
            </div>

            <div className="grid gap-4 sm:grid-cols-3">
              {[
                { value: "12", label: "Flat sections" },
                { value: "06", label: "Shape styles" },
                { value: "100%", label: "2D energy" },
              ].map((item) => (
                <Card
                  className="rounded-[28px] border-2 border-[#2a3556] bg-white shadow-none"
                  key={item.label}
                >
                  <CardContent className="p-5">
                    <p className="text-3xl font-black text-[#22304f]">{item.value}</p>
                    <p className="mt-1 text-sm font-medium text-[#60708c]">
                      {item.label}
                    </p>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>

          <motion.div
            animate={{ opacity: 1, scale: 1 }}
            className="relative mx-auto w-full max-w-2xl"
            initial={{ opacity: 0, scale: 0.96 }}
            transition={{ duration: 0.5, ease: "easeOut" }}
          >
            <div className="relative aspect-[11/10] overflow-hidden rounded-[42px] border-2 border-[#2a3556] bg-[#fff8e6] shadow-[12px_12px_0_#2a3556]">
              <div className="absolute left-8 top-8 h-24 w-24 rounded-full border-2 border-[#2a3556] bg-[#ffcb57]" />
              <div className="absolute right-10 top-12 h-14 w-32 rounded-full border-2 border-[#2a3556] bg-white" />
              <div className="absolute right-20 top-28 h-14 w-24 rounded-full border-2 border-[#2a3556] bg-white" />

              <div className="absolute bottom-0 left-0 right-0 h-40 bg-[#5ec7a4]" />

              <div className="absolute left-10 top-[8.5rem] h-56 w-44 rounded-[32px] border-2 border-[#2a3556] bg-[#7ba5ff]">
                <div className="absolute left-6 top-6 h-8 w-20 rounded-full border-2 border-[#2a3556] bg-white" />
                <div className="absolute left-6 top-20 h-24 w-32 rounded-[24px] border-2 border-[#2a3556] bg-[#ff8f6b]" />
                <div className="absolute left-8 top-[10.5rem] h-4 w-24 rounded-full bg-[#2a3556]" />
              </div>

              <div className="absolute left-64 top-24 h-64 w-56 rounded-[36px] border-2 border-[#2a3556] bg-white">
                <div className="absolute left-6 top-6 right-6 h-12 rounded-[20px] border-2 border-[#2a3556] bg-[#ffcb57]" />
                <div className="absolute left-6 right-6 top-[6.5rem] grid grid-cols-2 gap-4">
                  <div className="h-24 rounded-[22px] border-2 border-[#2a3556] bg-[#5ec7a4]" />
                  <div className="h-24 rounded-[22px] border-2 border-[#2a3556] bg-[#7ba5ff]" />
                </div>
                <div className="absolute bottom-6 left-6 right-6 h-24 rounded-[24px] border-2 border-[#2a3556] bg-[#ffdfe1]" />
              </div>

              <div className="absolute bottom-10 right-10 h-40 w-40 rounded-[30px] border-2 border-[#2a3556] bg-[#ff8f6b]">
                <div className="absolute left-7 top-7 h-14 w-14 rounded-full border-2 border-[#2a3556] bg-[#fff8e6]" />
                <div className="absolute bottom-7 left-7 right-7 h-6 rounded-full bg-[#2a3556]" />
              </div>

              <div className="absolute left-52 top-[4.5rem] h-20 w-20 -rotate-12 rounded-[24px] border-2 border-[#2a3556] bg-[#5ec7a4]" />
              <div className="absolute right-8 bottom-44 h-16 w-16 rotate-12 rounded-full border-2 border-[#2a3556] bg-[#7ba5ff]" />
            </div>
          </motion.div>
        </motion.section>

        <section className="space-y-8" id="about">
          <div className="max-w-2xl space-y-3">
            <Badge className="rounded-full border-2 border-[#2a3556] bg-[#fff3cf] px-4 py-2 text-[#2a3556]" variant="secondary">
              Why 2D
            </Badge>
            <h2 className="text-3xl font-black tracking-tight text-[#22304f] md:text-4xl">
              A 2D website feels clean, memorable, and easier to control visually.
            </h2>
            <p className="text-base leading-7 text-[#5a6a86]">
              The design stays expressive without relying on glossy materials or
              heavy 3D illustration. That keeps the interface lighter and the brand
              more distinct.
            </p>
          </div>

          <div className="grid gap-5 md:grid-cols-3" id="features">
            {featureCards.map((item, index) => (
              <motion.div
                animate={{ opacity: 1, y: 0 }}
                initial={{ opacity: 0, y: 14 }}
                key={item.title}
                transition={{ delay: index * 0.08, duration: 0.35 }}
              >
                <Card className="h-full rounded-[30px] border-2 border-[#2a3556] bg-white shadow-[8px_8px_0_#2a3556]">
                  <CardContent className="p-6">
                    <div
                      className={`inline-flex rounded-[22px] border-2 border-[#2a3556] p-4 text-[#2a3556] ${item.color}`}
                    >
                      <item.icon className="size-6" />
                    </div>
                    <h3 className="mt-5 text-xl font-black text-[#22304f]">
                      {item.title}
                    </h3>
                    <p className="mt-3 text-sm leading-6 text-[#5a6a86]">
                      {item.description}
                    </p>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </div>
        </section>

        <section className="grid gap-8 lg:grid-cols-[0.92fr_1.08fr]" id="process">
          <Card className="rounded-[34px] border-2 border-[#2a3556] bg-[#2a3556] text-white shadow-none">
            <CardContent className="p-8">
              <div className="space-y-5">
                <Badge className="rounded-full border-0 bg-white/14 px-4 py-2 text-white" variant="secondary">
                  Process
                </Badge>
                <h2 className="text-3xl font-black tracking-tight md:text-4xl">
                  A simple system to turn flat shapes into a full website.
                </h2>
                <p className="text-sm leading-7 text-white/74">
                  The page is already structured into reusable hero, feature,
                  process, and call-to-action sections, so it is easy to reshape
                  around your actual product content.
                </p>

                <div className="grid gap-4 pt-3">
                  {processCards.map((item) => (
                    <div
                      className="rounded-[24px] border border-white/14 bg-white/10 p-5"
                      key={item.step}
                    >
                      <div className="mb-3 inline-flex rounded-full bg-[#ffcb57] px-3 py-1 text-xs font-black text-[#2a3556]">
                        {item.step}
                      </div>
                      <h3 className="text-lg font-bold">{item.title}</h3>
                      <p className="mt-2 text-sm leading-6 text-white/72">
                        {item.description}
                      </p>
                    </div>
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>

          <div className="grid gap-5 sm:grid-cols-2">
            <Card className="rounded-[30px] border-2 border-[#2a3556] bg-[#ffcb57] shadow-[8px_8px_0_#2a3556]">
              <CardContent className="p-6">
                <div className="mb-4 inline-flex rounded-[20px] border-2 border-[#2a3556] bg-white p-3 text-[#2a3556]">
                  <Shapes className="size-6" />
                </div>
                <h3 className="text-xl font-black text-[#22304f]">
                  Shape-first composition
                </h3>
                <p className="mt-3 text-sm leading-6 text-[#30425f]">
                  Rounded blocks, outlined circles, and flat surfaces make the page
                  feel like a designed illustration rather than a template.
                </p>
              </CardContent>
            </Card>

            <Card className="rounded-[30px] border-2 border-[#2a3556] bg-[#7ba5ff] shadow-[8px_8px_0_#2a3556]">
              <CardContent className="p-6">
                <div className="mb-4 inline-flex rounded-[20px] border-2 border-[#2a3556] bg-white p-3 text-[#2a3556]">
                  <MousePointer2 className="size-6" />
                </div>
                <h3 className="text-xl font-black text-[#22304f]">
                  Easy to extend
                </h3>
                <p className="mt-3 text-sm leading-6 text-[#30425f]">
                  You can plug in pricing, portfolio, team, or product sections
                  without changing the overall style direction.
                </p>
              </CardContent>
            </Card>

            <Card className="sm:col-span-2 rounded-[34px] border-2 border-[#2a3556] bg-white shadow-[8px_8px_0_#2a3556]">
              <CardContent className="flex flex-col gap-6 p-6 md:flex-row md:items-center md:justify-between md:p-8">
                <div className="max-w-2xl space-y-3">
                  <Badge className="rounded-full border-2 border-[#2a3556] bg-[#e8fff2] px-4 py-2 text-[#2a3556]" variant="secondary">
                    Ready next
                  </Badge>
                  <h2 className="text-3xl font-black tracking-tight text-[#22304f]">
                    The page is now a 2D website concept instead of a login screen.
                  </h2>
                  <p className="text-sm leading-7 text-[#5a6a86]">
                    If you want, the next step can be turning this into a product
                    landing page, agency site, portfolio, or a full multi-page
                    marketing website.
                  </p>
                </div>

                <Button className="rounded-full border-2 border-[#2a3556] bg-[#ff8f6b] px-6 py-6 text-[#2a3556] shadow-none hover:bg-[#ff9f80]" size="lg">
                  Continue with full site
                  <ArrowRight className="size-4" />
                </Button>
              </CardContent>
            </Card>
          </div>
        </section>
      </main>
    </div>
  );
}
