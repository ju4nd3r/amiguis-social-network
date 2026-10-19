package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	direction := flag.String("dir", "up", "Migration direction: up or down")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://amiguis:amiguis@localhost:5432/amiguis?sslmode=disable"
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Starting migration", "direction", *direction, "db_target", maskURL(dbURL))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		logger.Error("Failed to connect to database for migration", "error", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	var migrationFile string
	if *direction == "down" {
		migrationFile = "db/migrations/000001_initial_schema.down.sql"
	} else {
		migrationFile = "db/migrations/000001_initial_schema.up.sql"
	}

	content, err := os.ReadFile(migrationFile)
	if err != nil {
		logger.Error("Failed to read migration file", "file", migrationFile, "error", err)
		os.Exit(1)
	}

	_, err = conn.Exec(ctx, string(content))
	if err != nil {
		logger.Error("Failed to execute migration", "file", migrationFile, "error", err)
		os.Exit(1)
	}

	logger.Info("Migration applied successfully", "file", migrationFile)
}

func maskURL(url string) string {
	parts := strings.Split(url, "@")
	if len(parts) > 1 {
		return fmt.Sprintf("postgres://***@%s", parts[1])
	}
	return url
}
