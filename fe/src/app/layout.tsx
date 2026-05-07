import type { Metadata } from "next";
import { Manrope, Sora, Inter } from "next/font/google";
import { AppProviders } from "@/app/providers";
import "@/styles.css";

const manrope = Manrope({
  subsets: ["latin"],
  variable: "--font-manrope",
});

const sora = Sora({
  subsets: ["latin", "latin-ext"],
  variable: "--font-sora",
});

const inter = Inter({
  subsets: ["latin"],
  display: "swap",
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
      <body className={`${inter.className} ${sora.variable}`}>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
