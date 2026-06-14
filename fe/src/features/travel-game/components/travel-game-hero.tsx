"use client";

import {
  CheckCircle2,
  Flame,
  Gift,
  MapPinned,
  Plus,
  Route,
  Sparkles,
  Target,
  Trophy,
  Users,
  Zap,
} from "lucide-react";
import dynamic from "next/dynamic";
import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { TravelGameHud } from "@/features/travel-game/components/travel-game-hud";
import {
  destinationNodes,
  type DestinationNodeId,
  type TravelGameCopy,
} from "@/features/travel-game/data/travel-game-copy";
import { getScenicImageUrl } from "@/lib/scenic-images";
import { notifySuccess } from "@/lib/toast";
import { ROUTES } from "@/lib/routes";

const TravelGameScene = dynamic(
  () =>
    import("@/features/travel-game/components/travel-game-scene").then(
      (mod) => mod.TravelGameScene,
    ),
  {
    ssr: false,
    loading: () => (
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_55%_42%,rgba(84,214,156,0.32),rgba(8,27,24,0.92)_56%,rgba(5,15,18,0.98)_100%)]" />
    ),
  },
);

const leaderboardRankClasses = [
  "bg-[#f7c948] text-[#183022] shadow-[0_0_18px_rgb(247_201_72/0.42)]",
  "bg-[#d7e2ef] text-[#16312a] shadow-[0_0_18px_rgb(215_226_239/0.32)]",
  "bg-[#d69b63] text-[#1f140c] shadow-[0_0_18px_rgb(214_155_99/0.32)]",
];

const initialXp = 680;
const xpTarget = 1000;

function missionRewardValue(reward: string) {
  const match = reward.match(/\d+/);
  return match ? Number.parseInt(match[0], 10) : 0;
}

