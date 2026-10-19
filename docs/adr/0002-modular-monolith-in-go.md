# ADR 0002: Modular Monolith in Go with Clean Architecture

- **Status:** Accepted
- **Date:** 2026-10-18

## Context

Early-stage social platforms need low hosting costs, high concurrency for real-time notifications (SSE) and background media processing, and rapid iteration without distributed-systems overhead. At the same time, careless monoliths quickly turn into a "big ball of mud."

## Decision

1. **Language & Runtime:** Go (1.23+) with `net/http` and `chi` router.
2. **Modular Monolith Topology:** One repository and two primary deployable binaries (`cmd/api` and `cmd/worker`) sharing bounded contexts under `internal/`:
   - `identity`, `profiles`, `socialgraph`, `content`, `engagement`, `feed`, `notifications`, `media`, `moderation`.
3. **Clean Architecture Inside Each Module:**
   - `domain/`: Pure business entities, value objects, and validation rules (standard library imports only).
   - `app/`: Use cases and small consumer-defined port interfaces.
   - `adapters/`: PostgreSQL (`pgx`), in-memory test/dev stores, HTTP handlers, and storage adapters.
   - `module.go`: Explicit constructor-based wiring (no reflection DI frameworks).
4. **Cross-Module Communication:** Modules communicate only via exported application services or domain events (`internal/platform/events`), never by reaching into another module's database tables or adapters.

## Consequences

- **Positive:** Tiny container images, low RAM/CPU footprint, sub-second test execution for `domain` and `app`, and clean seams to extract a microservice in the future if scale requires it.
- **Negative:** Requires discipline and linter rules (`.golangci.yml` with `depguard`) to prevent cross-module import violations.
