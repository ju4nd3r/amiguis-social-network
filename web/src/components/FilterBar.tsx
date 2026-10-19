"use client";

interface FilterBarProps {
  currentType: string;
  onSelectType: (type: string) => void;
}

export function FilterBar({ currentType, onSelectType }: FilterBarProps) {
  const filters = [
    { id: "all", label: "🇨🇴 Todo" },
    { id: "review", label: "⭐ Reseñas" },
    { id: "photo", label: "📸 Fotos" },
    { id: "video", label: "🎥 Videos" },
    { id: "text", label: "📝 Texto" },
  ];

  return (
    <div className="flex items-center space-x-2 overflow-x-auto py-2.5 no-scrollbar px-4 border-b border-gray-200 dark:border-gray-800/80 bg-white dark:bg-[#121417] transition-colors">
      <span className="text-xs text-gray-500 dark:text-gray-400 font-semibold uppercase tracking-wider pr-1">
        Filtro:
      </span>
      {filters.map((f) => {
        const active = currentType === f.id;
        return (
          <button
            key={f.id}
            onClick={() => onSelectType(f.id)}
            className={`touch-target px-3 py-1.5 rounded-full text-xs font-medium whitespace-nowrap transition-all ${
              active
                ? "bg-[#FCD116] text-[#121417] shadow-sm font-semibold"
                : "bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700"
            }`}
          >
            {f.label}
          </button>
        );
      })}
    </div>
  );
}
