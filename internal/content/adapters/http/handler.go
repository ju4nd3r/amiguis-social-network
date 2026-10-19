package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"amiguis/internal/content/app"
	"amiguis/internal/content/domain"
	"amiguis/internal/platform/middleware"

	"github.com/go-chi/chi/v5"
)

type ContentHandler struct {
	service *app.ContentService
}

func NewContentHandler(service *app.ContentService) *ContentHandler {
	return &ContentHandler{service: service}
}

func (h *ContentHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler, optionalAuth func(http.Handler) http.Handler) {
	r.Group(func(pr chi.Router) {
		pr.Use(optionalAuth)
		pr.Get("/posts/{id}", h.GetPost)
		pr.Get("/users/{id}/posts", h.GetUserPosts)
	})

	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware)
		protected.Post("/posts", h.CreatePost)
		protected.Delete("/posts/{id}", h.DeletePost)
	})
}

type createPostRequest struct {
	Type     string             `json:"type"`
	Body     string             `json:"body"`
	MediaIDs []string           `json:"media_ids"`
	Media    []domain.PostMedia `json:"media"`
}

func (h *ContentHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	authorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	post, err := h.service.CreatePost(r.Context(), app.CreatePostInput{
		AuthorID: authorID,
		Type:     domain.PostType(req.Type),
		Body:     req.Body,
		Media:    req.Media,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, post)
}

func (h *ContentHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	viewerID, _ := middleware.UserIDFromContext(r.Context())

	post, err := h.service.GetPost(r.Context(), id, viewerID)
	if err != nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "Post not found")
		return
	}

	respondJSON(w, http.StatusOK, post)
}

func (h *ContentHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	authorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	id := chi.URLParam(r, "id")

	if err := h.service.DeletePost(r.Context(), id, authorID); err != nil {
		if err == domain.ErrUnauthorized {
			respondError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		respondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Post deleted successfully"})
}

func (h *ContentHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	authorID := chi.URLParam(r, "id")
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
		limit = val
	}

	posts, nextCursor, err := h.service.ListAuthorPosts(r.Context(), authorID, cursor, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"items":       posts,
		"next_cursor": nextCursor,
		"has_more":    nextCursor != "",
	})
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
