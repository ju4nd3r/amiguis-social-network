package postgres

import (
	"context"
	"sort"
	"sync"
	"time"

	contentDomain "amiguis/internal/content/domain"
	feedDomain "amiguis/internal/feed/domain"
	"amiguis/internal/platform/database"
)

type MemoryFeedRepository struct {
	mu          sync.RWMutex
	posts       map[string]*contentDomain.Post
	preferences map[string]*feedDomain.FeedPreferences
	lists       map[string]*feedDomain.FeedList
	listMembers map[string]map[string]bool // listID -> userID set
}

func NewMemoryFeedRepository(posts map[string]*contentDomain.Post) *MemoryFeedRepository {
	if posts == nil {
		posts = make(map[string]*contentDomain.Post)
	}
	return &MemoryFeedRepository{
		posts:       posts,
		preferences: make(map[string]*feedDomain.FeedPreferences),
		lists:       make(map[string]*feedDomain.FeedList),
		listMembers: make(map[string]map[string]bool),
	}
}

// AddPost is a helper for testing and seeding
func (r *MemoryFeedRepository) AddPost(p *contentDomain.Post) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.posts[p.ID] = p
}

func (r *MemoryFeedRepository) GetChronologicalFeed(ctx context.Context, authorIDs []string, filter feedDomain.FeedFilter, cursor string, limit int) ([]*contentDomain.Post, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	authorMap := make(map[string]bool)
	for _, id := range authorIDs {
		authorMap[id] = true
	}

	var candidates []*contentDomain.Post
	for _, p := range r.posts {
		if p.DeletedAt != nil {
			continue
		}
		if len(authorIDs) > 0 && !authorMap[p.AuthorID] {
			continue
		}
		if filter.Type != "" && filter.Type != "all" && string(p.Type) != filter.Type {
			continue
		}
		candidates = append(candidates, p)
	}

	// GUARANTEE: Strictly chronological - created_at DESC, id DESC
	// No ranking, no score, no engagement multiplier.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].CreatedAt.Equal(candidates[j].CreatedAt) {
			return candidates[i].ID > candidates[j].ID
		}
		return candidates[i].CreatedAt.After(candidates[j].CreatedAt)
	})

	startIdx := 0
	if cursor != "" {
		for i, p := range candidates {
			if p.ID == cursor {
				startIdx = i + 1
				break
			}
		}
	}

	if startIdx >= len(candidates) {
		return []*contentDomain.Post{}, "", nil
	}

	endIdx := startIdx + limit
	var nextCursor string
	if endIdx < len(candidates) {
		nextCursor = candidates[endIdx-1].ID
	} else {
		endIdx = len(candidates)
	}

	return candidates[startIdx:endIdx], nextCursor, nil
}

func (r *MemoryFeedRepository) GetPreferences(ctx context.Context, userID string) (*feedDomain.FeedPreferences, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if prefs, ok := r.preferences[userID]; ok {
		return prefs, nil
	}
	return feedDomain.DefaultFeedPreferences(userID), nil
}

func (r *MemoryFeedRepository) SavePreferences(ctx context.Context, prefs *feedDomain.FeedPreferences) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preferences[prefs.UserID] = prefs
	return nil
}

func (r *MemoryFeedRepository) GetLists(ctx context.Context, userID string) ([]*feedDomain.FeedList, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*feedDomain.FeedList
	for _, l := range r.lists {
		if l.UserID == userID {
			l.MembersCount = len(r.listMembers[l.ID])
			res = append(res, l)
		}
	}
	return res, nil
}

func (r *MemoryFeedRepository) CreateList(ctx context.Context, list *feedDomain.FeedList) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lists[list.ID] = list
	return nil
}

func (r *MemoryFeedRepository) AddListMember(ctx context.Context, listID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.listMembers[listID] == nil {
		r.listMembers[listID] = make(map[string]bool)
	}
	r.listMembers[listID][userID] = true
	return nil
}

func (r *MemoryFeedRepository) RemoveListMember(ctx context.Context, listID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.listMembers[listID] != nil {
		delete(r.listMembers[listID], userID)
	}
	return nil
}

func (r *MemoryFeedRepository) GetListMembers(ctx context.Context, listID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []string
	if members, ok := r.listMembers[listID]; ok {
		for u := range members {
			res = append(res, u)
		}
	}
	return res, nil
}

// PostgresFeedRepository
type PostgresFeedRepository struct {
	db *database.DB
}

func NewPostgresFeedRepository(db *database.DB) *PostgresFeedRepository {
	return &PostgresFeedRepository{db: db}
}

