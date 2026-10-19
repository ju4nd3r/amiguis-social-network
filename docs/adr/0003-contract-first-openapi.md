# ADR 0003: Contract-First API with OpenAPI 3.1

- **Status:** Accepted
- **Date:** 2026-10-18

## Context

Amiguis uses Go on the backend and TypeScript (Next.js PWA) on the frontend within a single monorepo. Without a formal contract, backend and frontend payloads can drift apart.

## Decision

1. `api/openapi.yaml` (OpenAPI 3.1) is the single source of truth for all HTTP endpoints under `/api/v1`.
2. Every endpoint uses uniform error responses (`{"error": "...", "code": "..."}`) and cursor-based pagination for lists.
3. Both the Go HTTP handlers and the TypeScript API client (`web/src/api/client.ts`) adhere strictly to the schemas defined in `api/openapi.yaml`.

## Consequences

- **Positive:** Frontend and backend teams work against an explicit contract; API documentation is always up to date.
- **Negative:** Schema changes require updating `api/openapi.yaml` first before modifying handlers or client code.
