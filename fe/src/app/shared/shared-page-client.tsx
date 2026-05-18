"use client";

import { useSearchParams } from "next/navigation";
import { PublicSavedCollectionScreen } from "@/features/scenic/screens/public-saved-collection-screen";

export function SharedPageClient() {
  const searchParams = useSearchParams();
  return (
    <PublicSavedCollectionScreen
      shareSlug={searchParams.get("collection") ?? ""}
    />
  );
}
