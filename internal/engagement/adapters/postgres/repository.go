package postgres

import (
	"context"
	"sort"
	"sync"
	"time"

	"amiguis/internal/engagement/domain"
	"amiguis/internal/platform/database"
)

type MemoryEngagementRepository struct {
	mu        sync.RWMutex
	reactions map[string]map[string]*domain.Reaction // postID -> userID -> Reaction
	comments  map[string]*domain.Comment
}

func NewMemoryEngagementRepository() *MemoryEngagementRepository {
	return &MemoryEngagementRepository{
		reactions: make(map[string]map[string]*domain.Reaction),
		comments:  make(map[string]*domain.Comment),
	}
}

func (r *MemoryEngagementRepository) AddReaction(ctx context.Context, reaction *domain.Reaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.reactions[reaction.PostID] == nil {
		r.reactions[reaction.PostID] = make(map[string]*domain.Reaction)
	}
	r.reactions[reaction.PostID][reaction.UserID] = reaction
	return nil
}

func (r *MemoryEngagementRepository) RemoveReaction(ctx context.Context, postID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.reactions[postID] != nil {
		delete(r.reactions[postID], userID)
	}
	return nil
}

func (r *MemoryEngagementRepository) GetReactionsCount(ctx context.Context, postID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if set, ok := r.reactions[postID]; ok {
		return len(set), nil
	}
	return 0, nil
}

func (r *MemoryEngagementRepository) GetUserReaction(ctx context.Context, postID, userID string) (*domain.Reaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if set, ok := r.reactions[postID]; ok {
		if reaction, exists := set[userID]; exists {
			return reaction, nil
		}
	}
	return nil, nil
}

func (r *MemoryEngagementRepository) AddComment(ctx context.Context, c *domain.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.comments[c.ID] = c
	return nil
}

func (r *MemoryEngagementRepository) GetCommentsByPostID(ctx context.Context, postID string) ([]*domain.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*domain.Comment
	for _, c := range r.comments {
		if c.PostID == postID && c.DeletedAt == nil {
			list = append(list, c)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})

	return list, nil
}

func (r *MemoryEngagementRepository) FindCommentByID(ctx context.Context, id string) (*domain.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.comments[id]
	if !ok || c.DeletedAt != nil {
		return nil, domain.ErrCommentNotFound
	}
	return c, nil
}

func (r *MemoryEngagementRepository) DeleteComment(ctx context.Context, id, authorID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.comments[id]
	if !ok || c.DeletedAt != nil {
		return domain.ErrCommentNotFound
	}
	if c.AuthorID != authorID {
		return domain.ErrUnauthorizedComment
	}
	now := time.Now().UTC()
	c.DeletedAt = &now
	return nil
}

type PostgresEngagementRepository struct {
	db *database.DB
}

func NewPostgresEngagementRepository(db *database.DB) *PostgresEngagementRepository {
	return &PostgresEngagementRepository{db: db}
}

func (r *PostgresEngagementRepository) AddReaction(ctx context.Context, react *domain.Reaction) error {
	query := `INSERT INTO reactions (post_id, user_id, type, created_at)
	          VALUES ($1, $2, $3, $4)
	          ON CONFLICT (post_id, user_id) DO UPDATE SET type = EXCLUDED.type`
	_, err := r.db.Pool.Exec(ctx, query, react.PostID, react.UserID, string(react.Type), react.CreatedAt)
	return err
}

func (r *PostgresEngagementRepository) RemoveReaction(ctx context.Context, postID, userID string) error {
	query := `DELETE FROM reactions WHERE post_id = $1 AND user_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, postID, userID)
	return err
}

func (r *PostgresEngagementRepository) GetReactionsCount(ctx context.Context, postID string) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM reactions WHERE post_id = $1`, postID).Scan(&count)
	return count, err
}

func (r *PostgresEngagementRepository) GetUserReaction(ctx context.Context, postID, userID string) (*domain.Reaction, error) {
	query := `SELECT post_id, user_id, type, created_at FROM reactions WHERE post_id = $1 AND user_id = $2`
	var react domain.Reaction
	var rType string
	err := r.db.Pool.QueryRow(ctx, query, postID, userID).Scan(&react.PostID, &react.UserID, &rType, &react.CreatedAt)
	if err != nil {
		return nil, nil
	}
	react.Type = domain.ReactionType(rType)
	return &react, nil
}

func (r *PostgresEngagementRepository) AddComment(ctx context.Context, c *domain.Comment) error {
	query := `INSERT INTO comments (id, post_id, author_id, parent_id, body, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Pool.Exec(ctx, query, c.ID, c.PostID, c.AuthorID, c.ParentID, c.Body, c.CreatedAt)
	return err
}

func (r *PostgresEngagementRepository) GetCommentsByPostID(ctx context.Context, postID string) ([]*domain.Comment, error) {
	query := `SELECT id, post_id, author_id, parent_id, body, created_at, deleted_at
	          FROM comments
	          WHERE post_id = $1 AND deleted_at IS NULL
	          ORDER BY created_at ASC`
	rows, err := r.db.Pool.Query(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		c := &domain.Comment{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.Body, &c.CreatedAt, &c.DeletedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *PostgresEngagementRepository) FindCommentByID(ctx context.Context, id string) (*domain.Comment, error) {
	query := `SELECT id, post_id, author_id, parent_id, body, created_at, deleted_at FROM comments WHERE id = $1 AND deleted_at IS NULL`
	c := &domain.Comment{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&c.ID, &c.PostID, &c.AuthorID, &c.ParentID, &c.Body, &c.CreatedAt, &c.DeletedAt)
	if err != nil {
		return nil, domain.ErrCommentNotFound
	}
	return c, nil
}

func (r *PostgresEngagementRepository) DeleteComment(ctx context.Context, id, authorID string) error {
	query := `UPDATE comments SET deleted_at = NOW() WHERE id = $1 AND author_id = $2 AND deleted_at IS NULL`
	cmd, err := r.db.Pool.Exec(ctx, query, id, authorID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrUnauthorizedComment
	}
	return nil
}
