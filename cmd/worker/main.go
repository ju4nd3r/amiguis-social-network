package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"amiguis/internal/platform/config"
	"amiguis/internal/platform/database"
	"amiguis/internal/platform/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Environment)

	log.Info("Starting Amiguis Background Worker",
		"env", cfg.Environment,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var db *database.DB
	if cfg.DatabaseURL != "" && cfg.DatabaseURL != "memory" {
		var err error
		db, err = database.ConnectPostgres(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Warn("Worker DB unavailable, running idle simulator", "error", err)
		} else {
			defer db.Close()
		}
	}

	// Ticker for outbox polling
	outboxTicker := time.NewTicker(2 * time.Second)
	defer outboxTicker.Stop()

	// Ticker for media processing jobs
	mediaTicker := time.NewTicker(5 * time.Second)
	defer mediaTicker.Stop()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	log.Info("Worker loops initialized. Ready to process outbox and media queues.")

	for {
		select {
		case <-ctx.Done():
			log.Info("Worker context canceled")
			return
		case sig := <-shutdown:
			log.Info("Worker received shutdown signal", "signal", sig)
			return
		case <-outboxTicker.C:
			// Process pending outbox events
			if db != nil && db.Pool != nil {
				processOutboxBatch(ctx, db, log)
			}
		case <-mediaTicker.C:
			// Process pending media transcode jobs
			if db != nil && db.Pool != nil {
				processMediaBatch(ctx, db, log)
			}
		}
	}
}

func processOutboxBatch(ctx context.Context, db *database.DB, log any) {
	// Query unprocessed events
	rows, err := db.Pool.Query(ctx, `SELECT id, type FROM outbox_events WHERE processed_at IS NULL ORDER BY created_at ASC LIMIT 10`)
	if err != nil {
		return
	}
	defer rows.Close()

	var processedIDs []string
	for rows.Next() {
		var id, evtType string
		if err := rows.Scan(&id, &evtType); err == nil {
			processedIDs = append(processedIDs, id)
		}
	}

	for _, id := range processedIDs {
		_, _ = db.Pool.Exec(ctx, `UPDATE outbox_events SET processed_at = NOW() WHERE id = $1`, id)
	}
}

func processMediaBatch(ctx context.Context, db *database.DB, log any) {
	// Query pending media
	rows, err := db.Pool.Query(ctx, `SELECT id, kind FROM post_media WHERE status = 'pending' LIMIT 5`)
	if err != nil {
		return
	}
	defer rows.Close()

	var mediaIDs []string
	for rows.Next() {
		var id, kind string
		if err := rows.Scan(&id, &kind); err == nil {
			mediaIDs = append(mediaIDs, id)
		}
	}

	for _, id := range mediaIDs {
		// Mock transcode / thumbnail generation
		_, _ = db.Pool.Exec(ctx, `UPDATE post_media SET status = 'ready' WHERE id = $1`, id)
	}
}
