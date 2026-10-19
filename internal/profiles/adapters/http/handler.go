package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"amiguis/internal/platform/middleware"
	"amiguis/internal/profiles/app"

	"github.com/go-chi/chi/v5"
)

type ProfileHandler struct {
	service *app.ProfileService
}

func NewProfileHandler(service *app.ProfileService) *ProfileHandler {
	return &ProfileHandler{service: service}
}

func (h *ProfileHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler, optionalAuth func(http.Handler) http.Handler) {
	r.Group(func(pr chi.Router) {
		pr.Use(optionalAuth)
		pr.Get("/profiles/{username}", h.GetProfile)
		pr.Get("/profiles/search", h.SearchProfiles)
	})

	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware)
		protected.Put("/profiles/me", h.UpdateProfile)
	})
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	viewerID, _ := middleware.UserIDFromContext(r.Context())

	profile, err := h.service.GetProfile(r.Context(), username, viewerID)
	if err != nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "Profile not found")
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

type updateProfileReq struct {
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	City        *string `json:"city"`
	Website     *string `json:"website"`
	AvatarURL   *string `json:"avatar_url"`
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req updateProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	updated, err := h.service.UpdateProfile(r.Context(), app.UpdateProfileInput{
		UserID:      userID,
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		City:        req.City,
		Website:     req.Website,
		AvatarURL:   req.AvatarURL,
	})
	if err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, updated)
}

func (h *ProfileHandler) SearchProfiles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	city := r.URL.Query().Get("city")
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
		limit = val
	}

	results, err := h.service.SearchProfiles(r.Context(), q, city, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, results)
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
