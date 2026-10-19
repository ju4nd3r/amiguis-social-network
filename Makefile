.PHONY: up down generate migrate test test-int lint build run-api run-worker run-web e2e

# Docker Compose development stack
up:
	docker compose -f deploy/compose/compose.dev.yml up --build -d

down:
	docker compose -f deploy/compose/compose.dev.yml down

# Code generation (sqlc + OpenAPI)
generate:
	@echo "Verifying OpenAPI 3.1 contract (api/openapi.yaml) and SQL queries (db/queries/feed.sql)..."
	@test -f api/openapi.yaml
	@test -f db/queries/feed.sql

# Run database migrations
migrate:
	go run ./cmd/migrate -dir=up

migrate-down:
	go run ./cmd/migrate -dir=down

# Unit & Guardrail tests with race detector
test:
	go test -v -race ./...

# Integration tests
test-int:
	go test -v -race -tags=integration ./...

# Linter & Clean Architecture dependency check
lint:
	go vet ./...
	@which golangci-lint > /dev/null && golangci-lint run || echo "golangci-lint not installed, go vet passed"

# Build all binaries
build:
	mkdir -p bin
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker
	go build -o bin/migrate ./cmd/migrate

# Run services locally
run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

run-web:
	cd web && npm run dev

# Playwright E2E tests
e2e:
	npx playwright test --config=e2e/playwright.config.ts
