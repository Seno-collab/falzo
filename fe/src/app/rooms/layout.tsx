import { SessionGuard } from "@/components/session-guard";

export default function RoomsLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return <SessionGuard>{children}</SessionGuard>;
}
