import { Suspense } from "react";
import { UsersPageClient } from "@/app/users/users-page-client";

export default function UsersPage() {
  return (
    <Suspense fallback={null}>
      <UsersPageClient />
    </Suspense>
  );
}
