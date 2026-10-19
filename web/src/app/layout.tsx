import type { Metadata, Viewport } from "next";
import "./globals.css";
import { BottomNav } from "@/components/BottomNav";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Amiguis 🇨🇴 | Sin algoritmos, sin publicidad",
  description:
    "La primera red social colombiana. Sin algoritmos, sin publicidad. Ves lo que eliges ver.",
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
    <html lang="es">
      <body className="bg-[#121417] text-gray-100 antialiased pb-20">
        {/* Top Colombian Tricolor Accent Bar */}
        <div className="h-1.5 w-full flex sticky top-0 z-50">
          <div className="w-2/4 bg-[#FCD116]" />
          <div className="w-1/4 bg-[#003893]" />
          <div className="w-1/4 bg-[#CE1126]" />
        </div>

        {/* Sticky App Header */}
        <header className="sticky top-1.5 z-40 bg-[#121417]/95 backdrop-blur-md border-b border-gray-800/80">
          <div className="max-w-md mx-auto px-4 h-14 flex items-center justify-between">
            <Link href="/" className="flex items-center space-x-2">
              <span className="text-xl font-extrabold tracking-tight text-white">
                Amiguis
              </span>
              <span className="text-xs px-2 py-0.5 rounded-full bg-gray-800 text-[#FCD116] font-semibold border border-gray-700">
                100% Cronológico
              </span>
            </Link>
            <Link
              href="/login"
              className="text-xs font-semibold px-3 py-1.5 rounded-xl bg-gray-800 hover:bg-gray-700 text-gray-200 transition-colors"
            >
              Cuenta
            </Link>
          </div>
        </header>

        {/* Mobile-First Container (360px -> max-w-md) */}
        <main className="max-w-md mx-auto min-h-screen">{children}</main>

        <BottomNav />
      </body>
    </html>
  );
}
