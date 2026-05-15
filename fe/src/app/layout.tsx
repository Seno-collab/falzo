import type { Metadata } from "next";
import { AppProviders } from "@/app/providers";
import "@/styles.css";

export const metadata: Metadata = {
  title: "Travel Discovery",
  description:
    "Discover places, explore on map, and save your dream destinations.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
