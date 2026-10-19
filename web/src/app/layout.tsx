import type { Metadata, Viewport } from "next";
import "./globals.css";
import { BottomNav } from "@/components/BottomNav";
import { Header } from "@/components/Header";

export const metadata: Metadata = {
  title: "Amiguis 🇨🇴",
  description: "La red social colombiana para compartir con tus amiguis.",
  manifest: "/manifest.json",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  themeColor: "#121417",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="es" className="dark" suppressHydrationWarning>
      <body className="bg-gray-50 dark:bg-[#121417] text-gray-900 dark:text-gray-100 antialiased pb-20 transition-colors">
        {/* Top Colombian Tricolor Accent Bar */}
        <div className="h-1.5 w-full flex sticky top-0 z-50">
          <div className="w-2/4 bg-[#FCD116]" />
          <div className="w-1/4 bg-[#003893]" />
          <div className="w-1/4 bg-[#CE1126]" />
        </div>

        {/* Sticky App Header with Pinky Promise Logo & Theme Switcher */}
        <Header />

        {/* Mobile-First Container (360px -> max-w-md) */}
        <main className="max-w-md mx-auto min-h-screen">{children}</main>

        <BottomNav />
      </body>
    </html>
  );
}
