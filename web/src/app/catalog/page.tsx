"use client";

import { useEffect, useState, useCallback } from "react";
import { CatalogItem, Post, api } from "@/api/client";
import { getCategoryLabel } from "@/data/catalog";
import { CatalogItemModal } from "@/components/CatalogItemModal";
import { useRouter } from "next/navigation";
import {
  Search,
  Star,
  Film,
  Tv,
  BookOpen,
  Sparkles,
  Users,
} from "lucide-react";

export default function CatalogPage() {
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState<string>("all");
  const [items, setItems] = useState<CatalogItem[]>([]);
  const [selectedItem, setSelectedItem] = useState<CatalogItem | null>(null);
  const [loading, setLoading] = useState(true);

  const loadCatalog = useCallback(async () => {
    setLoading(true);
    try {
      const results = await api.searchCatalog(query, category);
      setItems(results);
    } finally {
      setLoading(false);
    }
  }, [query, category]);

  useEffect(() => {
    loadCatalog();
  }, [loadCatalog]);

  const handleReviewSubmitted = (_post: Post, updatedItem: CatalogItem) => {
    setItems((prev) =>
      prev.map((it) => (it.id === updatedItem.id ? updatedItem : it))
    );
    setSelectedItem(updatedItem);
  };

  const categories = [
    { id: "all", label: "Todo", icon: Sparkles },
    { id: "movie", label: "Películas", icon: Film },
    { id: "series", label: "Series", icon: Tv },
    { id: "book", label: "Libros", icon: BookOpen },
  ];

  return (
    <div className="p-4 space-y-4">
      {selectedItem && (
        <CatalogItemModal
          item={selectedItem}
          onClose={() => setSelectedItem(null)}
          onReviewSubmitted={(post, updatedItem) => {
            handleReviewSubmitted(post, updatedItem);
            setTimeout(() => {
              setSelectedItem(null);
              router.push("/");
            }, 900);
          }}
        />
      )}

      {/* Header & Search Bar */}
      <div className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-800 rounded-2xl p-4 shadow-sm space-y-3 transition-colors">
        <div>
          <h1 className="text-base font-extrabold text-gray-900 dark:text-white">
            Películas, Series y Libros
          </h1>
          <p className="text-xs text-gray-500 dark:text-gray-400">
            Busca una obra, consulta su ficha técnica, márcala como vista/visto y comparte tu reseña en el feed.
          </p>
        </div>

        {/* Search Input */}
        <div className="relative">
          <Search className="w-4 h-4 text-gray-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            data-testid="catalog-search-input"
            placeholder="Buscar por título, elenco, director, autor o productora..."
            className="w-full bg-gray-50 dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-xl pl-10 pr-3 py-2.5 text-xs text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-[#FCD116]"
          />
        </div>

        {/* Category Tabs */}
        <div className="flex items-center space-x-1.5 overflow-x-auto no-scrollbar pt-1">
          {categories.map((cat) => {
            const Icon = cat.icon;
            const active = category === cat.id;
            return (
              <button
                key={cat.id}
                type="button"
                onClick={() => setCategory(cat.id)}
                data-testid={`catalog-filter-${cat.id}`}
                className={`touch-target px-3 py-1.5 rounded-xl text-xs font-bold flex items-center space-x-1.5 whitespace-nowrap transition-all ${
                  active
                    ? "bg-[#FCD116] text-[#121417] shadow-sm"
                    : "bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700"
                }`}
              >
                <Icon className="w-3.5 h-3.5" />
                <span>{cat.label}</span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Results List */}
      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((n) => (
            <div
              key={n}
              className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-800 rounded-2xl p-4 h-28 animate-pulse"
            />
          ))}
        </div>
      ) : items.length === 0 ? (
        <div className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-800 rounded-2xl p-8 text-center text-xs text-gray-500">
          No encontramos películas, series ni libros con esa búsqueda.
        </div>
      ) : (
        <div className="space-y-3">
          {items.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => setSelectedItem(item)}
              data-testid={`catalog-item-${item.id}`}
              className="w-full text-left bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-800 hover:border-[#FCD116] dark:hover:border-[#FCD116] rounded-2xl p-3.5 shadow-sm flex space-x-3.5 transition-all group"
            >
              <img
                src={item.poster_url}
                alt={item.title}
                className="w-16 h-24 object-cover rounded-xl shadow shrink-0 border border-gray-200 dark:border-gray-700 group-hover:scale-[1.02] transition-transform"
              />

              <div className="flex-1 min-w-0 space-y-1">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-extrabold uppercase tracking-wider px-2 py-0.5 rounded-full bg-[#003893]/15 dark:bg-[#FCD116]/15 text-[#003893] dark:text-[#FCD116]">
                    {getCategoryLabel(item.category)} • {item.year}
                  </span>

                  <span className="inline-flex items-center space-x-1 text-xs font-extrabold text-gray-900 dark:text-white">
                    <Star className="w-3.5 h-3.5 text-[#FCD116] fill-[#FCD116]" />
                    <span>{item.average_rating.toFixed(1)}</span>
                  </span>
                </div>

                <h2 className="text-sm font-extrabold text-gray-900 dark:text-white truncate">
                  {item.title}
                </h2>

                <p className="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 leading-relaxed">
                  {item.description}
                </p>

                <div className="flex items-center text-[11px] text-gray-500 dark:text-gray-400 pt-0.5 truncate">
                  <Users className="w-3 h-3 mr-1 shrink-0 text-amber-600 dark:text-[#FCD116]" />
                  <span className="truncate">{item.cast.join(", ")}</span>
                </div>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
