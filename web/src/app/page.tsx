"use client";

import { useEffect, useState, useCallback } from "react";
import { api, Post, User } from "@/api/client";
import { FilterBar } from "@/components/FilterBar";
import { CreatePost } from "@/components/CreatePost";
import { PostCard } from "@/components/PostCard";
import { RefreshCw, Sparkles } from "lucide-react";
import Link from "next/link";

const DEMO_POSTS: Post[] = [
  {
    id: "demo-1",
    author: {
      id: "u-1",
      username: "camila_medallo",
      display_name: "Camila Restrepo",
      city: "Medellín",
      followers_count: 142,
      following_count: 89,
      posts_count: 12,
      created_at: new Date().toISOString(),
    },
    type: "text",
    body: "¡Qué chimba tener una red social donde veo las publicaciones de mis amigos en orden real y sin publicidad! 🇨🇴☕ #Amiguis #SinAlgoritmo",
    reactions_count: 24,
    comments_count: 3,
    created_at: new Date(Date.now() - 1000 * 60 * 15).toISOString(),
  },
  {
    id: "demo-2",
    author: {
      id: "u-2",
      username: "andres_bogota",
      display_name: "Andrés Ospina",
      city: "Bogotá",
      followers_count: 95,
      following_count: 110,
      posts_count: 8,
      created_at: new Date().toISOString(),
    },
    type: "text",
    body: "Tarde fría en Chapinero con un buen tinto de origen huilense. Recuerden que aquí nuestros datos están protegidos bajo la Ley 1581 de Habeas Data.",
    reactions_count: 18,
    comments_count: 1,
    created_at: new Date(Date.now() - 1000 * 60 * 45).toISOString(),
  },
];

export default function FeedPage() {
  const [filterType, setFilterType] = useState("all");
  const [posts, setPosts] = useState<Post[]>([]);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [nextCursor, setNextCursor] = useState<string | undefined>(undefined);

  const fetchFeed = useCallback(
    async (cursor = "", append = false) => {
      setLoading(true);
      try {
        const res = await api.getFeed(cursor, filterType);
        if (res.items && res.items.length > 0) {
          setPosts((prev) => (append ? [...prev, ...res.items] : res.items));
          setNextCursor(res.next_cursor);
        } else if (!append) {
          setPosts(
            DEMO_POSTS.filter(
              (p) => filterType === "all" || p.type === filterType
            )
          );
        }
      } catch {
        if (!append) {
          setPosts(
            DEMO_POSTS.filter(
              (p) => filterType === "all" || p.type === filterType
            )
          );
        }
      } finally {
        setLoading(false);
      }
    },
    [filterType]
  );

  useEffect(() => {
    if (typeof window !== "undefined") {
      const saved = localStorage.getItem("amiguis_user");
      if (saved) {
        try {
          setCurrentUser(JSON.parse(saved));
        } catch {}
      }
    }
    fetchFeed();
  }, [fetchFeed]);

  const handlePostCreated = (newPost: Post) => {
    setPosts((prev) => [newPost, ...prev]);
  };

  const handlePostDeleted = (postId: string) => {
    setPosts((prev) => prev.filter((p) => p.id !== postId));
  };

  return (
    <div>
      {/* Manifesto Banner */}
      <div className="bg-gradient-to-r from-[#003893]/40 via-[#1E2229] to-[#CE1126]/30 px-4 py-3 border-b border-gray-800 flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <Sparkles className="w-4 h-4 text-[#FCD116] shrink-0" />
          <p className="text-xs text-gray-300">
            <strong>Sin algoritmos ni anuncios.</strong> Orden 100% cronológico:{" "}
            <code className="text-[#FCD116]">created_at DESC</code>
          </p>
        </div>
        <button
          onClick={() => fetchFeed()}
          className="p-1.5 text-gray-400 hover:text-white rounded-lg"
          title="Actualizar feed"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
        </button>
      </div>

      {/* User-Controlled Filter Bar */}
      <FilterBar currentType={filterType} onSelectType={setFilterType} />

      <div className="p-4">
        {!currentUser && (
          <div className="bg-[#1E2229] border border-[#FCD116]/30 rounded-2xl p-4 mb-4 flex items-center justify-between">
            <div>
              <h2 className="text-sm font-bold text-white">
                ¡Bienvenido a Amiguis! 🇨🇴
              </h2>
              <p className="text-xs text-gray-400 mt-0.5">
                Únete para publicar, seguir parceros y armar tus círculos.
              </p>
            </div>
            <Link
              href="/login"
              className="touch-target px-4 py-2 bg-[#FCD116] text-[#121417] font-bold text-xs rounded-xl shrink-0"
            >
              Entrar
            </Link>
          </div>
        )}

        <CreatePost onPostCreated={handlePostCreated} />

        {/* Feed Timeline */}
        {loading && posts.length === 0 ? (
          <div className="space-y-4">
            {[1, 2].map((n) => (
              <div
                key={n}
                className="bg-[#1E2229] border border-gray-800 rounded-2xl p-4 h-36 animate-pulse"
              />
            ))}
          </div>
        ) : posts.length === 0 ? (
          <div className="text-center py-12 text-gray-500 text-sm">
            No hay publicaciones en este filtro todavía.
          </div>
        ) : (
          posts.map((post) => (
            <PostCard
              key={post.id}
              post={post}
              currentUserId={currentUser?.id}
              onPostDeleted={handlePostDeleted}
            />
          ))
        )}

        {nextCursor && (
          <button
            onClick={() => fetchFeed(nextCursor, true)}
            className="w-full touch-target py-3 bg-gray-800 hover:bg-gray-700 text-xs font-semibold text-gray-200 rounded-xl mt-2"
          >
            Cargar más publicaciones antiguas
          </button>
        )}
      </div>
    </div>
  );
}
