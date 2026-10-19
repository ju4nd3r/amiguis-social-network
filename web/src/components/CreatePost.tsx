"use client";

import { useState } from "react";
import { api, Post, MediaItem } from "@/api/client";
import { Image, Video, FileText, Send } from "lucide-react";

interface CreatePostProps {
  onPostCreated?: (post: Post) => void;
}

export function CreatePost({ onPostCreated }: CreatePostProps) {
  const [body, setBody] = useState("");
  const [type, setType] = useState<"text" | "photo" | "video">("text");
  const [mediaUrl, setMediaUrl] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!body.trim() && !mediaUrl.trim()) return;

    setSubmitting(true);
    setError(null);
    try {
      const media: MediaItem[] | undefined =
        type !== "text" && mediaUrl.trim()
          ? [
              {
                id: crypto.randomUUID(),
                storage_key: `${type}/external`,
                url: mediaUrl.trim(),
                kind: type === "video" ? "video" : "photo",
                status: "ready",
              },
            ]
          : undefined;

      const created = await api.createPost(type, body, media);
      setBody("");
      setMediaUrl("");
      setType("text");
      if (onPostCreated) onPostCreated(created);
    } catch (err: any) {
      setError(err.message || "Error al publicar. Verifica haber iniciado sesión.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="bg-[#1E2229] border border-gray-800/80 rounded-2xl p-4 mb-4 shadow-sm"
    >
      <div className="flex items-center space-x-2 mb-3">
        <button
          type="button"
          onClick={() => setType("text")}
          className={`touch-target px-3 py-1 rounded-full text-xs font-medium flex items-center space-x-1.5 transition-colors ${
            type === "text"
              ? "bg-[#FCD116] text-[#121417] font-semibold"
              : "bg-gray-800 text-gray-300 hover:bg-gray-700"
          }`}
        >
          <FileText className="w-3.5 h-3.5" />
          <span>Texto</span>
        </button>
        <button
          type="button"
          onClick={() => setType("photo")}
          className={`touch-target px-3 py-1 rounded-full text-xs font-medium flex items-center space-x-1.5 transition-colors ${
            type === "photo"
              ? "bg-[#003893] text-white font-semibold"
              : "bg-gray-800 text-gray-300 hover:bg-gray-700"
          }`}
        >
          <Image className="w-3.5 h-3.5" />
          <span>Foto</span>
        </button>
        <button
          type="button"
          onClick={() => setType("video")}
          className={`touch-target px-3 py-1 rounded-full text-xs font-medium flex items-center space-x-1.5 transition-colors ${
            type === "video"
              ? "bg-[#CE1126] text-white font-semibold"
              : "bg-gray-800 text-gray-300 hover:bg-gray-700"
          }`}
        >
          <Video className="w-3.5 h-3.5" />
          <span>Video</span>
        </button>
      </div>

      <textarea
        rows={3}
        maxLength={1000}
        value={body}
        onChange={(e) => setBody(e.target.value)}
        placeholder="¿Qué está pasando hoy, parce? Sin algoritmos que escondan tu voz..."
        className="w-full bg-gray-900/90 border border-gray-700/80 rounded-xl p-3 text-sm text-gray-100 placeholder-gray-500 focus:outline-none focus:border-[#FCD116] resize-none"
      />

      {type !== "text" && (
        <input
          type="url"
          value={mediaUrl}
          onChange={(e) => setMediaUrl(e.target.value)}
          placeholder={
            type === "photo"
              ? "URL de la foto (ej. https://images.unsplash.com/...)"
              : "URL del video MP4"
          }
          className="w-full mt-2 bg-gray-900/90 border border-gray-700/80 rounded-xl px-3 py-2 text-xs text-gray-100 placeholder-gray-500 focus:outline-none focus:border-[#FCD116]"
        />
      )}

      {error && <p className="text-xs text-red-400 mt-2">{error}</p>}

      <div className="flex items-center justify-between mt-3 pt-2 border-t border-gray-800/60">
        <span className="text-xs text-gray-500">{body.length} / 1000</span>
        <button
          type="submit"
          disabled={submitting || (!body.trim() && !mediaUrl.trim())}
          className="touch-target px-4 py-2 bg-[#FCD116] text-[#121417] rounded-xl text-xs font-bold hover:bg-amber-300 disabled:opacity-40 flex items-center space-x-1.5 transition-all"
        >
          <Send className="w-3.5 h-3.5" />
          <span>{submitting ? "Publicando..." : "Publicar"}</span>
        </button>
      </div>
    </form>
  );
}
