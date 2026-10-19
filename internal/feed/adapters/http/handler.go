package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	feedApp "amiguis/internal/feed/app"
	"amiguis/internal/feed/domain"
	"amiguis/internal/platform/middleware"

	"github.com/go-chi/chi/v5"
)

type FeedHandler struct {
	service *feedApp.FeedService
}

func NewFeedHandler(service *feedApp.FeedService) *FeedHandler {
	return &FeedHandler{service: service}
}

func (h *FeedHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler, optionalAuth func(http.Handler) http.Handler) {
	r.Group(func(pr chi.Router) {
		pr.Use(optionalAuth)
		pr.Get("/feed", h.GetFeed)
	})

	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware)
		protected.Get("/feed/preferences", h.GetPreferences)
		protected.Put("/feed/preferences", h.UpdatePreferences)
		protected.Get("/feed/lists", h.GetLists)
		protected.Post("/feed/lists", h.CreateList)
		protected.Post("/feed/lists/{id}/members", h.AddListMember)
		protected.Delete("/feed/lists/{id}/members/{user_id}", h.RemoveListMember)
	})
}

func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	viewerID, _ := middleware.UserIDFromContext(r.Context())
	cursor := r.URL.Query().Get("cursor")
	postType := r.URL.Query().Get("type")
	listID := r.URL.Query().Get("list_id")

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
		limit = val
	}

	filter := domain.FeedFilter{
		Type:   postType,
		ListID: listID,
	}

	feedResp, err := h.service.GetFeed(r.Context(), viewerID, filter, cursor, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, feedResp)
}

func (h *FeedHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	prefs, err := h.service.GetPreferences(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, prefs)
}

func (h *FeedHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var prefs domain.FeedPreferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	prefs.UserID = userID

	if err := h.service.UpdatePreferences(r.Context(), &prefs); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, prefs)
}

func (h *FeedHandler) GetLists(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	lists, err := h.service.GetLists(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, lists)
}

type createListReq struct {
	Name string `json:"name"`
}

func (h *FeedHandler) CreateList(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var req createListReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	list, err := h.service.CreateList(r.Context(), userID, req.Name)
	if err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, list)
}

func (h *FeedHandler) AddListMember(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "id")
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.AddListMember(r.Context(), listID, req.UserID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Member added to circle"})
}

func (h *FeedHandler) RemoveListMember(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "id")
	targetUserID := chi.URLParam(r, "user_id")

	if err := h.service.RemoveListMember(r.Context(), listID, targetUserID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Member removed from circle"})
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
