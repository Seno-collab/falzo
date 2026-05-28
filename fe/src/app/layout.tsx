import type { Metadata } from "next";
import { AppProviders } from "@/app/providers";
import "@/styles.css";

export const metadata: Metadata = {
  title: "Visual Places",
  description:
    "Discover real places through community photos, explore them on the map, and save where you want to go next.",
  icons: {
    icon: [
      {
        url: "/icon.svg",
        type: "image/svg+xml",
      },
    ],
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body suppressHydrationWarning>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
