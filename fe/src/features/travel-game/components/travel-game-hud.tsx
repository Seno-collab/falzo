"use client";

import {
  BarChart3,
  Camera,
  CheckCircle2,
  Flame,
  Play,
  Shield,
  Trophy,
  Zap,
} from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ROUTES } from "@/lib/routes";
import type { TravelGameCopy } from "@/features/travel-game/data/travel-game-copy";

type TravelGameHudProps = {
  copy: TravelGameCopy;
  claimed: boolean;
  combo: number;
  onClaimMission: () => void;
  selectedDestination: TravelGameCopy["destinations"][keyof TravelGameCopy["destinations"]];
  selectedMission: TravelGameCopy["missions"][number];
  xpPercent: number;
  xpValue: number;
  xpTarget: number;
};

export function TravelGameHud({
  claimed,
  combo,
  copy,
  onClaimMission,
  selectedDestination,
  selectedMission,
  xpPercent,
  xpTarget,
  xpValue,
}: Readonly<TravelGameHudProps>) {
  return (
    <aside className="rounded-lg border border-[#58e6b2]/24 bg-[#071b18]/78 p-4 text-white shadow-[0_24px_54px_-34px_rgb(0_0_0/0.9)] backdrop-blur-xl">
      <div className="mb-4 grid grid-cols-2 gap-2">
        <div className="rounded-lg border border-[#f7c948]/20 bg-[#f7c948]/12 px-3 py-2">
          <p className="inline-flex items-center gap-1.5 text-[0.66rem] font-black uppercase tracking-[0.1em] text-[#f8dfa0]">
            <Flame className="size-3.5" />
            {copy.dailyStreakLabel}
          </p>
          <p className="mt-1 text-lg font-black text-white">
            {copy.dailyStreakValue}
          </p>
        </div>
        <div className="rounded-lg border border-[#45d89d]/20 bg-[#45d89d]/12 px-3 py-2">
          <p className="inline-flex items-center gap-1.5 text-[0.66rem] font-black uppercase tracking-[0.1em] text-[#9fe7cb]">
            <Zap className="size-3.5" />
            {copy.comboLabel}
          </p>
          <p className="mt-1 text-lg font-black text-white">x{combo}</p>
        </div>
      </div>

      <div className="mb-4 flex items-start justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.14em] text-[#9fe7cb]">
            {copy.hudRank}
          </p>
          <h2 className="mt-2 text-xl font-semibold">
            {selectedDestination.name}
          </h2>
          <p className="mt-1 text-sm font-bold text-[#f7c948]">
            {selectedDestination.realm}
          </p>
          <p className="mt-2 text-sm leading-6 text-white/76">
            {selectedDestination.mission}
          </p>
        </div>
        <span className="rounded-lg bg-[#f7c948] p-2 text-[#183022]">
          <Trophy className="size-5" />
        </span>
      </div>

      <div className="mb-4 rounded-lg border border-[#f7c948]/20 bg-[#f7c948]/12 px-3 py-3">
        <div className="mb-3 flex items-center justify-between gap-3">
          <div>
            <p className="text-[0.65rem] font-black uppercase tracking-[0.12em] text-[#f8dfa0]">
              {copy.activeQuestLabel}
            </p>
            <p className="mt-1 text-sm font-bold text-white">
              {selectedMission.title}
            </p>
          </div>
          <span className="shrink-0 rounded-full bg-[#f7c948] px-2 py-1 text-xs font-black text-[#183022]">
            {selectedMission.reward}
          </span>
        </div>
        <p className="text-sm leading-5 text-white/76">
          {selectedMission.description}
        </p>
        <div className="mt-4 space-y-2">
          <div className="flex items-center justify-between text-xs font-bold uppercase tracking-[0.12em] text-white/58">
            <span>{copy.xpProgressLabel}</span>
            <span>
              {xpValue} / {xpTarget} XP
            </span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-white/12">
            <div
              className="h-full rounded-full bg-[linear-gradient(90deg,#45d89d,#f7c948)] shadow-[0_0_18px_rgb(247_201_72/0.48)] transition-[width] duration-500"
              style={{ width: `${xpPercent}%` }}
            />
          </div>
        </div>
      </div>

      <Button
        className="mb-2 w-full"
        disabled={claimed}
        onClick={onClaimMission}
        size="lg"
        type="button"
        variant="default"
      >
        {claimed ? (
          <CheckCircle2 className="size-4" />
        ) : (
          <Zap className="size-4" />
        )}
        {claimed ? copy.claimedRewardLabel : copy.claimRewardCta}
      </Button>

      <Button asChild className="mb-4 w-full" size="sm" variant="outline">
        <Link href={ROUTES.upload}>
          <Play className="size-4" />
          {copy.startQuestCta}
        </Link>
      </Button>

      <div className="grid gap-2">
        {copy.hudBadges.map((badge, index) => {
          const Icon = index === 0 ? Camera : index === 1 ? BarChart3 : Shield;
          return (
            <div
              className="flex items-center justify-between rounded-lg border border-white/10 bg-white/10 px-3 py-2"
              key={badge.label}
            >
              <span className="inline-flex min-w-0 items-center gap-2 text-xs font-semibold uppercase tracking-[0.1em] text-white/68">
                <Icon className="size-3.5 shrink-0 text-[#f7c948]" />
                <span className="truncate">{badge.label}</span>
              </span>
              <span className="shrink-0 text-sm font-bold text-white">
                {badge.value}
              </span>
            </div>
          );
        })}
      </div>
    </aside>
  );
}
