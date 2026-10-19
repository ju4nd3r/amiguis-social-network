package http

import (
	"encoding/json"
	"net/http"

	"amiguis/internal/engagement/app"
	"amiguis/internal/engagement/domain"
	"amiguis/internal/platform/middleware"

	"github.com/go-chi/chi/v5"
)

type EngagementHandler struct {
	service *app.EngagementService
}

func NewEngagementHandler(service *app.EngagementService) *EngagementHandler {
	return &EngagementHandler{service: service}
}

func (h *EngagementHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler, optionalAuth func(http.Handler) http.Handler) {
	r.Group(func(pr chi.Router) {
		pr.Use(optionalAuth)
		pr.Get("/posts/{id}/comments", h.GetComments)
	})

	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware)
		protected.Post("/posts/{id}/reactions", h.AddReaction)
		protected.Delete("/posts/{id}/reactions", h.RemoveReaction)
		protected.Post("/posts/{id}/comments", h.AddComment)
		protected.Delete("/comments/{id}", h.DeleteComment)
	})
}

type reactionReq struct {
	Type string `json:"type"`
}

func (h *EngagementHandler) AddReaction(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	postID := chi.URLParam(r, "id")

	var req reactionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.AddReaction(r.Context(), postID, userID, domain.ReactionType(req.Type)); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Reaction saved"})
}

func (h *EngagementHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	postID := chi.URLParam(r, "id")

	if err := h.service.RemoveReaction(r.Context(), postID, userID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Reaction removed"})
}

type commentReq struct {
	Body     string  `json:"body"`
	ParentID *string `json:"parent_id,omitempty"`
}

func (h *EngagementHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	authorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	postID := chi.URLParam(r, "id")

	var req commentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	comment, err := h.service.AddComment(r.Context(), postID, authorID, req.ParentID, req.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, comment)
}

func (h *EngagementHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	comments, err := h.service.GetComments(r.Context(), postID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, comments)
}

func (h *EngagementHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	authorID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteComment(r.Context(), id, authorID); err != nil {
		if err == domain.ErrUnauthorizedComment {
			respondError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		respondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Comment deleted"})
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": msg,
		"code":  code,
	})
}
