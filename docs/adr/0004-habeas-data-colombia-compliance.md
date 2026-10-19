# ADR 0004: Privacy by Design & Colombian Habeas Data Compliance (Ley 1581 de 2012)

- **Status:** Accepted
- **Date:** 2026-10-18

## Context

As a Colombian social network operating under **Ley Estatutaria 1581 de 2012 (Habeas Data)** and **Decreto 1377 de 2013**, Amiguis must guarantee citizens' constitutional rights to know, update, rectify, and delete their personal data, and must collect only data strictly necessary for the service without third-party ad profiling.

## Decision

1. **Zero Behavioral Profiling:** No ad-tech SDKs, no tracking pixels, and no selling or sharing of user data with third parties. Operational telemetry (`log/slog`, Prometheus, OpenTelemetry) records only system health and error rates.
2. **Self-Service Data Portability (`GET /api/v1/privacy/export`):** Authenticated users can export all of their personal information, profile data, and posts in machine-readable JSON at any time.
3. **Right to Erasure (`DELETE /api/v1/privacy/delete-account`):** Users can permanently delete their account and cascade-delete all associated personal records (`ON DELETE CASCADE` in PostgreSQL) after confirming their password.
4. **Password Security:** Passwords are hashed using **Argon2id** (`golang.org/x/crypto/argon2`) with 64 MB memory cost and cryptographically random 16-byte salts.

## Consequences

- **Positive:** Full legal compliance with Superintendencia de Industria y Comercio (SIC) requirements and strong user trust.
- **Negative:** Account deletion is immediate and irreversible; backups must be encrypted and rotated on a strict retention window.
