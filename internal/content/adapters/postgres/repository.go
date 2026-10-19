package postgres

import (
	"context"
	"sort"
	"sync"
	"time"

	"amiguis/internal/content/domain"
	"amiguis/internal/platform/database"
)

type MemoryPostRepository struct {
	mu    sync.RWMutex
	posts map[string]*domain.Post
}

func NewMemoryPostRepository(posts map[string]*domain.Post) *MemoryPostRepository {
	if posts == nil {
		posts = make(map[string]*domain.Post)
	}
	return &MemoryPostRepository{
		posts: posts,
	}
}

func (r *MemoryPostRepository) Create(ctx context.Context, post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.posts[post.ID] = post
	return nil
}

func (r *MemoryPostRepository) FindByID(ctx context.Context, id string) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.posts[id]
	if !ok || p.DeletedAt != nil {
		return nil, domain.ErrPostNotFound
	}
	return p, nil
}

func (r *MemoryPostRepository) Delete(ctx context.Context, id, authorID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.posts[id]
	if !ok || p.DeletedAt != nil {
		return domain.ErrPostNotFound
	}
	if p.AuthorID != authorID {
		return domain.ErrUnauthorized
	}
	now := time.Now().UTC()
	p.DeletedAt = &now
	return nil
}

func (r *MemoryPostRepository) FindByAuthorID(ctx context.Context, authorID string, cursor string, limit int) ([]*domain.Post, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*domain.Post
	for _, p := range r.posts {
		if p.AuthorID == authorID && p.DeletedAt == nil {
			all = append(all, p)
		}
	}

	// Strictly chronological: created_at DESC, id DESC
	sort.Slice(all, func(i, j int) bool {
		if all[i].CreatedAt.Equal(all[j].CreatedAt) {
			return all[i].ID > all[j].ID
		}
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	startIdx := 0
	if cursor != "" {
		for i, p := range all {
			if p.ID == cursor {
				startIdx = i + 1
				break
			}
		}
	}

	if startIdx >= len(all) {
		return []*domain.Post{}, "", nil
	}

	endIdx := startIdx + limit
	var nextCursor string
	if endIdx < len(all) {
		nextCursor = all[endIdx-1].ID
	} else {
		endIdx = len(all)
	}

	return all[startIdx:endIdx], nextCursor, nil
}

func (r *MemoryPostRepository) GetPostsByIDs(ctx context.Context, ids []string) ([]*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.Post
	for _, id := range ids {
		if p, ok := r.posts[id]; ok && p.DeletedAt == nil {
			res = append(res, p)
		}
	}
	return res, nil
}

func (r *MemoryPostRepository) CountByAuthorID(ctx context.Context, authorID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, p := range r.posts {
		if p.AuthorID == authorID && p.DeletedAt == nil {
			count++
		}
	}
	return count, nil
}

type PostgresPostRepository struct {
	db *database.DB
}

func NewPostgresPostRepository(db *database.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

func (r *PostgresPostRepository) Create(ctx context.Context, p *domain.Post) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO posts (id, author_id, type, body, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(ctx, query, p.ID, p.AuthorID, string(p.Type), p.Body, p.CreatedAt)
	if err != nil {
		return err
	}

	for _, m := range p.Media {
		mediaQuery := `INSERT INTO post_media (id, post_id, storage_key, url, thumbnail_url, kind, width, height, duration, status, created_at)
		               VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
		_, err = tx.Exec(ctx, mediaQuery, m.ID, p.ID, m.StorageKey, m.URL, m.ThumbnailURL, m.Kind, m.Width, m.Height, m.Duration, m.Status, time.Now().UTC())
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresPostRepository) FindByID(ctx context.Context, id string) (*domain.Post, error) {
	query := `SELECT p.id, p.author_id, p.type, p.body, p.created_at, p.deleted_at,
	                 COALESCE((SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id), 0) AS reactions_count,
	                 COALESCE((SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL), 0) AS comments_count
	          FROM posts p
	          WHERE p.id = $1 AND p.deleted_at IS NULL`

	p := &domain.Post{}
	var postType string
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.AuthorID, &postType, &p.Body, &p.CreatedAt, &p.DeletedAt, &p.ReactionsCount, &p.CommentsCount,
	)
	if err != nil {
		return nil, domain.ErrPostNotFound
	}
	p.Type = domain.PostType(postType)

	// Fetch media
	mediaRows, err := r.db.Pool.Query(ctx, `SELECT id, storage_key, url, thumbnail_url, kind, width, height, duration, status FROM post_media WHERE post_id = $1`, id)
	if err == nil {
		defer mediaRows.Close()
		for mediaRows.Next() {
			var m domain.PostMedia
			_ = mediaRows.Scan(&m.ID, &m.StorageKey, &m.URL, &m.ThumbnailURL, &m.Kind, &m.Width, &m.Height, &m.Duration, &m.Status)
			p.Media = append(p.Media, m)
		}
	}

	return p, nil
}

func (r *PostgresPostRepository) Delete(ctx context.Context, id, authorID string) error {
	query := `UPDATE posts SET deleted_at = NOW() WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL`
	cmd, err := r.db.Pool.Exec(ctx, query, id, authorID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrUnauthorized
	}
	return nil
}

func (r *PostgresPostRepository) FindByAuthorID(ctx context.Context, authorID string, cursor string, limit int) ([]*domain.Post, string, error) {
	var query string
	var rows databaseQueryRows
	var err error

	if cursor == "" {
		query = `SELECT p.id, p.author_id, p.type, p.body, p.created_at,
		                COALESCE((SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id), 0) AS reactions_count,
		                COALESCE((SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL), 0) AS comments_count
		         FROM posts p
		         WHERE p.author_id = $1 AND p.deleted_at IS NULL
		         ORDER BY p.created_at DESC, p.id DESC
		         LIMIT $2`
		rows, err = r.db.Pool.Query(ctx, query, authorID, limit+1)
	} else {
		query = `SELECT p.id, p.author_id, p.type, p.body, p.created_at,
		                COALESCE((SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id), 0) AS reactions_count,
		                COALESCE((SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL), 0) AS comments_count
		         FROM posts p
		         WHERE p.author_id = $1 AND p.deleted_at IS NULL
		           AND (p.created_at, p.id) < ((SELECT created_at FROM posts WHERE id = $2), $2)
		         ORDER BY p.created_at DESC, p.id DESC
		         LIMIT $3`
		rows, err = r.db.Pool.Query(ctx, query, authorID, cursor, limit+1)
	}

	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		p := &domain.Post{}
		var postType string
		if err := rows.Scan(&p.ID, &p.AuthorID, &postType, &p.Body, &p.CreatedAt, &p.ReactionsCount, &p.CommentsCount); err != nil {
			return nil, "", err
		}
		p.Type = domain.PostType(postType)
		posts = append(posts, p)
	}

	var nextCursor string
	if len(posts) > limit {
		nextCursor = posts[limit-1].ID
		posts = posts[:limit]
	}

	return posts, nextCursor, nil
}

func (r *PostgresPostRepository) GetPostsByIDs(ctx context.Context, ids []string) ([]*domain.Post, error) {
	if len(ids) == 0 {
		return []*domain.Post{}, nil
	}
	query := `SELECT p.id, p.author_id, p.type, p.body, p.created_at,
	                 COALESCE((SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id), 0) AS reactions_count,
	                 COALESCE((SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL), 0) AS comments_count
	          FROM posts p
	          WHERE p.id = ANY($1) AND p.deleted_at IS NULL`
	rows, err := r.db.Pool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*domain.Post
	for rows.Next() {
		p := &domain.Post{}
		var postType string
		if err := rows.Scan(&p.ID, &p.AuthorID, &postType, &p.Body, &p.CreatedAt, &p.ReactionsCount, &p.CommentsCount); err != nil {
			return nil, err
		}
		p.Type = domain.PostType(postType)
		posts = append(posts, p)
	}
	return posts, nil
}

func (r *PostgresPostRepository) CountByAuthorID(ctx context.Context, authorID string) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM posts WHERE author_id = $1 AND deleted_at IS NULL`, authorID).Scan(&count)
	return count, err
}

type databaseQueryRows interface {
	Next() bool
	Scan(dest ...any) error
	Close()
}
