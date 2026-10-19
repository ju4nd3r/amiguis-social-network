-- name: GetChronologicalFeed :many
-- Guaranteed strictly chronological feed: NO ranking, NO algorithm.
SELECT
    p.id,
    p.author_id,
    p.type,
    p.body,
    p.created_at,
    COALESCE((SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id), 0) AS reactions_count,
    COALESCE((SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL), 0) AS comments_count
FROM posts p
WHERE p.author_id = ANY($1::varchar[])
  AND p.deleted_at IS NULL
  AND ($2::varchar = '' OR p.type = $2)
ORDER BY p.created_at DESC, p.id DESC
LIMIT $3;

-- name: GetChronologicalFeedKeyset :many
SELECT
    p.id,
    p.author_id,
    p.type,
    p.body,
    p.created_at,
    COALESCE((SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id), 0) AS reactions_count,
    COALESCE((SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL), 0) AS comments_count
FROM posts p
WHERE p.author_id = ANY($1::varchar[])
  AND p.deleted_at IS NULL
  AND ($2::varchar = '' OR p.type = $2)
  AND (p.created_at, p.id) < ((SELECT created_at FROM posts WHERE id = $3), $3)
ORDER BY p.created_at DESC, p.id DESC
LIMIT $4;
