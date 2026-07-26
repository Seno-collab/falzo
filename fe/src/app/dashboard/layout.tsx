import { SessionGuard } from "@/components/session-guard";

export default function DashboardLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return <SessionGuard>{children}</SessionGuard>;
}