export function TravelGameHero({ copy }: Readonly<{ copy: TravelGameCopy }>) {
  const [activeMissionIndex, setActiveMissionIndex] = useState(0);
  const [claimedMissionIndexes, setClaimedMissionIndexes] = useState<
    Set<number>
  >(() => new Set());
  const [combo, setCombo] = useState(1);
  const [lastReward, setLastReward] = useState<string | null>(null);
  const rewardTimerRef = useRef<ReturnType<typeof globalThis.setTimeout> | null>(
    null,
  );
  const [selectedNodeId, setSelectedNodeId] =
    useState<DestinationNodeId>("phu-yen");
  const selectedNode =
    destinationNodes.find((node) => node.id === selectedNodeId) ??
    destinationNodes[0];
  const selectedDestination = copy.destinations[selectedNodeId];
  const selectedImageUrl = getScenicImageUrl(selectedNode.imageId);
  const selectedMission = copy.missions[activeMissionIndex] ?? copy.missions[0];
  const xpValue = useMemo(() => {
    const earnedXp = [...claimedMissionIndexes].reduce(
      (total, missionIndex) =>
        total + missionRewardValue(copy.missions[missionIndex]?.reward ?? ""),
      0,
    );
    return Math.min(xpTarget, initialXp + earnedXp);
  }, [claimedMissionIndexes, copy.missions]);
  const xpPercent = Math.round((xpValue / xpTarget) * 100);
  const selectedMissionClaimed = claimedMissionIndexes.has(activeMissionIndex);

  useEffect(() => {
    return () => {
      if (rewardTimerRef.current) {
        globalThis.clearTimeout(rewardTimerRef.current);
      }
    };
  }, []);

  function claimMission(missionIndex: number) {
    setActiveMissionIndex(missionIndex);
    if (claimedMissionIndexes.has(missionIndex)) {
      return;
    }

    const reward = copy.missions[missionIndex]?.reward ?? "";
    setClaimedMissionIndexes((currentIndexes) => {
      const nextIndexes = new Set(currentIndexes);
      nextIndexes.add(missionIndex);
      return nextIndexes;
    });
    setCombo((currentCombo) => Math.min(currentCombo + 1, 5));
    setLastReward(reward);
    notifySuccess(copy.rewardFeedbackTitle, {
      description: copy.rewardFeedbackDescription,
    });
    if (rewardTimerRef.current) {
      globalThis.clearTimeout(rewardTimerRef.current);
    }
    rewardTimerRef.current = globalThis.setTimeout(
      () => setLastReward(null),
      2200,
    );
  }

  return (
    <>
      <section className="relative min-h-[76svh] overflow-hidden rounded-lg bg-[#081b18] text-white shadow-[0_28px_72px_-48px_rgb(20_37_31/0.8)]">
        {selectedImageUrl ? (
          <img
            alt=""
            className="absolute inset-0 h-full w-full object-cover opacity-[0.92] transition-opacity duration-500"
            decoding="async"
            fetchPriority="high"
            key={selectedImageUrl}
            src={selectedImageUrl}
          />
        ) : null}
        <div className="absolute inset-0 bg-[linear-gradient(90deg,rgba(5,17,17,0.94),rgba(5,17,17,0.78)_36%,rgba(5,17,17,0.24)_70%,rgba(5,17,17,0.12))]" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_76%_34%,rgba(247,201,72,0.18),rgba(69,216,157,0.1)_28%,rgba(5,17,17,0)_56%)]" />
        <div className="absolute inset-x-0 bottom-0 h-40 bg-[linear-gradient(0deg,rgba(5,17,17,0.95),rgba(5,17,17,0))]" />
        <div className="absolute inset-x-0 top-16 h-[42%] opacity-55 sm:h-[48%] lg:inset-y-0 lg:left-auto lg:right-0 lg:h-auto lg:w-[58%] lg:opacity-72">
          <TravelGameScene
            copy={copy}
            nodes={destinationNodes}
            onSelectNode={setSelectedNodeId}
            selectedNodeId={selectedNodeId}
          />
        </div>

        <div className="relative grid min-h-[76svh] items-end gap-6 px-5 py-8 sm:px-8 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-10">
          <div className="flex max-w-3xl flex-col justify-end gap-6">
            <div className="space-y-4">
              <p className="inline-flex w-fit items-center gap-2 rounded-full border border-white/20 bg-white/12 px-3 py-1 text-xs font-semibold uppercase tracking-[0.14em] text-white/92 backdrop-blur">
                <Route className="size-3.5" />
                {copy.heroBadge}
              </p>
              <p className="inline-flex w-fit items-center gap-2 rounded-lg border border-[#f7c948]/30 bg-[#f7c948]/18 px-3 py-1 text-sm font-semibold text-[#fff1b8] backdrop-blur">
                <Sparkles className="size-4" />
                {copy.gameKicker}
              </p>
              <h1 className="max-w-3xl text-4xl font-semibold leading-[1.04] tracking-normal sm:text-5xl lg:text-6xl">
                {copy.headline}
              </h1>
              <p className="max-w-2xl text-base leading-7 text-white/84 sm:text-lg">
                {copy.subHeadline}
              </p>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
              <Button asChild className="sm:min-w-48" size="lg" variant="default">
                <Link href={ROUTES.explore}>
                  <Trophy className="size-4" />
                  {copy.sampleCta}
                </Link>
              </Button>
              <Button asChild className="sm:min-w-44" size="lg" variant="outline">
                <Link href={ROUTES.upload}>
                  <Plus className="size-4" />
                  {copy.createCta}
                </Link>
              </Button>
              <Button asChild className="sm:min-w-44" size="lg" variant="outline">
                <Link href={ROUTES.locations}>
                  <MapPinned className="size-4" />
                  {copy.exploreCta}
                </Link>
              </Button>
              <Button asChild className="sm:min-w-44" size="lg" variant="outline">
                <Link href={ROUTES.itineraries}>
                  <Users className="size-4" />
                  {copy.partyCta}
                </Link>
              </Button>
            </div>

            <div className="grid gap-2 text-sm font-semibold text-white/86 sm:grid-cols-3">
              {copy.proofItems.map((item) => (
                <div
                  className="rounded-lg border border-white/14 bg-white/10 px-3 py-2 backdrop-blur"
                  key={item}
                >
                  {item}
                </div>
              ))}
            </div>

            <div className="grid max-w-3xl gap-2 sm:grid-cols-3">
              <div className="rounded-lg border border-[#f7c948]/30 bg-[#f7c948]/14 px-3 py-2 backdrop-blur">
                <p className="inline-flex items-center gap-1.5 text-[0.65rem] font-black uppercase tracking-[0.12em] text-[#f8dfa0]">
                  <Flame className="size-3.5" />
                  {copy.dailyStreakLabel}
                </p>
                <p className="mt-1 text-lg font-black text-white">
                  {copy.dailyStreakValue}
                </p>
              </div>
              <div className="rounded-lg border border-[#45d89d]/30 bg-[#45d89d]/14 px-3 py-2 backdrop-blur">
                <p className="inline-flex items-center gap-1.5 text-[0.65rem] font-black uppercase tracking-[0.12em] text-[#9fe7cb]">
                  <Zap className="size-3.5" />
                  {copy.comboLabel}
                </p>
                <p className="mt-1 text-lg font-black text-white">x{combo}</p>
              </div>
              <div className="rounded-lg border border-white/16 bg-white/10 px-3 py-2 backdrop-blur">
                <p className="text-[0.65rem] font-black uppercase tracking-[0.12em] text-white/58">
                  {copy.xpProgressLabel}
                </p>
                <p className="mt-1 text-lg font-black text-white">
                  {xpValue}/{xpTarget} XP
                </p>
              </div>
            </div>

            {lastReward ? (
              <div className="w-fit rounded-lg border border-[#f7c948]/50 bg-[#f7c948] px-4 py-2 text-sm font-black text-[#16312a] shadow-[0_0_32px_rgb(247_201_72/0.42)]">
                {copy.rewardFeedbackTitle}: {lastReward}
              </div>
            ) : null}

            <div className="flex flex-wrap gap-2">
              {copy.gameplayFlow.map((step, index) => (
                <span
                  className="inline-flex items-center gap-2 rounded-full border border-white/14 bg-white/10 px-3 py-1.5 text-xs font-bold text-white/82 backdrop-blur"
                  key={step}
                >
                  <span className="inline-flex size-5 items-center justify-center rounded-full bg-[#f7c948] text-[0.65rem] text-[#183022]">
                    {index + 1}
                  </span>
                  {step}
                </span>
              ))}
            </div>

            <div className="max-w-3xl rounded-lg border border-white/14 bg-[#071b18]/58 p-3 backdrop-blur-xl">
              <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
                <p className="inline-flex items-center gap-2 text-xs font-black uppercase tracking-[0.14em] text-[#9fe7cb]">
                  <MapPinned className="size-3.5" />
                  {copy.landSelectorTitle}
                </p>
                <p className="text-xs font-semibold text-[#f7c948]">
                  {copy.selectedLandLabel}: {selectedDestination.realm}
                </p>
              </div>
              <div className="flex gap-2 overflow-x-auto pb-1">
                {destinationNodes.map((node) => {
                  const destination = copy.destinations[node.id];
                  const selected = node.id === selectedNodeId;

                  return (
                    <button
                      aria-pressed={selected}
                      className={[
                        "min-w-40 rounded-lg border px-3 py-2 text-left transition hover:-translate-y-0.5",
                        selected
                          ? "border-[#f7c948]/70 bg-[#f7c948] text-[#16312a] shadow-[0_0_24px_rgb(247_201_72/0.34)]"
                          : "border-white/12 bg-white/[0.09] text-white/78 hover:border-[#9fe7cb]/60 hover:bg-white/[0.14]",
                      ].join(" ")}
                      key={node.id}
                      onClick={() => setSelectedNodeId(node.id)}
                      type="button"
                    >
                      <span className="block text-sm font-black">
                        {destination.name}
                      </span>
                      <span
                        className={[
                          "mt-0.5 block text-xs font-semibold leading-4",
                          selected ? "text-[#315447]" : "text-white/58",
                        ].join(" ")}
                      >
                        {destination.realm}
                      </span>
                    </button>
                  );
                })}
              </div>
            </div>
          </div>

          <TravelGameHud
            claimed={selectedMissionClaimed}
            combo={combo}
            copy={copy}
            onClaimMission={() => claimMission(activeMissionIndex)}
            selectedDestination={selectedDestination}
            selectedMission={selectedMission}
            xpPercent={xpPercent}
            xpTarget={xpTarget}
            xpValue={xpValue}
          />
        </div>
      </section>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]">
        <div className="space-y-4">
          <div className="max-w-3xl space-y-2">
            <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#397168]">
              {copy.missionsTitle}
            </p>
            <h2 className="text-2xl font-semibold leading-tight text-[#162c28] sm:text-3xl">
              {copy.questTitle}
            </h2>
            <p className="text-sm leading-6 text-[#5f716f] sm:text-base">
              {copy.missionsDescription}
            </p>
          </div>

          <div className="grid gap-3 md:grid-cols-3">
            {copy.missions.map((mission, index) => (
              <button
                aria-pressed={activeMissionIndex === index}
                className={[
                  "group rounded-lg border p-4 text-left shadow-[0_16px_40px_-34px_rgb(22_44_40/0.72)] transition hover:-translate-y-1",
                  activeMissionIndex === index
                    ? "border-[#2f8f68] bg-[#10231f] text-white shadow-[0_20px_50px_-34px_rgb(16_35_31/0.84)]"
                    : "border-[#d8e6de] bg-white text-[#10231f] hover:border-[#86c8ad]",
                ].join(" ")}
                key={mission.title}
                onClick={() => setActiveMissionIndex(index)}
                type="button"
              >
                <div className="mb-4 flex items-center justify-between gap-3">
                  <span
                    className={[
                      "inline-flex size-10 items-center justify-center rounded-lg",
                      activeMissionIndex === index
                        ? "bg-[#45d89d] text-[#10231f]"
                        : "bg-[#eaf6ef] text-[#1c6b50]",
                    ].join(" ")}
                  >
                    {claimedMissionIndexes.has(index) ? (
                      <CheckCircle2 className="size-5" />
                    ) : (
                      <Target className="size-5" />
                    )}
                  </span>
                  <span className="rounded-full bg-[#fff4cf] px-3 py-1 text-xs font-bold text-[#7a5a18]">
                    {mission.reward}
                  </span>
                </div>
                <p
                  className={[
                    "text-xs font-semibold uppercase tracking-[0.14em]",
                    activeMissionIndex === index
                      ? "text-[#9fe7cb]"
                      : "text-[#78908b]",
                  ].join(" ")}
                >
                  #{index + 1} ·{" "}
                  {claimedMissionIndexes.has(index)
                    ? copy.claimedRewardLabel
                    : mission.status}
                </p>
                <h3 className="mt-2 text-lg font-semibold">
                  {mission.title}
                </h3>
                <p
                  className={[
                    "mt-2 text-sm leading-6",
                    activeMissionIndex === index
                      ? "text-white/70"
                      : "text-[#5f716f]",
                  ].join(" ")}
                >
                  {mission.description}
                </p>
              </button>
            ))}
          </div>
        </div>

        <aside className="rounded-lg border border-[#d8e6de] bg-[#10231f] p-4 text-white shadow-[0_18px_48px_-34px_rgb(16_35_31/0.84)]">
          <div className="mb-4 flex items-center gap-3">
            <span className="rounded-lg bg-[#f7c948] p-2 text-[#183022]">
              <Trophy className="size-5" />
            </span>
            <div>
              <h2 className="text-lg font-semibold">{copy.leagueTitle}</h2>
              <p className="text-sm leading-5 text-white/66">
                {copy.leagueDescription}
              </p>
            </div>
          </div>
          <div className="space-y-2">
            {copy.leaderboard.map((player, index) => (
              <div
                className="flex items-center justify-between gap-3 rounded-lg border border-white/10 bg-white/8 px-3 py-3"
                key={player.name}
              >
                <div className="flex min-w-0 items-center gap-3">
                  <span
                    className={[
                      "inline-flex size-8 shrink-0 items-center justify-center rounded-lg text-sm font-black",
                      leaderboardRankClasses[index] ??
                        "bg-white text-[#10231f]",
                    ].join(" ")}
                  >
                    {index + 1}
                  </span>
                  <div className="min-w-0">
                    <p className="truncate font-semibold">{player.name}</p>
                    <p className="truncate text-xs text-white/56">
                      {player.badge}
                    </p>
                  </div>
                </div>
                <span className="shrink-0 text-sm font-bold text-[#f7c948]">
                  {player.score}
                </span>
              </div>
            ))}
          </div>
        </aside>
      </section>

      <section className="overflow-hidden rounded-lg border border-[#d8e6de] bg-[#f3f8f5] p-4 sm:p-6">
        <div className="grid gap-6 lg:grid-cols-[320px_minmax(0,1fr)] lg:items-center">
          <div className="space-y-2">
            <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#397168]">
              {copy.pathLabel}
            </p>
            <h2 className="text-2xl font-semibold leading-tight text-[#162c28]">
              {copy.pathTitle}
            </h2>
            <p className="text-sm leading-6 text-[#5f716f]">
              {copy.pathDescription}
            </p>
          </div>
          <div className="grid gap-3 sm:grid-cols-4">
            {copy.pathSteps.map((step, index) => (
              <div
                className="relative rounded-lg border border-[#cfe0d7] bg-white p-4"
                key={step}
              >
                <span className="inline-flex size-8 items-center justify-center rounded-full bg-[#183a32] text-sm font-bold text-white">
                  {index + 1}
                </span>
                <p className="mt-4 text-sm font-semibold text-[#10231f]">
                  {step}
                </p>
                {index === copy.pathSteps.length - 1 ? (
                  <Gift className="absolute right-4 top-4 size-5 text-[#c79213]" />
                ) : null}
              </div>
            ))}
          </div>
        </div>
      </section>
    </>
  );
}
