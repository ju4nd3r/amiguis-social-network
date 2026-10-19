"use client";

import { useState } from "react";
import { Post, Comment, api } from "@/api/client";
import { MessageCircle, Trash2, MapPin } from "lucide-react";

interface PostCardProps {
  post: Post;
  currentUserId?: string;
  onPostDeleted?: (id: string) => void;
}

export function PostCard({ post, currentUserId, onPostDeleted }: PostCardProps) {
  const [reactionsCount, setReactionsCount] = useState(post.reactions_count);
  const [myReaction, setMyReaction] = useState(post.my_reaction);
  const [showComments, setShowComments] = useState(false);
  const [comments, setComments] = useState<Comment[]>([]);
  const [commentText, setCommentText] = useState("");
  const [loadingComments, setLoadingComments] = useState(false);

  const toggleReaction = async (type: "like" | "love" | "arepa" | "parce" | "genius") => {
    try {
      if (myReaction === type) {
        await api.removeReaction(post.id);
        setMyReaction(undefined);
        setReactionsCount((prev) => Math.max(0, prev - 1));
      } else {
        await api.addReaction(post.id, type);
        if (!myReaction) {
          setReactionsCount((prev) => prev + 1);
        }
        setMyReaction(type);
      }
    } catch (err) {
      console.error("Error al reaccionar", err);
    }
  };

  const loadComments = async () => {
    if (!showComments) {
      setLoadingComments(true);
      try {
        const list = await api.getComments(post.id);
        setComments(list);
      } catch (err) {
        console.error("Error al cargar comentarios", err);
      } finally {
        setLoadingComments(false);
      }
    }
    setShowComments(!showComments);
  };

  const handleAddComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commentText.trim()) return;

    try {
      const newComment = await api.addComment(post.id, commentText);
      setComments((prev) => [...prev, newComment]);
      setCommentText("");
    } catch (err) {
      alert("Error al enviar comentario. Inicia sesión si no lo has hecho.");
    }
  };

  const handleDeletePost = async () => {
    if (!confirm("¿Deseas eliminar esta publicación?")) return;
    try {
      await api.deletePost(post.id);
      if (onPostDeleted) onPostDeleted(post.id);
    } catch (err) {
      alert("No se pudo eliminar la publicación");
    }
  };

  return (
    <article className="bg-[#1E2229] border border-gray-800/80 rounded-2xl p-4 mb-4 shadow-sm">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 rounded-full bg-gradient-to-tr from-[#003893] to-[#FCD116] flex items-center justify-center font-bold text-white text-sm shadow">
            {post.author.display_name?.charAt(0).toUpperCase() || "A"}
          </div>
          <div>
            <div className="flex items-center space-x-1.5">
              <span className="font-semibold text-sm text-gray-100">{post.author.display_name}</span>
              <span className="text-xs text-gray-400">@{post.author.username}</span>
            </div>
            {post.author.city && (
              <div className="flex items-center text-xs text-amber-400 font-medium mt-0.5">
                <MapPin className="w-3 h-3 mr-0.5" />
                <span>{post.author.city}</span>
              </div>
            )}
          </div>
        </div>

        {currentUserId === post.author.id && (
          <button
            onClick={handleDeletePost}
            className="text-gray-500 hover:text-red-400 p-1.5 rounded-lg transition-colors"
            title="Eliminar publicación"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Body */}
      <p className="text-sm text-gray-200 leading-relaxed whitespace-pre-line mb-3">
        {post.body}
      </p>

      {/* Media if present */}
      {post.media && post.media.length > 0 && (
        <div className="mb-3 rounded-xl overflow-hidden border border-gray-800">
          {post.media[0].kind === "video" ? (
            <video src={post.media[0].url} controls className="w-full max-h-96 object-cover" />
          ) : (
            <img src={post.media[0].url} alt="Contenido multimedia" className="w-full max-h-96 object-cover" />
          )}
        </div>
      )}

      {/* Reactions Bar (Colombian themed: Me gusta, Parce, Arepa, Genial) */}
      <div className="flex items-center justify-between pt-2 border-t border-gray-800/80">
        <div className="flex items-center space-x-1.5">
          <button
            onClick={() => toggleReaction("like")}
            className={`touch-target px-2.5 py-1 rounded-full text-xs font-medium transition-all ${
              myReaction === "like" ? "bg-red-500/20 text-red-400 border border-red-500/30" : "text-gray-400 hover:bg-gray-800"
            }`}
          >
            ❤️ <span className="ml-1">Me gusta</span>
          </button>

          <button
            onClick={() => toggleReaction("parce")}
            className={`touch-target px-2.5 py-1 rounded-full text-xs font-medium transition-all ${
              myReaction === "parce" ? "bg-blue-500/20 text-blue-400 border border-blue-500/30" : "text-gray-400 hover:bg-gray-800"
            }`}
          >
            🤝 <span className="ml-1">Parce</span>
          </button>

          <button
            onClick={() => toggleReaction("arepa")}
            className={`touch-target px-2.5 py-1 rounded-full text-xs font-medium transition-all ${
              myReaction === "arepa" ? "bg-amber-500/20 text-amber-300 border border-amber-500/30" : "text-gray-400 hover:bg-gray-800"
            }`}
          >
            🫓 <span className="ml-1">Arepa</span>
          </button>
        </div>

        <button
          onClick={loadComments}
          className="touch-target px-2.5 py-1 text-xs text-gray-400 hover:text-gray-200 flex items-center space-x-1"
        >
          <MessageCircle className="w-4 h-4" />
          <span>{post.comments_count || comments.length}</span>
        </button>
      </div>

      {/* Comments section */}
      {showComments && (
        <div className="mt-3 pt-3 border-t border-gray-800/60">
          {loadingComments ? (
            <p className="text-xs text-gray-500 py-2">Cargando comentarios...</p>
          ) : (
            <div className="space-y-2 mb-3">
              {comments.length === 0 ? (
                <p className="text-xs text-gray-500 italic">No hay comentarios aún. Sé el primero.</p>
              ) : (
                comments.map((c) => (
                  <div key={c.id} className="bg-gray-800/50 p-2.5 rounded-xl text-xs">
                    <span className="font-semibold text-gray-200">@{c.author.username}</span>:{" "}
                    <span className="text-gray-300">{c.body}</span>
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
              className="flex-1 bg-gray-900 border border-gray-700 rounded-xl px-3 py-2 text-xs text-gray-100 focus:outline-none focus:border-[#FCD116]"
            />
            <button
              type="submit"
              className="touch-target px-3 py-1.5 bg-[#003893] text-white rounded-xl text-xs font-semibold hover:bg-blue-800"
            >
              Enviar
            </button>
          </form>
        </div>
      )}
    </article>
  );
}
