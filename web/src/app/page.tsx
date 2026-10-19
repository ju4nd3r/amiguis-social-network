"use client";

import { useEffect, useState, useCallback } from "react";
import { api, Post, User } from "@/api/client";
import { FilterBar } from "@/components/FilterBar";
import { CreatePost } from "@/components/CreatePost";
import { PostCard } from "@/components/PostCard";
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
    body: "¡Qué alegría tener un espacio tranquilo para compartir con mis amiguis de verdad! 🇨🇴☕ #Amiguis",
    reactions_count: 24,
    comments_count: 2,
    comments: [
      {
        id: "c-demo-1",
        post_id: "demo-1",
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
        body: "¡Totalmente de acuerdo, Cami! Se siente súper bien estar aquí.",
        created_at: new Date(Date.now() - 1000 * 60 * 10).toISOString(),
      },
      {
        id: "c-demo-2",
        post_id: "demo-1",
        author: {
          id: "u-3",
          username: "sofia_cali",
          display_name: "Sofía Gómez",
          city: "Cali",
          followers_count: 64,
          following_count: 72,
          posts_count: 5,
          created_at: new Date().toISOString(),
        },
        body: "¡Un abrazo desde Cali, amigui! 💃✨",
        created_at: new Date(Date.now() - 1000 * 60 * 5).toISOString(),
      },
    ],
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
    body: "Tarde fría en Chapinero con un buen tinto de origen huilense. ¡Un saludo para todos mis amiguis!",
    reactions_count: 18,
    comments_count: 1,
    comments: [
      {
        id: "c-demo-3",
        post_id: "demo-2",
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
        body: "Nada como un buen café del Huila para el frío bogotano ☕",
        created_at: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
      },
    ],
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
    const syncUser = () => {
      if (typeof window !== "undefined") {
        const saved = localStorage.getItem("amiguis_user");
        if (saved) {
          try {
            setCurrentUser(JSON.parse(saved));
          } catch {
            setCurrentUser(null);
          }
        } else {
          setCurrentUser(null);
        }
      }
    };
    syncUser();
    window.addEventListener("amiguis:auth-change", syncUser);
    fetchFeed();
    return () => {
      window.removeEventListener("amiguis:auth-change", syncUser);
    };
  }, [fetchFeed]);

  const handlePostCreated = (newPost: Post) => {
    setPosts((prev) => [newPost, ...prev]);
  };

  const handlePostDeleted = (postId: string) => {
    setPosts((prev) => prev.filter((p) => p.id !== postId));
  };

  return (
    <div>
      {/* User-Controlled Filter Bar */}
      <FilterBar currentType={filterType} onSelectType={setFilterType} />

      <div className="p-4">
        {!currentUser && (
          <div className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-[#FCD116]/30 rounded-2xl p-4 mb-4 flex items-center justify-between shadow-sm transition-colors">
            <div>
              <h2 className="text-sm font-bold text-gray-900 dark:text-white">
                ¡Bienvenido a Amiguis! 🇨🇴
              </h2>
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
                className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-800 rounded-2xl p-4 h-36 animate-pulse"
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
            className="w-full touch-target py-3 bg-gray-200 dark:bg-gray-800 hover:bg-gray-300 dark:hover:bg-gray-700 text-xs font-semibold text-gray-800 dark:text-gray-200 rounded-xl mt-2 transition-colors"
          >
            Ver más publicaciones
          </button>
        )}
      </div>
    </div>
  );
}
