"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Sun, Moon } from "lucide-react";
import { PinkyPromiseLogo } from "./PinkyPromiseLogo";
import { User } from "@/api/client";

export function Header() {
  const pathname = usePathname();
  const [darkMode, setDarkMode] = useState(true);
  const [currentUser, setCurrentUser] = useState<User | null>(null);

  const syncUserFromStorage = useCallback(() => {
    if (typeof window === "undefined") return;
    const raw = localStorage.getItem("amiguis_user");
    if (!raw) {
      setCurrentUser(null);
      return;
    }
    try {
      setCurrentUser(JSON.parse(raw));
    } catch {
      setCurrentUser(null);
    }
  }, []);

  useEffect(() => {
    const savedTheme = localStorage.getItem("amiguis_theme");
    if (savedTheme === "light") {
      setDarkMode(false);
      document.documentElement.classList.remove("dark");
    } else {
      setDarkMode(true);
      document.documentElement.classList.add("dark");
    }

    syncUserFromStorage();

    window.addEventListener("amiguis:auth-change", syncUserFromStorage);
    window.addEventListener("storage", syncUserFromStorage);
    return () => {
      window.removeEventListener("amiguis:auth-change", syncUserFromStorage);
      window.removeEventListener("storage", syncUserFromStorage);
    };
  }, [syncUserFromStorage]);

  useEffect(() => {
    syncUserFromStorage();
  }, [pathname, syncUserFromStorage]);

  const toggleTheme = () => {
    const nextDark = !darkMode;
    setDarkMode(nextDark);
    if (nextDark) {
      document.documentElement.classList.add("dark");
      localStorage.setItem("amiguis_theme", "dark");
    } else {
      document.documentElement.classList.remove("dark");
      localStorage.setItem("amiguis_theme", "light");
    }
  };

  const userDisplayName = currentUser
    ? currentUser.display_name || currentUser.username
    : null;

  return (
    <header className="sticky top-1.5 z-40 bg-white/95 dark:bg-[#121417]/95 backdrop-blur-md border-b border-gray-200 dark:border-gray-800/80 transition-colors">
      <div className="max-w-md mx-auto px-4 h-14 flex items-center justify-between">
        <Link href="/" className="flex items-center space-x-2.5 group">
          <PinkyPromiseLogo className="w-8 h-8 transition-transform group-hover:scale-105" />
          <span className="text-xl font-extrabold tracking-tight text-gray-900 dark:text-white">
            Amiguis
          </span>
        </Link>

        <div className="flex items-center space-x-2">
          <button
            type="button"
            onClick={toggleTheme}
            data-testid="theme-toggle"
            aria-label={darkMode ? "Cambiar a modo claro" : "Cambiar a modo oscuro"}
            className="touch-target p-2 rounded-xl bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-200 transition-colors"
          >
            {darkMode ? (
              <Sun className="w-4 h-4 text-[#FCD116]" />
            ) : (
              <Moon className="w-4 h-4 text-[#003893]" />
            )}
          </button>

          <Link
            href={currentUser ? "/profile" : "/login"}
            data-testid="header-account-button"
            className="text-xs font-semibold px-3 py-2 rounded-xl bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-800 dark:text-gray-200 transition-colors flex items-center space-x-1.5 max-w-[160px]"
          >
            {userDisplayName ? (
              <>
                <span className="w-5 h-5 rounded-full bg-gradient-to-tr from-[#003893] to-[#FCD116] text-white font-bold text-[10px] flex items-center justify-center shrink-0">
                  {userDisplayName.charAt(0).toUpperCase()}
                </span>
                <span className="truncate">{userDisplayName}</span>
              </>
            ) : (
              <span>Cuenta</span>
            )}
          </Link>
        </div>
      </div>
    </header>
  );
}
