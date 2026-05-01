import type { Metadata } from "next";
import { Manrope, Sora } from "next/font/google";
import { AppProviders } from "@/app/providers";
import "@/styles.css";

const manrope = Manrope({
  subsets: ["latin", "vietnamese"],
  variable: "--font-manrope",
});

const sora = Sora({
  subsets: ["latin", "latin-ext"],
  variable: "--font-sora",
});

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
      <body className={`${manrope.variable} ${sora.variable}`}>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
