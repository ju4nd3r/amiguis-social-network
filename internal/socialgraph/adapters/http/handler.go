package http

import (
	"encoding/json"
	"net/http"

	"amiguis/internal/platform/middleware"
	"amiguis/internal/socialgraph/app"

	"github.com/go-chi/chi/v5"
)

type SocialGraphHandler struct {
	service *app.SocialGraphService
}

func NewSocialGraphHandler(service *app.SocialGraphService) *SocialGraphHandler {
	return &SocialGraphHandler{service: service}
}

func (h *SocialGraphHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Get("/users/{id}/followers", h.GetFollowers)
	r.Get("/users/{id}/following", h.GetFollowing)

	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware)
		protected.Put("/users/{id}/follow", h.Follow)
		protected.Delete("/users/{id}/follow", h.Unfollow)
		protected.Put("/users/{id}/block", h.Block)
		protected.Delete("/users/{id}/block", h.Unblock)
		protected.Put("/users/{id}/mute", h.Mute)
		protected.Delete("/users/{id}/mute", h.Unmute)
	})
}

func (h *SocialGraphHandler) Follow(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	targetID := chi.URLParam(r, "id")

	if err := h.service.Follow(r.Context(), currentUserID, targetID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Followed successfully"})
}

func (h *SocialGraphHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	currentUserID, _ := middleware.UserIDFromContext(r.Context())
	targetID := chi.URLParam(r, "id")

	if err := h.service.Unfollow(r.Context(), currentUserID, targetID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Unfollowed successfully"})
}

func (h *SocialGraphHandler) Block(w http.ResponseWriter, r *http.Request) {
	currentUserID, _ := middleware.UserIDFromContext(r.Context())
	targetID := chi.URLParam(r, "id")

	if err := h.service.Block(r.Context(), currentUserID, targetID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Blocked successfully"})
}

func (h *SocialGraphHandler) Unblock(w http.ResponseWriter, r *http.Request) {
	currentUserID, _ := middleware.UserIDFromContext(r.Context())
	targetID := chi.URLParam(r, "id")

	if err := h.service.Unblock(r.Context(), currentUserID, targetID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Unblocked successfully"})
}

func (h *SocialGraphHandler) Mute(w http.ResponseWriter, r *http.Request) {
	currentUserID, _ := middleware.UserIDFromContext(r.Context())
	targetID := chi.URLParam(r, "id")

	if err := h.service.Mute(r.Context(), currentUserID, targetID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Muted successfully"})
}

func (h *SocialGraphHandler) Unmute(w http.ResponseWriter, r *http.Request) {
	currentUserID, _ := middleware.UserIDFromContext(r.Context())
	targetID := chi.URLParam(r, "id")

	if err := h.service.Unmute(r.Context(), currentUserID, targetID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Unmuted successfully"})
}

func (h *SocialGraphHandler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	followers, err := h.service.GetFollowers(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, followers)
}

func (h *SocialGraphHandler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	following, err := h.service.GetFollowing(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, following)
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
