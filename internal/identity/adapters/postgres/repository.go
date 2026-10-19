package postgres

import (
	"context"
	"strings"
	"sync"

	"amiguis/internal/identity/domain"
	"amiguis/internal/platform/database"
)

type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (r *MemoryUserRepository) Create(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}

func (r *MemoryUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *MemoryUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if strings.EqualFold(u.Username, username) {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *MemoryUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *MemoryUserRepository) Update(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID] = u
	return nil
}

func (r *MemoryUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}

func (r *MemoryUserRepository) Search(ctx context.Context, query, city string, limit int) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.User
	q := strings.ToLower(query)
	c := strings.ToLower(city)
	for _, u := range r.users {
		matchQ := q == "" || strings.Contains(strings.ToLower(u.Username), q) || strings.Contains(strings.ToLower(u.DisplayName), q)
		matchC := c == "" || strings.Contains(strings.ToLower(u.City), c)
		if matchQ && matchC {
			res = append(res, u)
			if limit > 0 && len(res) >= limit {
				break
			}
		}
	}
	return res, nil
}

// PostgresUserRepository implements UserRepository with PostgreSQL
type PostgresUserRepository struct {
	db *database.DB
}

func NewPostgresUserRepository(db *database.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (id, email, username, password_hash, verified_at, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Pool.Exec(ctx, query, u.ID, u.Email, u.Username, u.PasswordHash, u.VerifiedAt, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, username, password_hash, verified_at, created_at, updated_at FROM users WHERE LOWER(email) = LOWER($1)`
	u := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.VerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *PostgresUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, email, username, password_hash, verified_at, created_at, updated_at FROM users WHERE LOWER(username) = LOWER($1)`
	u := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.VerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, username, password_hash, verified_at, created_at, updated_at FROM users WHERE id = $1`
	u := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.VerifiedAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *domain.User) error {
	query := `UPDATE users SET email = $2, username = $3, password_hash = $4, verified_at = $5, updated_at = $6 WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, u.ID, u.Email, u.Username, u.PasswordHash, u.VerifiedAt, u.UpdatedAt)
	return err
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *PostgresUserRepository) Search(ctx context.Context, query, city string, limit int) ([]*domain.User, error) {
	sql := `SELECT id, email, username, password_hash, verified_at, created_at, updated_at 
	        FROM users WHERE ($1 = '' OR username ILIKE '%' || $1 || '%') LIMIT $2`
	rows, err := r.db.Pool.Query(ctx, sql, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.VerifiedAt, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
