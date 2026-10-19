# Amiguis 🇨🇴

> **La primera red social colombiana. Sin algoritmos, sin publicidad. Ves lo que eliges ver.**

Amiguis is a modular monolith built in **Go** (Clean Architecture) and **Next.js + TypeScript** (Mobile-First PWA) in a single repository, designed around strict user autonomy, chronological feeds, and Colombian **Ley 1581 de 2012 (Habeas Data)** compliance.

---

## Architecture Highlights

1. **Strictly Chronological Feed (`ORDER BY created_at DESC, id DESC`):**
   - Zero recommendation or engagement-ranking code exists in the codebase.
   - Guarded by automated guardrail tests (`internal/feed/guardrail_test.go`).
   - Users personalize their feed explicitly through **Mutes**, **Content-Type Filters** (`all`, `photo`, `video`, `text`), and **Circles/Lists**.
2. **Clean Architecture in Go (`internal/`):**
   - Bounded contexts: `identity`, `profiles`, `socialgraph`, `content`, `engagement`, `feed`, `notifications`, `media`, `moderation`.
   - `domain/` depends only on the Go standard library (enforced via `.golangci.yml` `depguard` rules).
   - `app/` defines use cases and consumer-owned ports.
   - `adapters/` provides PostgreSQL (`pgx/v5`), in-memory fallbacks, storage adapters, and `chi` HTTP handlers.
3. **Contract-First API (`api/openapi.yaml`):**
   - OpenAPI 3.1 specification serves as the single source of truth between the Go server and TypeScript PWA client (`web/src/api/client.ts`).
4. **Colombian Privacy & Security (Ley 1581 de 2012):**
   - **Argon2id** password hashing + JWT access/refresh tokens.
   - Self-service personal data JSON export (`GET /api/v1/privacy/export`) and permanent account erasure (`DELETE /api/v1/privacy/delete-account`).

---

## Repository Structure

```text
amiguis/
├── api/
│   └── openapi.yaml          # OpenAPI 3.1 contract (single source of truth)
├── cmd/
│   ├── api/main.go           # HTTP API server (chi router + graceful shutdown)
│   ├── worker/main.go        # Background worker (outbox & media transcoding)
│   └── migrate/main.go       # SQL migration runner
├── internal/
│   ├── identity/             # Auth, registration, Argon2id, Habeas Data
│   ├── profiles/             # Bio, Colombian city, links, visibility
│   ├── socialgraph/          # Follow, unfollow, block, mute
│   ├── content/              # Text, photo, and video posts
│   ├── engagement/           # Colombian reactions (like, parce, arepa) & nested comments
│   ├── feed/                 # Strictly chronological timeline + guardrail test
│   ├── notifications/        # SSE real-time stream & preferences
│   ├── media/                # Presigned uploads & validation
│   ├── moderation/           # Abuse reporting & takedown workflows
│   └── platform/             # Config, slog JSON logger, pgx pool, event bus, middleware
├── db/
│   ├── migrations/           # PostgreSQL schema migrations
│   └── queries/              # Canonical sqlc queries
├── web/                      # Next.js 14 + TypeScript + Tailwind Mobile-First PWA
├── e2e/                      # Playwright E2E test suite (Pixel 5 360px + Desktop)
├── deploy/
│   ├── docker/               # Multi-stage Dockerfiles (api, worker, web)
│   └── compose/              # Docker Compose stacks (dev, test, e2e)
├── docs/adr/                 # Architecture Decision Records (ADRs 0001-0004)
├── .github/workflows/ci.yml  # GitHub Actions CI/CD pipeline
├── .golangci.yml             # Linter & Clean Architecture depguard rules
└── Makefile                  # Developer workflow commands
```

---

## Quick Start

### 1. Run Backend & Tests Locally (Zero External Dependencies Needed)
The API automatically uses in-memory repositories if PostgreSQL is not running, allowing instant local testing:

```bash
# Run all unit & chronological guardrail tests
make test

# Start the Go API server on :8080
make run-api
```

### 2. Run the Next.js Mobile-First PWA
```bash
# Start the web app on :3000
make run-web
```

### 3. Full Containerized Stack (Postgres, Redis, MinIO, Mailpit, API, Worker, Web)
```bash
make up
```
