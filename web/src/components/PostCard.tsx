"use client";

import { useState } from "react";
import { Post, Comment, api } from "@/api/client";
import { MessageCircle, Trash2, MapPin, Lightbulb } from "lucide-react";

interface PostCardProps {
  post: Post;
  currentUserId?: string;
  onPostDeleted?: (id: string) => void;
}

export function PostCard({ post, currentUserId, onPostDeleted }: PostCardProps) {
  const [reactionsCount, setReactionsCount] = useState(post.reactions_count || 0);
  const [myReaction, setMyReaction] = useState<string | undefined>(post.my_reaction);
  const [showComments, setShowComments] = useState(false);
  const initialComments = post.comments || [];
  const [comments, setComments] = useState<Comment[]>(initialComments);
  const [commentsCount, setCommentsCount] = useState(
    initialComments.length > 0 ? initialComments.length : post.comments_count || 0
  );
  const [commentText, setCommentText] = useState("");
  const [loadingComments, setLoadingComments] = useState(false);

  const toggleInteresante = async () => {
    const wasActive = myReaction === "interesante";
    if (wasActive) {
      setMyReaction(undefined);
      setReactionsCount((prev) => Math.max(0, prev - 1));
      try {
        await api.removeReaction(post.id);
      } catch {
        // Optimistic update preserved for demo/offline posts
      }
    } else {
      setMyReaction("interesante");
      setReactionsCount((prev) => prev + 1);
      try {
        await api.addReaction(post.id, "interesante");
      } catch {
        // Optimistic update preserved for demo/offline posts
      }
    }
  };

  const toggleCommentsDrawer = async () => {
    const nextState = !showComments;
    setShowComments(nextState);
    if (nextState) {
      setLoadingComments(true);
      try {
        const list = await api.getComments(post.id);
        if (Array.isArray(list) && list.length > 0) {
          setComments((prev) => {
            const existingIds = new Set(prev.map((c) => c.id));
            const merged = [...prev];
            for (const item of list) {
              if (!existingIds.has(item.id)) {
                merged.push(item);
              }
            }
            setCommentsCount(merged.length);
            return merged;
          });
        }
      } catch {
        // Keep existing comments if backend is offline
      } finally {
        setLoadingComments(false);
      }
    }
  };

  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault();
    const cleanText = commentText.trim();
    if (!cleanText) return;

    setCommentText("");

    let currentUsername = "amigui";
    let currentDisplayName = "Amigui";
    if (typeof window !== "undefined") {
      const saved = localStorage.getItem("amiguis_user");
      if (saved) {
        try {
          const parsed = JSON.parse(saved);
          if (parsed.username) currentUsername = parsed.username;
          if (parsed.display_name) currentDisplayName = parsed.display_name;
        } catch {}
      }
    }

    try {
      const newComment = await api.addComment(post.id, cleanText);
      if (newComment && newComment.id) {
        setComments((prev) => [...prev, newComment]);
        setCommentsCount((prev) => prev + 1);
        return;
      }
    } catch {
      // Fallback to local comment if backend is offline or demo post
    }

    const fallbackComment: Comment = {
      id: `local-${Date.now()}`,
      post_id: post.id,
      author: {
        id: currentUserId || "local-user",
        username: currentUsername,
        display_name: currentDisplayName,
        followers_count: 0,
        following_count: 0,
        posts_count: 0,
        created_at: new Date().toISOString(),
      },
      body: cleanText,
      created_at: new Date().toISOString(),
    };
    setComments((prev) => [...prev, fallbackComment]);
    setCommentsCount((prev) => prev + 1);
  };

  const handleDeletePost = async () => {
    if (!confirm("¿Deseas eliminar esta publicación?")) return;
    try {
      await api.deletePost(post.id);
    } catch {}
    if (onPostDeleted) onPostDeleted(post.id);
  };

  const authorName = post.author?.display_name || "Amigui";
  const authorUsername = post.author?.username || "amigui";

  return (
    <article className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-800/80 rounded-2xl p-4 mb-4 shadow-sm transition-colors">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 rounded-full bg-gradient-to-tr from-[#003893] to-[#FCD116] flex items-center justify-center font-bold text-white text-sm shadow">
            {authorName.charAt(0).toUpperCase()}
          </div>
          <div>
            <div className="flex items-center space-x-1.5">
              <span className="font-semibold text-sm text-gray-900 dark:text-gray-100">
                {authorName}
              </span>
              <span className="text-xs text-gray-500 dark:text-gray-400">
                @{authorUsername}
              </span>
            </div>
            {post.author?.city && (
              <div className="flex items-center text-xs text-amber-600 dark:text-amber-400 font-medium mt-0.5">
                <MapPin className="w-3 h-3 mr-0.5" />
                <span>{post.author.city}</span>
              </div>
            )}
          </div>
        </div>

        {currentUserId && currentUserId === post.author?.id && (
          <button
            onClick={handleDeletePost}
            className="text-gray-400 hover:text-red-500 p-1.5 rounded-lg transition-colors"
            title="Eliminar publicación"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Body */}
      <p className="text-sm text-gray-800 dark:text-gray-200 leading-relaxed whitespace-pre-line mb-3">
        {post.body}
      </p>

      {/* Media if present */}
      {post.media && post.media.length > 0 && (
        <div className="mb-3 rounded-xl overflow-hidden border border-gray-200 dark:border-gray-800">
          {post.media[0].kind === "video" ? (
            <video
              src={post.media[0].url}
              controls
              className="w-full max-h-96 object-cover"
            />
          ) : (
            <img
              src={post.media[0].url}
              alt="Contenido multimedia"
              className="w-full max-h-96 object-cover"
            />
          )}
        </div>
      )}

      {/* Single Reaction ("Interesante") & Comments Bar */}
      <div className="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800/80">
        <button
          type="button"
          onClick={toggleInteresante}
          data-testid={`react-interesante-${post.id}`}
          className={`touch-target px-3.5 py-1.5 rounded-full text-xs font-semibold flex items-center space-x-1.5 transition-all ${
            myReaction === "interesante"
              ? "bg-amber-500/20 text-amber-600 dark:text-[#FCD116] border border-amber-500/40"
              : "text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800"
          }`}
        >
          <Lightbulb className="w-4 h-4" />
          <span>Interesante</span>
          {reactionsCount > 0 && (
            <span className="ml-1 px-1.5 py-0.2 rounded-full bg-gray-200/70 dark:bg-gray-700/70 text-[11px]">
              {reactionsCount}
            </span>
          )}
        </button>

        <button
          type="button"
          onClick={toggleCommentsDrawer}
          data-testid={`comment-toggle-${post.id}`}
          aria-label="Comentarios"
          className="touch-target px-3 py-1.5 rounded-full text-xs font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 flex items-center space-x-1.5 transition-colors"
        >
          <MessageCircle className="w-4 h-4" />
          <span>Comentarios ({commentsCount})</span>
        </button>
      </div>

      {/* Comments section */}
      {showComments && (
        <div
          data-testid={`comments-section-${post.id}`}
          className="mt-3 pt-3 border-t border-gray-100 dark:border-gray-800/60"
        >
          {loadingComments ? (
            <p className="text-xs text-gray-500 py-2">Cargando comentarios...</p>
          ) : (
            <div className="space-y-2 mb-3">
              {comments.length === 0 ? (
                <p className="text-xs text-gray-500 italic">
                  No hay comentarios aún. Sé el primero en comentar.
                </p>
              ) : (
                comments.map((c) => (
                  <div
                    key={c.id}
                    className="bg-gray-100 dark:bg-gray-800/60 p-2.5 rounded-xl text-xs"
                  >
                    <span className="font-semibold text-gray-900 dark:text-gray-200">
                      @{c.author?.username || "amigui"}
                    </span>
                    :{" "}
                    <span className="text-gray-700 dark:text-gray-300">
                      {c.body}
                    </span>
                  </div>
                ))
              )}
            </div>
          )}

          <form onSubmit={handleAddComment} className="flex space-x-2">
            <input
              type="text"
              value={commentText}
              onChange={(e) => setCommentText(e.target.value)}
              placeholder="Escribe un comentario..."
              className="flex-1 bg-gray-50 dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-xl px-3 py-2 text-xs text-gray-900 dark:text-gray-100 focus:outline-none focus:border-[#FCD116]"
            />
            <button
              type="submit"
              className="touch-target px-3.5 py-1.5 bg-[#003893] text-white rounded-xl text-xs font-semibold hover:bg-blue-800 transition-colors"
            >
              Enviar
            </button>
          </form>
        </div>
      )}
    </article>
  );
}
