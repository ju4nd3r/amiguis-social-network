package postgres

import (
	"context"
	"strings"
	"sync"
	"time"

	"amiguis/internal/identity/app"
	"amiguis/internal/platform/database"
	"amiguis/internal/profiles/domain"
)

type MemoryProfileRepository struct {
	mu           sync.RWMutex
	profiles     map[string]*domain.Profile
	identityRepo app.UserRepository
}

func NewMemoryProfileRepository(identityRepo app.UserRepository) *MemoryProfileRepository {
	return &MemoryProfileRepository{
		profiles:     make(map[string]*domain.Profile),
		identityRepo: identityRepo,
	}
}

func (r *MemoryProfileRepository) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.profiles[userID]
	if !ok {
		// Auto-populate from user if exists
		u, err := r.identityRepo.FindByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		p = domain.NewProfile(u.ID, u.Username, u.DisplayName, u.City)
		r.profiles[userID] = p
	}
	return p, nil
}

func (r *MemoryProfileRepository) GetByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.profiles {
		if strings.EqualFold(p.Username, username) {
			return p, nil
		}
	}
	u, err := r.identityRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	p := domain.NewProfile(u.ID, u.Username, u.DisplayName, u.City)
	r.profiles[u.ID] = p
	return p, nil
}

func (r *MemoryProfileRepository) Save(ctx context.Context, profile *domain.Profile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	profile.UpdatedAt = time.Now().UTC()
	r.profiles[profile.UserID] = profile
	return nil
}

func (r *MemoryProfileRepository) Search(ctx context.Context, query, city string, limit int) ([]*domain.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.Profile
	q := strings.ToLower(query)
	c := strings.ToLower(city)
	for _, p := range r.profiles {
		matchQ := q == "" || strings.Contains(strings.ToLower(p.Username), q) || strings.Contains(strings.ToLower(p.DisplayName), q)
		matchC := c == "" || strings.Contains(strings.ToLower(p.City), c)
		if matchQ && matchC {
			res = append(res, p)
			if limit > 0 && len(res) >= limit {
				break
			}
		}
	}
	return res, nil
}

func (r *MemoryProfileRepository) GetSocialGraphCounts(ctx context.Context, userID string) (int, int, int, error) {
	return 0, 0, 0, nil
}

type PostgresProfileRepository struct {
	db *database.DB
}

func NewPostgresProfileRepository(db *database.DB) *PostgresProfileRepository {
	return &PostgresProfileRepository{db: db}
}

func (r *PostgresProfileRepository) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	query := `SELECT p.user_id, u.username, p.display_name, p.bio, p.avatar_url, p.city, p.website, p.created_at, p.updated_at
	          FROM profiles p
	          JOIN users u ON u.id = p.user_id
	          WHERE p.user_id = $1`
	p := &domain.Profile{}
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(
		&p.UserID, &p.Username, &p.DisplayName, &p.Bio, &p.AvatarURL, &p.City, &p.Website, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PostgresProfileRepository) GetByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	query := `SELECT p.user_id, u.username, p.display_name, p.bio, p.avatar_url, p.city, p.website, p.created_at, p.updated_at
	          FROM users u
	          LEFT JOIN profiles p ON u.id = p.user_id
	          WHERE LOWER(u.username) = LOWER($1)`
	p := &domain.Profile{}
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&p.UserID, &p.Username, &p.DisplayName, &p.Bio, &p.AvatarURL, &p.City, &p.Website, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PostgresProfileRepository) Save(ctx context.Context, p *domain.Profile) error {
	query := `INSERT INTO profiles (user_id, display_name, bio, avatar_url, city, website, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          ON CONFLICT (user_id) DO UPDATE SET
	            display_name = EXCLUDED.display_name,
	            bio = EXCLUDED.bio,
	            avatar_url = EXCLUDED.avatar_url,
	            city = EXCLUDED.city,
	            website = EXCLUDED.website,
	            updated_at = EXCLUDED.updated_at`
	_, err := r.db.Pool.Exec(ctx, query, p.UserID, p.DisplayName, p.Bio, p.AvatarURL, p.City, p.Website, time.Now().UTC())
	return err
}

func (r *PostgresProfileRepository) Search(ctx context.Context, query, city string, limit int) ([]*domain.Profile, error) {
	sql := `SELECT p.user_id, u.username, p.display_name, p.bio, p.avatar_url, p.city, p.website, p.created_at, p.updated_at
	        FROM profiles p
	        JOIN users u ON u.id = p.user_id
	        WHERE ($1 = '' OR u.username ILIKE '%' || $1 || '%' OR p.display_name ILIKE '%' || $1 || '%')
	          AND ($2 = '' OR p.city ILIKE '%' || $2 || '%')
	        LIMIT $3`
	rows, err := r.db.Pool.Query(ctx, sql, query, city, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Profile
	for rows.Next() {
		p := &domain.Profile{}
		if err := rows.Scan(&p.UserID, &p.Username, &p.DisplayName, &p.Bio, &p.AvatarURL, &p.City, &p.Website, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *PostgresProfileRepository) GetSocialGraphCounts(ctx context.Context, userID string) (int, int, int, error) {
	followers := 0
	following := 0
	posts := 0
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM follows WHERE followee_id = $1`, userID).Scan(&followers)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM follows WHERE follower_id = $1`, userID).Scan(&following)
	_ = r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM posts WHERE author_id = $1 AND deleted_at IS NULL`, userID).Scan(&posts)
	return followers, following, posts, nil
}
