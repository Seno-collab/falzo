import { Suspense } from "react";
import { SharedPageClient } from "@/app/shared/shared-page-client";

export default function SharedPage() {
  return (
    <Suspense fallback={null}>
      <SharedPageClient />
    </Suspense>
  );
}
