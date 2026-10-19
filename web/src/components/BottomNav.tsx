"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, PlusSquare, Bell, User, ShieldCheck } from "lucide-react";

export function BottomNav() {
  const pathname = usePathname();

  const navItems = [
    { href: "/", label: "Feed", icon: Home },
    { href: "/create", label: "Publicar", icon: PlusSquare },
    { href: "/notifications", label: "Avisos", icon: Bell },
    { href: "/profile", label: "Perfil", icon: User },
    { href: "/privacy", label: "Privacidad", icon: ShieldCheck },
  ];

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 bg-white/95 dark:bg-[#121417]/95 backdrop-blur-md border-t border-gray-200 dark:border-gray-800 pb-safe transition-colors">
      <div className="max-w-md mx-auto flex items-center justify-around h-16 px-2">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`touch-target flex-col px-3 py-1 text-xs font-medium transition-colors ${
                isActive
                  ? "text-[#003893] dark:text-[#FCD116] font-bold"
                  : "text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200"
              }`}
            >
              <Icon className="w-5 h-5 mb-0.5" />
              <span>{item.label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
