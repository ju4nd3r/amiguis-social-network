package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"amiguis/internal/media/adapters/storage"
	"amiguis/internal/media/app"

	"github.com/go-chi/chi/v5"
)

type MediaHandler struct {
	service *app.MediaService
	storage *storage.LocalStorageAdapter
}

func NewMediaHandler(service *app.MediaService, storage *storage.LocalStorageAdapter) *MediaHandler {
	return &MediaHandler{
		service: service,
		storage: storage,
	}
}

func (h *MediaHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Get("/media/{id}/status", h.GetMediaStatus)
	r.Put("/uploads/upload/*", h.HandleDirectUpload)

	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware)
		protected.Post("/media/upload-url", h.RequestUploadURL)
	})
}

type uploadURLReq struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

func (h *MediaHandler) RequestUploadURL(w http.ResponseWriter, r *http.Request) {
	var req uploadURLReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	ticket, err := h.service.RequestPresignedUpload(r.Context(), req.Filename, req.ContentType, req.SizeBytes)
	if err != nil {
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ticket)
}

func (h *MediaHandler) GetMediaStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.service.GetMediaStatus(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func (h *MediaHandler) HandleDirectUpload(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/uploads/upload/")
	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	if err := h.storage.SaveFile(key, data); err != nil {
		respondError(w, http.StatusInternalServerError, "SERVER_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true, "key": key})
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
