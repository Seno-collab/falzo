import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Falzo — Games for real friends",
  description: "Simple social games for dinners, road trips, and nights with friends.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
