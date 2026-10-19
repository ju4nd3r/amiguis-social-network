package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         int
	DatabaseURL  string
	RedisURL     string
	JWTSecret    string
	StorageType  string // "memory", "local", "s3"
	StorageDir   string
	StorageURL   string
	Environment  string // "development", "test", "production"
	AllowOrigins []string
}

func Load() *Config {
	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			port = val
		}
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://amiguis:amiguis@localhost:5432/amiguis?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "amiguis-colombia-super-secret-key-change-in-prod"
	}

	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "local"
	}

	storageDir := os.Getenv("STORAGE_DIR")
	if storageDir == "" {
		storageDir = "./data/uploads"
	}

	storageURL := os.Getenv("STORAGE_URL")
	if storageURL == "" {
		storageURL = "http://localhost:8080/uploads"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return &Config{
		Port:         port,
		DatabaseURL:  databaseURL,
		RedisURL:     redisURL,
		JWTSecret:    jwtSecret,
		StorageType:  storageType,
		StorageDir:   storageDir,
		StorageURL:   storageURL,
		Environment:  env,
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
	}
}
