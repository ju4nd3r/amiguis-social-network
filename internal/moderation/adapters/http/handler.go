package http

import (
	"encoding/json"
	"net/http"

	"amiguis/internal/moderation/app"
	"amiguis/internal/moderation/domain"
	"amiguis/internal/platform/middleware"

	"github.com/go-chi/chi/v5"
)

type ModerationHandler struct {
	service *app.ModerationService
}

func NewModerationHandler(service *app.ModerationService) *ModerationHandler {
	return &ModerationHandler{service: service}
}

func (h *ModerationHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware)
		protected.Post("/moderation/reports", h.CreateReport)
		protected.Get("/moderation/reports", h.ListReports)
	})
}

type reportReq struct {
	TargetType string `json:"target_type"`
	TargetID   string `json:"target_id"`
	Reason     string `json:"reason"`
	Details    string `json:"details"`
}

func (h *ModerationHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	reporterID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req reportReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	_, err := h.service.CreateReport(r.Context(), reporterID, domain.TargetType(req.TargetType), req.TargetID, req.Reason, req.Details)
	if err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "Report submitted successfully for moderation review",
	})
}

func (h *ModerationHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	reports, err := h.service.ListReports(r.Context(), status, 50)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, reports)
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
