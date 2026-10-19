"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Sun, Moon } from "lucide-react";
import { PinkyPromiseLogo } from "./PinkyPromiseLogo";

export function Header() {
  const [darkMode, setDarkMode] = useState(true);

  useEffect(() => {
    const savedTheme = localStorage.getItem("amiguis_theme");
    if (savedTheme === "light") {
      setDarkMode(false);
      document.documentElement.classList.remove("dark");
    } else {
      setDarkMode(true);
      document.documentElement.classList.add("dark");
    }
  }, []);

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
            href="/login"
            className="text-xs font-semibold px-3 py-2 rounded-xl bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-800 dark:text-gray-200 transition-colors"
          >
            Cuenta
          </Link>
        </div>
      </div>
    </header>
  );
}
