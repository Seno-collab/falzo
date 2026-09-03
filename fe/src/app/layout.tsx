import type { Metadata } from "next";
import { ApiLoadingProvider } from "@/components/api-loading";
import "./globals.css";

export const metadata: Metadata = {
  title: "Falzo — Games for real friends",
  description:
    "Simple social games for dinners, road trips, and nights with friends.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body suppressHydrationWarning>
        <ApiLoadingProvider>{children}</ApiLoadingProvider>
      </body>
    </html>
  );
}
