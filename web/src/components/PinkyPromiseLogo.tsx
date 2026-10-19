"use client";

export function PinkyPromiseLogo({ className = "w-8 h-8" }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      aria-label="Amiguis Pinky Promise Logo"
    >
      {/* Minimalist circular frame with subtle warm accent */}
      <circle
        cx="20"
        cy="20"
        r="18"
        className="stroke-amber-500 dark:stroke-[#FCD116]"
        strokeWidth="2"
      />
      {/* Left hand & wrist */}
      <path
        d="M6 25L11 21C12.2 20 13.8 20 15 21L16.5 22.5"
        className="stroke-gray-900 dark:stroke-white"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {/* Left curled fingers */}
      <path
        d="M12.5 19.5C13.5 18 15.5 18 16.5 19.5"
        className="stroke-gray-900 dark:stroke-white"
        strokeWidth="2"
        strokeLinecap="round"
      />
      {/* Right hand & wrist */}
      <path
        d="M34 25L29 21C27.8 20 26.2 20 25 21L23.5 22.5"
        className="stroke-gray-900 dark:stroke-white"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {/* Right curled fingers */}
      <path
        d="M27.5 19.5C26.5 18 24.5 18 23.5 19.5"
        className="stroke-gray-900 dark:stroke-white"
        strokeWidth="2"
        strokeLinecap="round"
      />
      {/* Interlocked Pinky Promise loop in center */}
      <path
        d="M15.5 22C15.5 16.5 20 15.5 20 19.5C20 23 24.5 22 24.5 16.5"
        className="stroke-[#003893] dark:stroke-[#FCD116]"
        strokeWidth="2.3"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      {/* Minimalist spark of friendship above the pinky lock */}
      <path
        d="M20 9V11.5M16.5 11L18 12.5M23.5 11L22 12.5"
        className="stroke-[#CE1126] dark:stroke-[#FCD116]"
        strokeWidth="1.8"
        strokeLinecap="round"
      />
    </svg>
  );
}
