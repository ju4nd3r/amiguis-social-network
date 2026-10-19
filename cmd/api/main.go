package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"amiguis/internal/content"
	contentDomain "amiguis/internal/content/domain"
	"amiguis/internal/engagement"
	"amiguis/internal/feed"
	"amiguis/internal/identity"
	"amiguis/internal/media"
	"amiguis/internal/moderation"
	"amiguis/internal/notifications"
	"amiguis/internal/platform/config"
	"amiguis/internal/platform/database"
	"amiguis/internal/platform/events"
	"amiguis/internal/platform/logger"
	"amiguis/internal/platform/middleware"
	"amiguis/internal/profiles"
	"amiguis/internal/socialgraph"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Environment)

	log.Info("Starting Amiguis API Monolith",
		"env", cfg.Environment,
		"port", cfg.Port,
	)

	// Context for graceful startup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to PostgreSQL (gracefully degrades to in-memory if DB is unavailable in dev)
	var db *database.DB
	if cfg.DatabaseURL != "" && cfg.DatabaseURL != "memory" {
		var err error
		db, err = database.ConnectPostgres(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Warn("PostgreSQL unavailable, running with in-memory stores for local session", "error", err)
			db = nil
		} else {
			log.Info("Connected to PostgreSQL successfully")
			defer db.Close()
		}
	}

	// Shared Event Bus
	eventBus := events.NewMemoryBus()

	// Memory cache for content when running in memory mode
	memoryPosts := make(map[string]*contentDomain.Post)

	// Clean Architecture: Manual Dependency Injection
	identityMod := identity.NewModule(db, cfg.JWTSecret)
	profileMod := profiles.NewModule(db, identityMod.Repo)
	sgMod := socialgraph.NewModule(db, profileMod.Repo, eventBus)
	contentMod := content.NewModule(db, profileMod.Repo, eventBus, memoryPosts)
	engagementMod := engagement.NewModule(db, profileMod.Repo, eventBus)
	feedMod := feed.NewModule(db, sgMod.Service, profileMod.Repo, memoryPosts)
	notifMod := notifications.NewModule(db, profileMod.Repo, eventBus)
	mediaMod := media.NewModule(cfg.StorageDir, cfg.StorageURL)
	moderationMod := moderation.NewModule()

	// Router setup
	r := chi.NewRouter()

	// Global Middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.StructuredLogger(log))
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.NewRateLimiter(100, 1*time.Minute).Middleware())

	// CORS config
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080", "*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health Checks
	r.Get("/health/live", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]any{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})
	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		code := http.StatusOK
		if db != nil && db.Pool != nil {
			if err := db.Pool.Ping(r.Context()); err != nil {
				status = "degraded"
				code = http.StatusServiceUnavailable
			}
		}
		respondJSON(w, code, map[string]any{
			"status":    status,
			"timestamp": time.Now().UTC(),
		})
	})

	// Static local upload server
	_ = os.MkdirAll(cfg.StorageDir, 0755)
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.StorageDir))))

	// Mount API v1
	authMW := middleware.RequireAuth(cfg.JWTSecret)
	optAuthMW := middleware.OptionalAuth(cfg.JWTSecret)

	r.Route("/api/v1", func(v1 chi.Router) {
		identityMod.Handler.RegisterRoutes(v1, authMW)
		profileMod.Handler.RegisterRoutes(v1, authMW, optAuthMW)
		sgMod.Handler.RegisterRoutes(v1, authMW)
		contentMod.Handler.RegisterRoutes(v1, authMW, optAuthMW)
		engagementMod.Handler.RegisterRoutes(v1, authMW, optAuthMW)
		feedMod.Handler.RegisterRoutes(v1, authMW, optAuthMW)
		notifMod.Handler.RegisterRoutes(v1, authMW)
		mediaMod.Handler.RegisterRoutes(v1, authMW)
		moderationMod.Handler.RegisterRoutes(v1, authMW)
	})

	// Server and Graceful Shutdown
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("Server listening", "addr", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Listen for OS signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Error("Server error", "error", err)
	case sig := <-shutdown:
		log.Info("Shutdown signal received", "signal", sig)
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := server.Shutdown(ctxShutdown); err != nil {
			log.Error("Graceful shutdown failed, forcing close", "error", err)
			_ = server.Close()
		}
		log.Info("Amiguis server stopped gracefully")
	}
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
