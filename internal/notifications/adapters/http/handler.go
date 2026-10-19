package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"amiguis/internal/notifications/app"
	"amiguis/internal/notifications/domain"
	"amiguis/internal/platform/middleware"

	"github.com/go-chi/chi/v5"
)

type NotificationHandler struct {
	service *app.NotificationService
	sseHub  *app.SSEHub
}

func NewNotificationHandler(service *app.NotificationService, sseHub *app.SSEHub) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		sseHub:  sseHub,
	}
}

func (h *NotificationHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware)
		protected.Get("/notifications", h.GetNotifications)
		protected.Put("/notifications/{id}/read", h.MarkRead)
		protected.Put("/notifications/read-all", h.MarkAllRead)
		protected.Get("/notifications/unread-count", h.GetUnreadCount)
		protected.Get("/notifications/stream", h.StreamSSE)
		protected.Get("/notifications/preferences", h.GetPreferences)
		protected.Put("/notifications/preferences", h.UpdatePreferences)
	})
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
		limit = val
	}

	items, err := h.service.GetNotifications(r.Context(), userID, cursor, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.service.MarkRead(r.Context(), id, userID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Notification marked as read"})
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	if err := h.service.MarkAllRead(r.Context(), userID); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "All notifications marked as read"})
}

func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	count, err := h.service.GetUnreadCount(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (h *NotificationHandler) StreamSSE(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(app.SSEClient, 10)
	h.sseHub.Register(userID, clientChan)
	defer h.sseHub.Unregister(userID, clientChan)

	// Send initial ping
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg, ok := <-clientChan:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

func (h *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	prefs, err := h.service.GetPreferences(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, prefs)
}

func (h *NotificationHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var prefs []*domain.NotificationPreference
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}
	for _, p := range prefs {
		p.UserID = userID
	}

	if err := h.service.UpdatePreferences(r.Context(), prefs); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Preferences updated"})
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
