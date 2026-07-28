import { SessionGuard } from "@/components/session-guard";

export default function FriendsLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return <SessionGuard>{children}</SessionGuard>;
}
