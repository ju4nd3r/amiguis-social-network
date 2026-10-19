# ADR 0001: Strictly Chronological Feed (No Algorithm, No Ads)

- **Status:** Accepted
- **Date:** 2026-10-18

## Context

Mainstream social networks optimize for engagement and time-on-screen using opaque ranking algorithms and behavioral ad targeting. This often leads to polarization, anxiety, and users missing updates from the people they actually care about. Amiguis is built as the first Colombian social network with a core promise: **people see what they choose to see, in the order it happened**.

## Decision

1. **No ranking or recommendation code exists** in the codebase.
2. The feed query is strictly defined as:
   ```sql
   SELECT * FROM posts
   WHERE author_id = ANY($followed_and_unmuted_authors)
     AND deleted_at IS NULL
   ORDER BY created_at DESC, id DESC
   ```
3. Pagination uses deterministic keyset pagination on `(created_at, id)`.
4. Personalization is 100% explicit and user-controlled:
   - Muting an author without unfollowing.
   - Filtering by media type (`all`, `text`, `photo`, `video`).
   - Viewing a custom circle/list (`Familia`, `Trabajo`, `Parceros`).
5. **Automated Guardrail Test:** `internal/feed/guardrail_test.go` inserts posts in randomized order with varying reaction/comment counts and asserts that the feed output is strictly ordered by `created_at DESC, id DESC`. CI fails if this invariant is ever violated.

## Consequences

- **Positive:** Predictable, transparent user experience; simpler query plan; zero ML/ranking infrastructure cost; complete alignment with product ethics.
- **Negative:** Discovery relies on explicit user actions (search by name or Colombian city, user-curated lists) rather than passive algorithmic recommendations.
