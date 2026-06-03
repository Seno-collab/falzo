"use client";

import {
  CalendarClock,
  Camera,
  Compass,
  Fingerprint,
  KeyRound,
  LogOut,
  Mail,
  Palette,
  Sparkles,
  ShieldCheck,
  Upload,
  UserRound,
} from "lucide-react";
import { zodResolver } from "@hookform/resolvers/zod";
import type { ChangeEvent, ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { z } from "zod";
import { LoadingPanel } from "@/components/feedback/loading-panel";
import { AppTopbar } from "@/components/layout/app-topbar";
import { PageShell } from "@/components/layout/page-shell";
import { SectionHeading } from "@/components/layout/section-heading";
import { UserPresenceBadge } from "@/components/layout/user-presence-badge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RainbowAvatar } from "@/components/ui/rainbow-avatar";
import {
  changePasswordApi,
  clearAuthSession,
  getApiErrorMessage,
  getMeApi,
  hasAuthSession,
  logoutApi,
  updateAvatarApi,
} from "@/features/auth/api";
import type { AuthUser } from "@/features/auth/types";
import {
  getAuthUserDisplayName,
  getAuthUserInitials,
  readAuthUserText,
} from "@/features/auth/user-display";
import { uploadImageApi } from "@/features/posts/api";
import { useI18n } from "@/i18n/locale-provider";
import { ROUTES } from "@/lib/routes";
import { useQueryClient } from "@tanstack/react-query";

type PasswordFormValues = {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
};

const passwordSchema = z
  .object({
    currentPassword: z.string().min(1, "Current password is required."),
    newPassword: z
      .string()
      .min(8, "New password must be at least 8 characters.")
      .regex(/[A-Za-z]/, "New password must contain a letter.")
      .regex(/\d/, "New password must contain a number."),
    confirmPassword: z.string().min(1, "Confirm your new password."),
  })
  .refine((values) => values.newPassword === values.confirmPassword, {
    message: "New passwords do not match.",
    path: ["confirmPassword"],
  });

function formatExpiry(expires?: AuthUser["expires"]) {
  if (!expires) {
    return "Active session";
  }

  const date =
    typeof expires === "number" ? new Date(expires * 1000) : new Date(expires);

  if (Number.isNaN(date.getTime())) {
    return "Active session";
  }

  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function ProfileField({
  icon,
  label,
  value,
}: Readonly<{
  icon: ReactNode;
  label: string;
  value: string;
}>) {
  return (
    <div className="flex items-start gap-3 border-[#dbe6f2] border-b py-3 last:border-b-0">
      <div className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-lg bg-[#f2f7fc] text-[#5178a2]">
        {icon}
      </div>
      <div className="min-w-0">
        <p className="text-xs font-semibold tracking-wide text-[#7892ad] uppercase">
          {label}
        </p>
        <p className="mt-1 wrap-break-word text-sm font-semibold text-[#1d3d64]">
          {value}
        </p>
      </div>
    </div>
  );
}

const generatedAvatarStyles = [
  {
    id: "mountain",
    label: "Mountain",
    description: "Peaks and cold-weather jacket",
    palette: ["#d9f0ff", "#7fb7df", "#15365a", "#f3b05b", "#2f6fb8"],
  },
  {
    id: "beach",
    label: "Beach",
    description: "Ocean, sunset, and sunglasses",
    palette: ["#dff8ff", "#86d8ea", "#15566c", "#ffbd63", "#1fb9a6"],
  },
  {
    id: "city",
    label: "City",
    description: "Skyline and camera-ready look",
    palette: ["#eef1ff", "#aab7ef", "#25214f", "#f4c86a", "#6b61d8"],
  },
  {
    id: "camping",
    label: "Camping",
    description: "Forest, tent, and trail gear",
    palette: ["#e8f7ed", "#92c9a4", "#1d3d2f", "#ffce73", "#3a8b62"],
  },
  {
    id: "roadtrip",
    label: "Road trip",
    description: "Open road and travel scarf",
    palette: ["#fff3dd", "#d4a46c", "#3d2a16", "#ffe08a", "#d1842f"],
  },
] as const;

type GeneratedAvatarStyleId = (typeof generatedAvatarStyles)[number]["id"];

type GeneratedAvatarDraft = {
  file: File;
  previewUrl: string;
  styleId: GeneratedAvatarStyleId;
};

const generatedAvatarSkinTones = [
  "#f6c6a6",
  "#e5a879",
  "#c9855c",
  "#8f563a",
  "#f2d2b6",
] as const;

const generatedAvatarHairColors = [
  "#2b1d17",
  "#463123",
  "#6b3f22",
  "#1f2933",
  "#8a4f2d",
] as const;

function hashText(value: string) {
  return Array.from(value).reduce(
    (hash, char) => (hash * 31 + char.charCodeAt(0)) >>> 0,
    2166136261,
  );
}

function avatarInitials(displayName: string) {
  const parts = displayName.trim().split(/\s+/).filter(Boolean);
  if (parts.length <= 1) {
    return (parts[0] ?? "F").slice(0, 2).toUpperCase();
  }

  return `${parts[0]?.[0] ?? "F"}${parts.at(-1)?.[0] ?? ""}`.toUpperCase();
}

function getGeneratedAvatarStyle(styleId: GeneratedAvatarStyleId) {
  return (
    generatedAvatarStyles.find((style) => style.id === styleId) ??
    generatedAvatarStyles[0]
  );
}

function canvasToPngFile(canvas: HTMLCanvasElement, fileName: string) {
  return new Promise<File>((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (!blob) {
        reject(new Error("Unable to create avatar image."));
        return;
      }

      resolve(new File([blob], fileName, { type: "image/png" }));
    }, "image/png");
  });
}

function drawRoundedRect(
  context: CanvasRenderingContext2D,
  x: number,
  y: number,
  width: number,
  height: number,
  radius: number,
) {
  const safeRadius = Math.min(radius, width / 2, height / 2);
  context.beginPath();
  context.moveTo(x + safeRadius, y);
  context.lineTo(x + width - safeRadius, y);
  context.quadraticCurveTo(x + width, y, x + width, y + safeRadius);
  context.lineTo(x + width, y + height - safeRadius);
  context.quadraticCurveTo(
    x + width,
    y + height,
    x + width - safeRadius,
    y + height,
  );
  context.lineTo(x + safeRadius, y + height);
  context.quadraticCurveTo(x, y + height, x, y + height - safeRadius);
  context.lineTo(x, y + safeRadius);
  context.quadraticCurveTo(x, y, x + safeRadius, y);
  context.closePath();
}

async function createGeneratedAvatarFile(
  displayName: string,
  seed: string,
  styleId: GeneratedAvatarStyleId,
) {
  const size = 512;
  const hash = hashText(seed);
  const avatarStyle = getGeneratedAvatarStyle(styleId);
  const palette = avatarStyle.palette;
  const skinTone =
    generatedAvatarSkinTones[hash % generatedAvatarSkinTones.length] ??
    generatedAvatarSkinTones[0];
  const hairColor =
    generatedAvatarHairColors[(hash >> 3) % generatedAvatarHairColors.length] ??
    generatedAvatarHairColors[0];
  const canvas = document.createElement("canvas");
  canvas.width = size;
  canvas.height = size;

  const context = canvas.getContext("2d");
  if (!context) {
    throw new Error("Unable to create avatar image.");
  }

  const gradient = context.createLinearGradient(0, 0, size, size);
  gradient.addColorStop(0, palette[0]);
  gradient.addColorStop(0.48, palette[1]);
  gradient.addColorStop(1, palette[0]);
  context.fillStyle = gradient;
  context.fillRect(0, 0, size, size);

  context.fillStyle = palette[3];
  context.beginPath();
  context.arc(392, 100, 44, 0, Math.PI * 2);
  context.fill();

  context.fillStyle = "rgb(255 255 255 / 0.72)";
  drawRoundedRect(context, 82, 88, 116, 30, 18);
  context.fill();
  drawRoundedRect(context, 322, 160, 98, 26, 15);
  context.fill();

  if (avatarStyle.id === "beach") {
    context.fillStyle = palette[4];
    context.fillRect(0, 292, size, 168);
    context.strokeStyle = "rgb(255 255 255 / 0.58)";
    context.lineWidth = 7;
    for (let y = 318; y < 412; y += 34) {
      context.beginPath();
      context.moveTo(36, y);
      context.quadraticCurveTo(126, y - 22, 216, y);
      context.quadraticCurveTo(318, y + 20, 476, y - 6);
      context.stroke();
    }
    context.fillStyle = "#8b5a2b";
    drawRoundedRect(context, 72, 218, 16, 112, 8);
    context.fill();
    context.fillStyle = "#2f8f5b";
    for (let index = 0; index < 5; index += 1) {
      context.beginPath();
      context.ellipse(
        84,
        208,
        64,
        18,
        (Math.PI / 5) * index,
        0,
        Math.PI * 2,
      );
      context.fill();
    }
  } else if (avatarStyle.id === "city") {
    context.fillStyle = "rgb(37 33 79 / 0.88)";
    for (let index = 0; index < 7; index += 1) {
      const width = 38 + ((hash >> index) % 26);
      const height = 76 + ((hash >> (index + 4)) % 96);
      const x = 30 + index * 66;
      drawRoundedRect(context, x, 302 - height, width, height + 62, 8);
      context.fill();
      context.fillStyle = "rgb(255 255 255 / 0.55)";
      context.fillRect(x + 10, 320 - height, 8, 8);
      context.fillRect(x + width - 18, 348 - height, 8, 8);
      context.fillStyle = "rgb(37 33 79 / 0.88)";
    }
  } else if (avatarStyle.id === "camping") {
    context.fillStyle = "#2d6b3d";
    for (let index = 0; index < 6; index += 1) {
      const x = 34 + index * 82;
      context.beginPath();
      context.moveTo(x, 324);
      context.lineTo(x + 38, 188);
      context.lineTo(x + 78, 324);
      context.closePath();
      context.fill();
    }
    context.fillStyle = palette[3];
    context.beginPath();
    context.moveTo(76, 360);
    context.lineTo(162, 250);
    context.lineTo(248, 360);
    context.closePath();
    context.fill();
    context.strokeStyle = "rgb(255 255 255 / 0.64)";
    context.lineWidth = 7;
    context.beginPath();
    context.moveTo(162, 250);
    context.lineTo(162, 360);
    context.stroke();
  } else if (avatarStyle.id === "roadtrip") {
    context.fillStyle = palette[2];
    context.globalAlpha = 0.86;
    context.beginPath();
    context.moveTo(0, 314);
    context.lineTo(150, 186);
    context.lineTo(288, 314);
    context.closePath();
    context.fill();
    context.globalAlpha = 1;
    context.fillStyle = "#3b3440";
    context.beginPath();
    context.moveTo(178, 460);
    context.lineTo(236, 286);
    context.lineTo(276, 286);
    context.lineTo(336, 460);
    context.closePath();
    context.fill();
    context.strokeStyle = "#fff2a8";
    context.lineWidth = 8;
    context.setLineDash([20, 22]);
    context.beginPath();
    context.moveTo(258, 302);
    context.lineTo(258, 452);
    context.stroke();
    context.setLineDash([]);
  } else {
    context.fillStyle = palette[2];
    context.globalAlpha = 0.92;
    context.beginPath();
    context.moveTo(18, 322);
    context.lineTo(158, 154);
    context.lineTo(296, 322);
    context.closePath();
    context.fill();

    context.globalAlpha = 0.78;
    context.fillStyle = palette[4];
    context.beginPath();
    context.moveTo(188, 326);
    context.lineTo(338, 132);
    context.lineTo(504, 326);
    context.closePath();
    context.fill();
    context.globalAlpha = 1;

    context.fillStyle = "rgb(255 255 255 / 0.82)";
    context.beginPath();
    context.moveTo(126, 192);
    context.lineTo(158, 154);
    context.lineTo(190, 193);
    context.lineTo(160, 181);
    context.closePath();
    context.fill();
    context.beginPath();
    context.moveTo(300, 182);
    context.lineTo(338, 132);
    context.lineTo(380, 181);
    context.lineTo(342, 166);
    context.closePath();
    context.fill();
  }

  context.fillStyle = "rgb(255 255 255 / 0.2)";
  drawRoundedRect(context, 34, 326, 444, 150, 52);
  context.fill();

  context.fillStyle = palette[2];
  context.globalAlpha = 0.18;
  context.beginPath();
  context.arc(256, 292, 172, 0, Math.PI * 2);
  context.fill();
  context.globalAlpha = 1;

  context.fillStyle = "rgb(32 47 64 / 0.18)";
  context.beginPath();
  context.ellipse(256, 438, 120, 24, 0, 0, Math.PI * 2);
  context.fill();

  context.fillStyle = palette[2];
  drawRoundedRect(context, 172, 252, 168, 190, 48);
  context.fill();

  context.fillStyle = palette[4];
  drawRoundedRect(context, 156, 258, 62, 148, 30);
  context.fill();
  drawRoundedRect(context, 294, 258, 62, 148, 30);
  context.fill();

  context.strokeStyle = "rgb(255 255 255 / 0.72)";
  context.lineWidth = 9;
  context.beginPath();
  context.moveTo(206, 264);
  context.quadraticCurveTo(238, 314, 256, 386);
  context.quadraticCurveTo(274, 314, 306, 264);
  context.stroke();

  context.fillStyle = skinTone;
  context.beginPath();
  context.arc(256, 198, 70, 0, Math.PI * 2);
  context.fill();

  context.fillStyle = hairColor;
  context.beginPath();
  context.arc(256, 180, 76, Math.PI, Math.PI * 2);
  context.quadraticCurveTo(194, 164, 196, 218);
  context.quadraticCurveTo(214, 190, 250, 190);
  context.quadraticCurveTo(302, 190, 320, 220);
  context.quadraticCurveTo(324, 166, 256, 180);
  context.fill();

  context.fillStyle = "rgb(31 41 51 / 0.86)";
  context.beginPath();
  context.arc(232, 206, 6, 0, Math.PI * 2);
  context.arc(280, 206, 6, 0, Math.PI * 2);
  context.fill();

  context.strokeStyle = "rgb(31 41 51 / 0.52)";
  context.lineWidth = 5;
  context.lineCap = "round";
  context.beginPath();
  context.moveTo(238, 232);
  context.quadraticCurveTo(256, 246, 276, 232);
  context.stroke();

  if (avatarStyle.id === "beach") {
    context.strokeStyle = "rgb(31 41 51 / 0.78)";
    context.lineWidth = 7;
    context.beginPath();
    context.moveTo(212, 204);
    context.lineTo(300, 204);
    context.stroke();
    context.fillStyle = "rgb(31 41 51 / 0.9)";
    drawRoundedRect(context, 208, 193, 42, 25, 10);
    context.fill();
    drawRoundedRect(context, 262, 193, 42, 25, 10);
    context.fill();
  } else if (avatarStyle.id === "city") {
    context.fillStyle = "rgb(255 255 255 / 0.9)";
    drawRoundedRect(context, 286, 356, 52, 38, 10);
    context.fill();
    context.fillStyle = "#1f2933";
    context.beginPath();
    context.arc(312, 375, 11, 0, Math.PI * 2);
    context.fill();
  } else if (avatarStyle.id === "camping") {
    context.fillStyle = "rgb(255 255 255 / 0.88)";
    drawRoundedRect(context, 176, 392, 160, 22, 11);
    context.fill();
    context.fillStyle = palette[4];
    context.fillRect(214, 392, 84, 22);
  } else if (avatarStyle.id === "roadtrip") {
    context.fillStyle = palette[3];
    context.beginPath();
    context.moveTo(202, 290);
    context.lineTo(256, 334);
    context.lineTo(310, 290);
    context.lineTo(294, 326);
    context.lineTo(256, 356);
    context.lineTo(218, 326);
    context.closePath();
    context.fill();
  } else {
    context.fillStyle = "rgb(255 255 255 / 0.82)";
    drawRoundedRect(context, 208, 132, 96, 34, 17);
    context.fill();
    context.fillStyle = palette[2];
    drawRoundedRect(context, 214, 126, 84, 18, 9);
    context.fill();
  }

  context.fillStyle = "rgb(255 255 255 / 0.9)";
  drawRoundedRect(context, 220, 330, 72, 42, 18);
  context.fill();

  context.fillStyle = palette[2];
  context.font = "800 26px Arial, sans-serif";
  context.textAlign = "center";
  context.textBaseline = "middle";
  context.fillText(avatarInitials(displayName), 256, 352);

  context.fillStyle = "rgb(255 255 255 / 0.78)";
  context.font = "700 20px Arial, sans-serif";
  context.fillText(avatarStyle.label.toUpperCase(), size / 2, size - 46);

  return canvasToPngFile(canvas, `falzo-${avatarStyle.id}-avatar-${hash}.png`);
}

export function ProfileScreen() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { messages } = useI18n();
  const copy = messages.profilePage;
  const [isSessionChecking, setIsSessionChecking] = useState(true);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const [isGeneratingAvatar, setIsGeneratingAvatar] = useState(false);
  const [selectedAvatarStyle, setSelectedAvatarStyle] =
    useState<GeneratedAvatarStyleId>("mountain");
  const [generatedAvatarDraft, setGeneratedAvatarDraft] =
    useState<GeneratedAvatarDraft | null>(null);
  const [profile, setProfile] = useState<AuthUser | null>(null);
  const passwordForm = useForm<PasswordFormValues>({
    resolver: zodResolver(passwordSchema),
    defaultValues: {
      currentPassword: "",
      newPassword: "",
      confirmPassword: "",
    },
  });

  useEffect(() => {
    document.title = copy.documentTitle;

    if (!hasAuthSession()) {
      router.replace(ROUTES.login);
      return;
    }

    let disposed = false;

    const loadProfile = async () => {
      try {
        const user = await getMeApi<AuthUser>();
        if (!disposed) {
          setProfile(user);
          setIsSessionChecking(false);
        }
      } catch {
        if (disposed) {
          return;
        }

        clearAuthSession();
        router.replace(ROUTES.login);
      }
    };

    void loadProfile();

    return () => {
      disposed = true;
    };
  }, [copy.documentTitle, router]);

  const displayName = useMemo(() => getAuthUserDisplayName(profile), [profile]);
  const email = readAuthUserText(profile, ["email"]);
  const username = readAuthUserText(profile, ["user_name"]);
  const avatarUrl = readAuthUserText(profile, ["avatar_url", "avatarUrl"]);
  const previewAvatarUrl = generatedAvatarDraft?.previewUrl ?? avatarUrl;
  const generatedAvatarDraftStyle = generatedAvatarDraft
    ? getGeneratedAvatarStyle(generatedAvatarDraft.styleId)
    : null;
  const subject = readAuthUserText(profile, ["subject", "id"]);
  const isAvatarBusy = isUploadingAvatar || isGeneratingAvatar;

  useEffect(() => {
    return () => {
      if (generatedAvatarDraft) {
        URL.revokeObjectURL(generatedAvatarDraft.previewUrl);
      }
    };
  }, [generatedAvatarDraft]);

  const applyAvatarFile = async (file: File) => {
    const uploaded = await uploadImageApi(file);
    const updatedProfile = await updateAvatarApi(uploaded.url);
    const nextProfile = {
      ...(profile ?? {}),
      ...updatedProfile,
      avatar_url: updatedProfile.avatar_url ?? uploaded.url,
      avatarUrl: updatedProfile.avatarUrl ?? uploaded.url,
    };

    setProfile(nextProfile);
    setGeneratedAvatarDraft(null);
    queryClient.setQueryData(["me"], nextProfile);
    queryClient.setQueriesData<AuthUser>({ queryKey: ["auth"] }, (current) => ({
      ...(current ?? {}),
      ...nextProfile,
    }));
    void queryClient.invalidateQueries({ queryKey: ["me"] });
    void queryClient.invalidateQueries({ queryKey: ["auth"] });
    window.dispatchEvent(
      new CustomEvent<AuthUser>("falzo:avatar-updated", {
        detail: nextProfile,
      }),
    );
  };

  const handleAvatarFileChange = async (
    event: ChangeEvent<HTMLInputElement>,
  ) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || isAvatarBusy) {
      return;
    }

    setIsUploadingAvatar(true);
    try {
      await applyAvatarFile(file);
      toast.success(copy.profilePhotoUpdated);
    } catch (error) {
      toast.error(copy.unableToUpdatePhoto, {
        description: getApiErrorMessage(error),
      });
    } finally {
      setIsUploadingAvatar(false);
    }
  };

  const handleGenerateAvatar = async () => {
    if (isAvatarBusy) {
      return;
    }

    setIsGeneratingAvatar(true);
    try {
      const seed = [
        displayName,
        username ?? "",
        email ?? "",
        subject ?? "",
        selectedAvatarStyle,
        Date.now().toString(),
      ].join(":");
      const file = await createGeneratedAvatarFile(
        displayName,
        seed,
        selectedAvatarStyle,
      );
      setGeneratedAvatarDraft({
        file,
        previewUrl: URL.createObjectURL(file),
        styleId: selectedAvatarStyle,
      });
      toast.success(copy.newLookReady);
    } catch (error) {
      toast.error(copy.unableToPrepareLook, {
        description: getApiErrorMessage(error),
      });
    } finally {
      setIsGeneratingAvatar(false);
    }
  };

  const handleSaveGeneratedAvatar = async () => {
    if (!generatedAvatarDraft || isAvatarBusy) {
      return;
    }

    setIsUploadingAvatar(true);
    try {
      await applyAvatarFile(generatedAvatarDraft.file);
      toast.success(copy.profilePhotoUpdated);
    } catch (error) {
      toast.error(copy.unableToUpdatePhoto, {
        description: getApiErrorMessage(error),
      });
    } finally {
      setIsUploadingAvatar(false);
    }
  };

  const handleChangePassword = passwordForm.handleSubmit(async (values) => {
    try {
      await changePasswordApi({
        currentPassword: values.currentPassword,
        newPassword: values.newPassword,
      });
      passwordForm.reset();
      toast.success(copy.passwordChanged);

      setIsLoggingOut(true);
      try {
        await logoutApi();
      } catch {
        clearAuthSession();
      } finally {
        queryClient.clear();
        setIsLoggingOut(false);
        router.replace(ROUTES.login);
      }
    } catch (error) {
      toast.error(copy.unableToChangePassword, {
        description: getApiErrorMessage(error),
      });
    }
  });

  const handleLogout = async () => {
    if (isLoggingOut) {
      return;
    }

    setIsLoggingOut(true);

    try {
      await logoutApi();
      queryClient.removeQueries({ queryKey: ["me", "explore", "auth"] });
      queryClient.clear();
      router.replace(ROUTES.login);
    } finally {
      setIsLoggingOut(false);
    }
  };

  return (
    <PageShell
      contentClassName="pb-12"
      topbar={
        <AppTopbar
          actions={[
            {
              id: "explore",
              icon: <Compass className="size-4" />,
              label: "Explore",
              to: ROUTES.explore,
              variant: "outline",
            },
            {
              id: "upload",
              icon: <Upload className="size-4" />,
              label: "Upload",
              to: ROUTES.upload,
              variant: "outline",
            },
            {
              id: "logout",
              icon: <LogOut className="size-4" />,
              label: "Logout",
              onClick: () => {
                void handleLogout();
              },
              variant: "default",
            },
          ]}
          brand={copy.brand}
          brandIcon={<UserRound className="size-3.5" />}
          meta={<UserPresenceBadge />}
          mobileMenuTitle={copy.mobileMenuTitle}
          subtitle={copy.subtitle}
        />
      }
    >
      {isSessionChecking ? (
        <LoadingPanel
          description={copy.loadingDescription}
          title={copy.loadingTitle}
        />
      ) : (
        <div className="mx-auto grid max-w-6xl gap-5 lg:grid-cols-[minmax(0,0.92fr)_minmax(360px,0.58fr)]">
          <Card className="app-panel app-hover overflow-hidden border-[#d6e5f6] bg-white/94 py-0">
            <CardContent className="p-0">
              <div className="border-[#dfe9f4] border-b bg-[linear-gradient(180deg,#fafdff_0%,#f5f9fd_100%)] p-5 sm:p-7">
                <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
                  <div className="flex min-w-0 items-center gap-4">
                    <label
                      className={`group relative size-20 shrink-0 rounded-full transition ${
                        isAvatarBusy
                          ? "cursor-not-allowed opacity-70"
                          : "cursor-pointer hover:brightness-105"
                      }`}
                      title={copy.uploadTitle}
                    >
                      <RainbowAvatar
                        alt={displayName}
                        className="pointer-events-none"
                        fallback={getAuthUserInitials(displayName)}
                        size="lg"
                        src={previewAvatarUrl}
                      />
                      <div className="absolute inset-0.75 z-20 flex items-center justify-center rounded-full bg-black/38 opacity-0 transition group-hover:opacity-100 group-focus-within:opacity-100">
                        <Camera className="size-6 text-white drop-shadow-sm" />
                      </div>
                      <input
                        accept="image/jpeg,image/png,image/webp"
                        className="sr-only"
                        disabled={isAvatarBusy}
                        onChange={(event) => {
                          void handleAvatarFileChange(event);
                        }}
                        type="file"
                      />
                    </label>
                    <div className="min-w-0">
                      <Badge>{copy.signedIn}</Badge>
                      <h1 className="mt-2 truncate text-2xl font-semibold tracking-normal text-[#143052] sm:text-3xl">
                        {displayName}
                      </h1>
                      <p className="mt-1 truncate text-sm text-[#527299]">
                        {email ?? username ?? copy.fallbackAccount}
                      </p>
                      {isUploadingAvatar ? (
                        <p className="mt-2 text-xs font-semibold text-[#5178a2]">
                          {copy.uploadingPhoto}
                        </p>
                      ) : null}
                      {isGeneratingAvatar ? (
                        <p className="mt-2 inline-flex items-center gap-1.5 text-xs font-semibold text-[#6b61d8]">
                          <Sparkles className="size-3.5 animate-pulse" />
                          {copy.preparingLook}
                        </p>
                      ) : null}
                      {generatedAvatarDraft && !isAvatarBusy ? (
                        <p className="mt-2 text-xs font-semibold text-[#6b61d8]">
                          {generatedAvatarDraftStyle?.label ?? "New"}{" "}
                          {copy.lookReady}
                        </p>
                      ) : null}
                    </div>
                  </div>

                  <div className="flex w-full flex-col gap-2 sm:w-auto sm:max-w-md sm:items-end">
                    <div className="flex w-full flex-wrap gap-1.5 sm:justify-end">
                      {generatedAvatarStyles.map((style) => {
                        const active = selectedAvatarStyle === style.id;

                        return (
                          <button
                            aria-pressed={active}
                            className={`rounded-full border px-2.5 py-1 text-xs font-bold transition ${
                              active
                                ? "border-[#5147a8] bg-[#5147a8] text-white shadow-[0_10px_22px_-16px_rgb(81_71_168/0.85)]"
                                : "border-[#d7e0f5] bg-white/88 text-[#5f6f91] hover:border-[#c4cdf4] hover:bg-[#f5f4ff] hover:text-[#5147a8]"
                            }`}
                            disabled={isAvatarBusy}
                            key={style.id}
                            onClick={() => setSelectedAvatarStyle(style.id)}
                            title={style.description}
                            type="button"
                          >
                            {style.label}
                          </button>
                        );
                      })}
                    </div>
                    <div className="grid w-full gap-2 sm:w-auto sm:grid-cols-[auto_auto_auto] sm:justify-end">
                      <Button
                        className="w-full border-[#cdd7ff] bg-[#f5f4ff] text-[#5147a8] hover:bg-[#eceaff] sm:w-auto"
                        disabled={isAvatarBusy}
                        onClick={() => {
                          void handleGenerateAvatar();
                        }}
                        size="sm"
                        type="button"
                        variant="outline"
                      >
                        <Palette className="size-4" />
                        {generatedAvatarDraft
                          ? copy.tryAnotherLook
                          : copy.tryLook}
                      </Button>
                      <Button
                        className="w-full sm:w-auto"
                        onClick={() => router.push(ROUTES.explore)}
                        size="sm"
                        type="button"
                        variant="outline"
                      >
                        <Compass className="size-4" />
                        {copy.explore}
                      </Button>
                      <Button
                        className="w-full sm:w-auto"
                        onClick={() => router.push(ROUTES.upload)}
                        size="sm"
                        type="button"
                        variant="outline"
                      >
                        <Upload className="size-4" />
                        {copy.upload}
                      </Button>
                    </div>
                    <div className="grid w-full gap-2 sm:grid-cols-2">
                      {generatedAvatarDraft ? (
                        <>
                          <Button
                            className="w-full"
                            disabled={isAvatarBusy}
                            onClick={() => {
                              void handleSaveGeneratedAvatar();
                            }}
                            size="sm"
                            type="button"
                          >
                            <Sparkles className="size-4" />
                            {copy.saveLook}
                          </Button>
                          <Button
                            className="w-full"
                            disabled={isAvatarBusy}
                            onClick={() => setGeneratedAvatarDraft(null)}
                            size="sm"
                            type="button"
                            variant="outline"
                          >
                            {copy.keepCurrentPhoto}
                          </Button>
                        </>
                      ) : null}
                    </div>
                  </div>
                </div>
              </div>

              <div className="space-y-5 p-5 sm:p-7">
                <SectionHeading
                  description={copy.overviewDescription}
                  title={copy.overviewTitle}
                />

                <div className="rounded-2xl border border-[#d7e5f4] bg-white px-4 py-1">
                  <ProfileField
                    icon={<UserRound className="size-4" />}
                    label={copy.fields.username}
                    value={username ?? copy.notProvided}
                  />
                  <ProfileField
                    icon={<Mail className="size-4" />}
                    label={copy.fields.email}
                    value={email ?? copy.notProvided}
                  />
                  <ProfileField
                    icon={<Fingerprint className="size-4" />}
                    label={copy.fields.subject}
                    value={subject ?? copy.notAvailable}
                  />
                  <ProfileField
                    icon={<CalendarClock className="size-4" />}
                    label={copy.fields.tokenExpires}
                    value={formatExpiry(profile?.expires)}
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="app-panel border-[#d6e5f6] bg-white/94 py-0 lg:sticky lg:top-28 lg:self-start">
            <CardContent className="space-y-5 p-5 sm:p-7">
              <div className="space-y-2">
                <Badge>{copy.security}</Badge>
                <SectionHeading
                  description={copy.changePasswordDescription}
                  title={copy.changePasswordTitle}
                />
              </div>

              <form
                className="space-y-4"
                onSubmit={(event) => {
                  void handleChangePassword(event);
                }}
              >
                <div className="space-y-2">
                  <Label htmlFor="currentPassword">
                    {copy.currentPassword}
                  </Label>
                  <Input
                    autoComplete="current-password"
                    id="currentPassword"
                    type="password"
                    {...passwordForm.register("currentPassword")}
                  />
                  {passwordForm.formState.errors.currentPassword ? (
                    <p className="app-error">
                      {passwordForm.formState.errors.currentPassword.message}
                    </p>
                  ) : null}
                </div>

                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="newPassword">{copy.newPassword}</Label>
                    <Input
                      autoComplete="new-password"
                      id="newPassword"
                      type="password"
                      {...passwordForm.register("newPassword")}
                    />
                    {passwordForm.formState.errors.newPassword ? (
                      <p className="app-error">
                        {passwordForm.formState.errors.newPassword.message}
                      </p>
                    ) : null}
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="confirmPassword">
                      {copy.confirmPassword}
                    </Label>
                    <Input
                      autoComplete="new-password"
                      id="confirmPassword"
                      type="password"
                      {...passwordForm.register("confirmPassword")}
                    />
                    {passwordForm.formState.errors.confirmPassword ? (
                      <p className="app-error">
                        {passwordForm.formState.errors.confirmPassword.message}
                      </p>
                    ) : null}
                  </div>
                </div>

                <Button
                  className="w-full justify-center"
                  disabled={passwordForm.formState.isSubmitting}
                  type="submit"
                  variant="gradient"
                >
                  <KeyRound className="size-4" />
                  {copy.update}
                </Button>
              </form>

              <div className="border-[#dbe6f2] border-t pt-4">
                <Button
                  className="w-full justify-center"
                  disabled={isLoggingOut}
                  onClick={() => {
                    void handleLogout();
                  }}
                  type="button"
                  variant="soft"
                >
                  <ShieldCheck className="size-4" />
                  {copy.logout}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </PageShell>
  );
}
