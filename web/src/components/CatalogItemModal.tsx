"use client";

import { useState } from "react";
import { CatalogItem, Post, api } from "@/api/client";
import { getCategoryLabel } from "@/data/catalog";
import {
  X,
  Star,
  Eye,
  Heart,
  Film,
  Tv,
  BookOpen,
  Users,
  Clapperboard,
  Clock,
  Globe2,
  CheckCircle2,
} from "lucide-react";

interface CatalogItemModalProps {
  item: CatalogItem;
  onClose: () => void;
  onReviewSubmitted?: (post: Post, updatedItem: CatalogItem) => void;
}

export function CatalogItemModal({
  item: initialItem,
  onClose,
  onReviewSubmitted,
}: CatalogItemModalProps) {
  const [item, setItem] = useState<CatalogItem>(initialItem);
  const [watched, setWatched] = useState(true);
  const [liked, setLiked] = useState(false);
  const [rating, setRating] = useState<number>(5);
  const [reviewBody, setReviewBody] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [publishedSuccess, setPublishedSuccess] = useState(false);

  const handleSubmitReview = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      const res = await api.submitCatalogReview(item.id, {
        body: reviewBody,
        rating,
        watched,
        liked,
      });
      setItem(res.item);
      setPublishedSuccess(true);
      if (onReviewSubmitted) {
        onReviewSubmitted(res.post, res.item);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const CategoryIcon =
    item.category === "movie"
      ? Film
      : item.category === "series"
      ? Tv
      : BookOpen;

  const creatorLabel =
    item.category === "book"
      ? "Autor(a)"
      : item.category === "series"
      ? "Creador(es)"
      : "Dirección";

  const producerLabel =
    item.category === "book" ? "Editorial / Productor" : "Productora";

  const castLabel =
    item.category === "book"
      ? "Reparto / Personajes principales"
      : "Reparto / Casting";

  return (
    <div
      role="dialog"
      aria-modal="true"
      data-testid="catalog-detail-modal"
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-3 overflow-y-auto"
    >
      <div className="bg-white dark:bg-[#1E2229] border border-gray-200 dark:border-gray-700 rounded-2xl max-w-md w-full max-h-[90vh] overflow-y-auto shadow-2xl transition-colors my-auto">
        {/* Top Banner with Poster & Core Metadata */}
        <div className="relative p-4 flex space-x-4 border-b border-gray-200 dark:border-gray-800">
          <button
            type="button"
            onClick={onClose}
            data-testid="close-catalog-modal"
            aria-label="Cerrar detalle"
            className="absolute top-3 right-3 p-1.5 rounded-full bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
          >
            <X className="w-4 h-4" />
          </button>

          <img
            src={item.poster_url}
            alt={item.title}
            className="w-24 h-36 object-cover rounded-xl shadow-md shrink-0 border border-gray-200 dark:border-gray-700"
          />

          <div className="flex-1 pr-6 space-y-1.5">
            <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-[11px] font-bold bg-[#003893]/15 dark:bg-[#FCD116]/15 text-[#003893] dark:text-[#FCD116]">
              <CategoryIcon className="w-3 h-3 mr-1" />
              <span>{getCategoryLabel(item.category)}</span>
            </span>

            <h2
              data-testid="detail-title"
              className="text-base font-extrabold text-gray-900 dark:text-white leading-snug"
            >
              {item.title}
            </h2>

            <div className="flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
              <span data-testid="detail-year" className="font-semibold text-gray-800 dark:text-gray-200">
                Año: {item.year}
              </span>
              <span>•</span>
              <span className="inline-flex items-center">
                <Clock className="w-3 h-3 mr-1" />
                {item.duration}
              </span>
            </div>

            {/* Average Rating */}
            <div
              data-testid="detail-average-rating"
              className="inline-flex items-center space-x-1.5 bg-amber-500/10 border border-amber-500/30 px-2.5 py-1 rounded-xl"
            >
              <Star className="w-4 h-4 text-[#FCD116] fill-[#FCD116]" />
              <span className="text-xs font-extrabold text-gray-900 dark:text-white">
                {item.average_rating.toFixed(1)} / 5.0
              </span>
              <span className="text-[11px] text-gray-500 dark:text-gray-400">
                ({item.ratings_count} votos)
              </span>
            </div>

            {/* Genres */}
            <div className="flex flex-wrap gap-1 pt-1">
              {item.genres.map((g) => (
                <span
                  key={g}
                  className="px-2 py-0.5 rounded-md bg-gray-100 dark:bg-gray-800 text-[10px] font-medium text-gray-600 dark:text-gray-300"
                >
                  {g}
                </span>
              ))}
            </div>
          </div>
        </div>

        {/* Detailed Information Section */}
        <div className="p-4 space-y-3.5 text-xs border-b border-gray-200 dark:border-gray-800">
          <div>
            <h3 className="font-bold text-gray-900 dark:text-white mb-1">
              Descripción / Sinopsis
            </h3>
            <p
              data-testid="detail-description"
              className="text-gray-600 dark:text-gray-300 leading-relaxed"
            >
              {item.description}
            </p>
          </div>

          <div className="grid grid-cols-1 gap-2.5 bg-gray-50 dark:bg-gray-900/60 p-3 rounded-xl border border-gray-200/70 dark:border-gray-800">
            <div className="flex items-start space-x-2">
              <Users className="w-4 h-4 text-amber-600 dark:text-[#FCD116] shrink-0 mt-0.5" />
              <div>
                <span className="font-bold text-gray-800 dark:text-gray-200 block">
                  {castLabel}:
                </span>
                <span
                  data-testid="detail-cast"
                  className="text-gray-600 dark:text-gray-300"
                >
                  {item.cast.join(", ")}
                </span>
              </div>
            </div>

            <div className="flex items-start space-x-2">
              <Clapperboard className="w-4 h-4 text-amber-600 dark:text-[#FCD116] shrink-0 mt-0.5" />
              <div>
                <span className="font-bold text-gray-800 dark:text-gray-200 block">
                  {producerLabel}:
                </span>
                <span
                  data-testid="detail-producer"
                  className="text-gray-600 dark:text-gray-300"
                >
                  {item.producer}
                </span>
              </div>
            </div>

            <div className="flex items-center justify-between pt-1 border-t border-gray-200/60 dark:border-gray-800 text-[11px] text-gray-500 dark:text-gray-400">
              <span>
                <strong className="text-gray-800 dark:text-gray-200">
                  {creatorLabel}:
                </strong>{" "}
                {item.creator}
              </span>
              <span className="inline-flex items-center">
                <Globe2 className="w-3 h-3 mr-1" />
                {item.country}
              </span>
            </div>
          </div>
        </div>

        {/* User Review, Vista/Visto, Like & 1-5 Stars Form */}
        <form onSubmit={handleSubmitReview} className="p-4 space-y-3.5">
          <h3 className="text-xs font-extrabold uppercase tracking-wider text-gray-800 dark:text-gray-200">
            Tu valoración y reseña
          </h3>

          {/* Vista/Visto & Like toggles */}
          <div className="flex items-center space-x-2">
            <button
              type="button"
              onClick={() => setWatched(!watched)}
              data-testid="toggle-watched-btn"
              className={`flex-1 touch-target py-2 px-3 rounded-xl text-xs font-bold flex items-center justify-center space-x-1.5 border transition-all ${
                watched
                  ? "bg-emerald-500/15 border-emerald-500/40 text-emerald-600 dark:text-emerald-400"
                  : "bg-gray-100 dark:bg-gray-800 border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400"
              }`}
            >
              <Eye className="w-4 h-4" />
              <span>{watched ? "Vista / Visto ✓" : "Marcar Vista / Visto"}</span>
            </button>

            <button
              type="button"
              onClick={() => setLiked(!liked)}
              data-testid="toggle-like-btn"
              className={`flex-1 touch-target py-2 px-3 rounded-xl text-xs font-bold flex items-center justify-center space-x-1.5 border transition-all ${
                liked
                  ? "bg-rose-500/15 border-rose-500/40 text-rose-600 dark:text-rose-400"
                  : "bg-gray-100 dark:bg-gray-800 border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400"
              }`}
            >
              <Heart className={`w-4 h-4 ${liked ? "fill-current" : ""}`} />
              <span>{liked ? "Me gusta ♥" : "Me gusta"}</span>
            </button>
          </div>

          {/* 1-5 Stars Rating Selector */}
          <div className="flex items-center justify-between bg-gray-50 dark:bg-gray-900/60 px-3 py-2.5 rounded-xl border border-gray-200/70 dark:border-gray-800">
            <span className="text-xs font-semibold text-gray-700 dark:text-gray-300">
              Calificación (1-5 estrellas):
            </span>
            <div className="flex items-center space-x-1">
              {[1, 2, 3, 4, 5].map((starValue) => (
                <button
                  key={starValue}
                  type="button"
                  onClick={() => setRating(starValue)}
                  data-testid={`star-rating-${starValue}`}
                  aria-label={`${starValue} estrellas`}
                  className="p-1 focus:outline-none transition-transform hover:scale-110"
                >
                  <Star
                    className={`w-5 h-5 ${
                      starValue <= rating
                        ? "text-[#FCD116] fill-[#FCD116]"
                        : "text-gray-300 dark:text-gray-600"
                    }`}
                  />
                </button>
              ))}
            </div>
          </div>

          {/* Review Textarea */}
          <div>
            <textarea
              rows={3}
              value={reviewBody}
              onChange={(e) => setReviewBody(e.target.value)}
              data-testid="review-textarea"
              placeholder="Escribe tu reseña para compartirla como publicación en el feed de tus amiguis..."
              className="w-full bg-gray-50 dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-xl p-3 text-xs text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:border-[#FCD116] resize-none"
            />
          </div>

          {publishedSuccess && (
            <div
              data-testid="review-published-banner"
              className="p-2.5 rounded-xl bg-emerald-500/15 border border-emerald-500/30 text-emerald-600 dark:text-emerald-400 text-xs font-semibold flex items-center space-x-2"
            >
              <CheckCircle2 className="w-4 h-4 shrink-0" />
              <span>
                ¡Tu reseña se publicó en el feed para tus seguidores!
              </span>
            </div>
          )}

          <div className="flex space-x-2">
            <button
              type="submit"
              disabled={submitting}
              data-testid="submit-review-btn"
              className="flex-1 touch-target py-2.5 bg-[#FCD116] hover:bg-amber-300 text-[#121417] font-extrabold text-xs rounded-xl transition-colors disabled:opacity-50"
            >
              {submitting ? "Publicando..." : "Publicar reseña en el feed"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
