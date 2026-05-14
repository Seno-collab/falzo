"use client";

import { useSearchParams } from "next/navigation";
import { PublicProfileScreen } from "@/features/social/screens/public-profile-screen";

export function UsersPageClient() {
  const searchParams = useSearchParams();
  return <PublicProfileScreen userId={Number(searchParams.get("userId"))} />;
}