func (r *PostgresFeedRepository) GetChronologicalFeed(ctx context.Context, authorIDs []string, filter feedDomain.FeedFilter, cursor string, limit int) ([]*contentDomain.Post, string, error) {
	if len(authorIDs) == 0 {
		return []*contentDomain.Post{}, "", nil
	}

	typeFilter := ""
	if filter.Type != "" && filter.Type != "all" {
		typeFilter = " AND p.type = '" + filter.Type + "'"
	}

	var query string
	var rows databaseRows
	var err error

	if cursor == "" {
		query = `SELECT p.id, p.author_id, p.type, p.body, p.created_at,
		                COALESCE((SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id), 0) AS reactions_count,
		                COALESCE((SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL), 0) AS comments_count
		         FROM posts p
		         WHERE p.author_id = ANY($1) AND p.deleted_at IS NULL` + typeFilter + `
		         ORDER BY p.created_at DESC, p.id DESC
		         LIMIT $2`
		rows, err = r.db.Pool.Query(ctx, query, authorIDs, limit+1)
	} else {
		query = `SELECT p.id, p.author_id, p.type, p.body, p.created_at,
		                COALESCE((SELECT COUNT(*) FROM reactions r WHERE r.post_id = p.id), 0) AS reactions_count,
		                COALESCE((SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL), 0) AS comments_count
		         FROM posts p
		         WHERE p.author_id = ANY($1) AND p.deleted_at IS NULL` + typeFilter + `
		           AND (p.created_at, p.id) < ((SELECT created_at FROM posts WHERE id = $2), $2)
		         ORDER BY p.created_at DESC, p.id DESC
		         LIMIT $3`
		rows, err = r.db.Pool.Query(ctx, query, authorIDs, cursor, limit+1)
	}

	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var posts []*contentDomain.Post
	for rows.Next() {
		p := &contentDomain.Post{}
		var postType string
		if err := rows.Scan(&p.ID, &p.AuthorID, &postType, &p.Body, &p.CreatedAt, &p.ReactionsCount, &p.CommentsCount); err != nil {
			return nil, "", err
		}
		p.Type = contentDomain.PostType(postType)
		posts = append(posts, p)
	}

	var nextCursor string
	if len(posts) > limit {
		nextCursor = posts[limit-1].ID
		posts = posts[:limit]
	}

	return posts, nextCursor, nil
}

func (r *PostgresFeedRepository) GetPreferences(ctx context.Context, userID string) (*feedDomain.FeedPreferences, error) {
	query := `SELECT user_id, show_text, show_photo, show_video, updated_at FROM feed_preferences WHERE user_id = $1`
	prefs := &feedDomain.FeedPreferences{}
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(&prefs.UserID, &prefs.ShowText, &prefs.ShowPhoto, &prefs.ShowVideo, &prefs.UpdatedAt)
	if err != nil {
		return feedDomain.DefaultFeedPreferences(userID), nil
	}
	return prefs, nil
}

func (r *PostgresFeedRepository) SavePreferences(ctx context.Context, prefs *feedDomain.FeedPreferences) error {
	query := `INSERT INTO feed_preferences (user_id, show_text, show_photo, show_video, updated_at)
	          VALUES ($1, $2, $3, $4, $5)
	          ON CONFLICT (user_id) DO UPDATE SET
	            show_text = EXCLUDED.show_text,
	            show_photo = EXCLUDED.show_photo,
	            show_video = EXCLUDED.show_video,
	            updated_at = EXCLUDED.updated_at`
	_, err := r.db.Pool.Exec(ctx, query, prefs.UserID, prefs.ShowText, prefs.ShowPhoto, prefs.ShowVideo, time.Now().UTC())
	return err
}

func (r *PostgresFeedRepository) GetLists(ctx context.Context, userID string) ([]*feedDomain.FeedList, error) {
	query := `SELECT l.id, l.user_id, l.name, l.created_at,
	                 COALESCE((SELECT COUNT(*) FROM feed_list_members m WHERE m.list_id = l.id), 0) AS members_count
	          FROM feed_lists l
	          WHERE l.user_id = $1`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []*feedDomain.FeedList
	for rows.Next() {
		l := &feedDomain.FeedList{}
		if err := rows.Scan(&l.ID, &l.UserID, &l.Name, &l.CreatedAt, &l.MembersCount); err != nil {
			return nil, err
		}
		lists = append(lists, l)
	}
	return lists, nil
}

func (r *PostgresFeedRepository) CreateList(ctx context.Context, list *feedDomain.FeedList) error {
	query := `INSERT INTO feed_lists (id, user_id, name, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Pool.Exec(ctx, query, list.ID, list.UserID, list.Name, list.CreatedAt)
	return err
}

func (r *PostgresFeedRepository) AddListMember(ctx context.Context, listID, userID string) error {
	query := `INSERT INTO feed_list_members (list_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Pool.Exec(ctx, query, listID, userID)
	return err
}

func (r *PostgresFeedRepository) RemoveListMember(ctx context.Context, listID, userID string) error {
	query := `DELETE FROM feed_list_members WHERE list_id = $1 AND user_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, listID, userID)
	return err
}

func (r *PostgresFeedRepository) GetListMembers(ctx context.Context, listID string) ([]string, error) {
	query := `SELECT user_id FROM feed_list_members WHERE list_id = $1`
	rows, err := r.db.Pool.Query(ctx, query, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err == nil {
			members = append(members, uid)
		}
	}
	return members, nil
}

type databaseRows interface {
	Next() bool
	Scan(dest ...any) error
	Close()
}
