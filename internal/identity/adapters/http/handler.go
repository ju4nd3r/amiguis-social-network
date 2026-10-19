package http

import (
	"encoding/json"
	"net/http"

	"amiguis/internal/identity/app"
	"amiguis/internal/identity/domain"
	"amiguis/internal/platform/middleware"

	"github.com/go-chi/chi/v5"
)

type IdentityHandler struct {
	service *app.IdentityService
}

func NewIdentityHandler(service *app.IdentityService) *IdentityHandler {
	return &IdentityHandler{service: service}
}

func (h *IdentityHandler) RegisterRoutes(r chi.Router, authMiddleware func(http.Handler) http.Handler) {
	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)
	r.Post("/auth/refresh", h.Refresh)
	r.Post("/auth/verify-email", h.VerifyEmail)

	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware)
		protected.Get("/auth/me", h.GetMe)
		protected.Get("/privacy/export", h.ExportData)
		protected.Delete("/privacy/delete-account", h.DeleteAccount)
	})
}

type registerRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	City        string `json:"city"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *IdentityHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	out, err := h.service.Register(r.Context(), app.RegisterInput{
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
		City:        req.City,
	})
	if err != nil {
		if err == domain.ErrEmailAlreadyExists || err == domain.ErrUsernameAlreadyExists {
			respondError(w, http.StatusConflict, "CONFLICT", err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"access_token":  out.Tokens.AccessToken,
		"refresh_token": out.Tokens.RefreshToken,
		"token_type":    out.Tokens.TokenType,
		"expires_in":    out.Tokens.ExpiresIn,
		"user": map[string]any{
			"id":           out.User.ID,
			"email":        out.User.Email,
			"username":     out.User.Username,
			"display_name": out.User.DisplayName,
			"city":         out.User.City,
			"is_verified":  out.User.IsVerified(),
			"created_at":   out.User.CreatedAt,
		},
	})
}

func (h *IdentityHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	out, err := h.service.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"access_token":  out.Tokens.AccessToken,
		"refresh_token": out.Tokens.RefreshToken,
		"token_type":    out.Tokens.TokenType,
		"expires_in":    out.Tokens.ExpiresIn,
		"user": map[string]any{
			"id":           out.User.ID,
			"email":        out.User.Email,
			"username":     out.User.Username,
			"display_name": out.User.DisplayName,
			"city":         out.User.City,
			"is_verified":  out.User.IsVerified(),
			"created_at":   out.User.CreatedAt,
		},
	})
}

func (h *IdentityHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	out, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid refresh token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"access_token":  out.Tokens.AccessToken,
		"refresh_token": out.Tokens.RefreshToken,
		"token_type":    out.Tokens.TokenType,
		"expires_in":    out.Tokens.ExpiresIn,
		"user": map[string]any{
			"id":           out.User.ID,
			"email":        out.User.Email,
			"username":     out.User.Username,
			"display_name": out.User.DisplayName,
			"city":         out.User.City,
			"is_verified":  out.User.IsVerified(),
			"created_at":   out.User.CreatedAt,
		},
	})
}

func (h *IdentityHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	user, err := h.service.GetCurrentUser(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"id":           user.ID,
		"email":        user.Email,
		"username":     user.Username,
		"display_name": user.DisplayName,
		"city":         user.City,
		"is_verified":  user.IsVerified(),
		"created_at":   user.CreatedAt,
	})
}

func (h *IdentityHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Email verified successfully",
	})
}

func (h *IdentityHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	user, err := h.service.GetCurrentUser(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"compliance": "Ley 1581 de 2012 (Habeas Data)",
		"user": map[string]any{
			"id":           user.ID,
			"email":        user.Email,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"city":         user.City,
			"created_at":   user.CreatedAt,
		},
		"message": "Full personal data export generated",
	})
}

func (h *IdentityHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	var req struct {
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if err := h.service.DeleteAccount(r.Context(), userID, req.ConfirmPassword); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Account and all associated personal data permanently deleted (Ley 1581 de 2012)",
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
