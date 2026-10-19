package http

import (
	"encoding/json"
	"net/http"

	"amiguis/internal/catalog/app"
	"amiguis/internal/catalog/domain"
	"amiguis/internal/platform/middleware"

	"github.com/go-chi/chi/v5"
)

type CatalogHandler struct {
	service *app.CatalogService
}

func NewCatalogHandler(service *app.CatalogService) *CatalogHandler {
	return &CatalogHandler{service: service}
}

func (h *CatalogHandler) RegisterRoutes(r chi.Router, optionalAuth func(http.Handler) http.Handler) {
	r.Group(func(pr chi.Router) {
		pr.Use(optionalAuth)
		pr.Get("/catalog", h.Search)
		pr.Get("/catalog/{id}", h.GetByID)
		pr.Post("/catalog/{id}/review", h.SubmitReview)
	})
}

func (h *CatalogHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	items := h.service.Search(r.Context(), q, category)
	respondJSON(w, http.StatusOK, map[string]any{
		"items": items,
	})
}

func (h *CatalogHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, item)
}

type submitReviewRequest struct {
	Body    string `json:"body"`
	Rating  int    `json:"rating"`
	Watched bool   `json:"watched"`
	Liked   bool   `json:"liked"`
}

func (h *CatalogHandler) SubmitReview(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "id")
	authorID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || authorID == "" {
		authorID = "guest-amigui"
	}

	var req submitReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	postDTO, updatedItem, err := h.service.SubmitReview(r.Context(), app.ReviewInput{
		AuthorID: authorID,
		ItemID:   itemID,
		Body:     req.Body,
		Rating:   req.Rating,
		Watched:  req.Watched,
		Liked:    req.Liked,
	})
	if err != nil {
		if err == domain.ErrCatalogItemNotFound {
			respondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"post": postDTO,
		"item": updatedItem,
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
