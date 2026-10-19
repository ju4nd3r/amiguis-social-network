package postgres

import (
	"context"
	"sync"
	"time"

	"amiguis/internal/platform/database"
	"amiguis/internal/profiles/app"
	"amiguis/internal/profiles/domain"
	sgDomain "amiguis/internal/socialgraph/domain"
)

type MemorySocialGraphRepository struct {
	mu          sync.RWMutex
	follows     map[string]map[string]time.Time // follower -> set of followees
	blocks      map[string]map[string]time.Time // blocker -> set of blocked
	mutes       map[string]map[string]time.Time // user -> set of muted
	profileRepo app.ProfileRepository
}

func NewMemorySocialGraphRepository(profileRepo app.ProfileRepository) *MemorySocialGraphRepository {
	return &MemorySocialGraphRepository{
		follows:     make(map[string]map[string]time.Time),
		blocks:      make(map[string]map[string]time.Time),
		mutes:       make(map[string]map[string]time.Time),
		profileRepo: profileRepo,
	}
}

func (r *MemorySocialGraphRepository) Follow(ctx context.Context, f *sgDomain.Follow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.follows[f.FollowerID] == nil {
		r.follows[f.FollowerID] = make(map[string]time.Time)
	}
	r.follows[f.FollowerID][f.FolloweeID] = f.CreatedAt
	return nil
}

func (r *MemorySocialGraphRepository) Unfollow(ctx context.Context, followerID, followeeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.follows[followerID] != nil {
		delete(r.follows[followerID], followeeID)
	}
	return nil
}

func (r *MemorySocialGraphRepository) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if set, ok := r.follows[followerID]; ok {
		_, found := set[followeeID]
		return found, nil
	}
	return false, nil
}

func (r *MemorySocialGraphRepository) Block(ctx context.Context, b *sgDomain.Block) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.blocks[b.BlockerID] == nil {
		r.blocks[b.BlockerID] = make(map[string]time.Time)
	}
	r.blocks[b.BlockerID][b.BlockedID] = b.CreatedAt
	return nil
}

func (r *MemorySocialGraphRepository) Unblock(ctx context.Context, blockerID, blockedID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.blocks[blockerID] != nil {
		delete(r.blocks[blockerID], blockedID)
	}
	return nil
}

func (r *MemorySocialGraphRepository) IsBlocked(ctx context.Context, userA, userB string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if set, ok := r.blocks[userA]; ok {
		if _, found := set[userB]; found {
			return true, nil
		}
	}
	if set, ok := r.blocks[userB]; ok {
		if _, found := set[userA]; found {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemorySocialGraphRepository) Mute(ctx context.Context, m *sgDomain.Mute) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.mutes[m.UserID] == nil {
		r.mutes[m.UserID] = make(map[string]time.Time)
	}
	r.mutes[m.UserID][m.MutedUserID] = m.CreatedAt
	return nil
}

func (r *MemorySocialGraphRepository) Unmute(ctx context.Context, userID, mutedUserID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.mutes[userID] != nil {
		delete(r.mutes[userID], mutedUserID)
	}
	return nil
}

func (r *MemorySocialGraphRepository) IsMuted(ctx context.Context, userID, targetUserID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if set, ok := r.mutes[userID]; ok {
		_, found := set[targetUserID]
		return found, nil
	}
	return false, nil
}

func (r *MemorySocialGraphRepository) GetFollowers(ctx context.Context, userID string) ([]*domain.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*domain.Profile
	for followerID, followees := range r.follows {
		if _, ok := followees[userID]; ok {
			if r.profileRepo != nil {
				p, err := r.profileRepo.GetByUserID(ctx, followerID)
				if err == nil && p != nil {
					list = append(list, p)
				}
			}
		}
	}
	return list, nil
}

func (r *MemorySocialGraphRepository) GetFollowing(ctx context.Context, userID string) ([]*domain.Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*domain.Profile
	if followees, ok := r.follows[userID]; ok {
		for followeeID := range followees {
			if r.profileRepo != nil {
				p, err := r.profileRepo.GetByUserID(ctx, followeeID)
				if err == nil && p != nil {
					list = append(list, p)
				}
			}
		}
	}
	return list, nil
}

func (r *MemorySocialGraphRepository) GetFollowingUserIDs(ctx context.Context, userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ids []string
	if followees, ok := r.follows[userID]; ok {
		for id := range followees {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (r *MemorySocialGraphRepository) GetMutedUserIDs(ctx context.Context, userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ids []string
	if mutes, ok := r.mutes[userID]; ok {
		for id := range mutes {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

type PostgresSocialGraphRepository struct {
	db          *database.DB
	profileRepo app.ProfileRepository
}

func NewPostgresSocialGraphRepository(db *database.DB, profileRepo app.ProfileRepository) *PostgresSocialGraphRepository {
	return &PostgresSocialGraphRepository{db: db, profileRepo: profileRepo}
}

func (r *PostgresSocialGraphRepository) Follow(ctx context.Context, f *sgDomain.Follow) error {
	query := `INSERT INTO follows (follower_id, followee_id, created_at)
	          VALUES ($1, $2, $3)
	          ON CONFLICT (follower_id, followee_id) DO NOTHING`
	_, err := r.db.Pool.Exec(ctx, query, f.FollowerID, f.FolloweeID, f.CreatedAt)
	return err
}

func (r *PostgresSocialGraphRepository) Unfollow(ctx context.Context, followerID, followeeID string) error {
	query := `DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, followerID, followeeID)
	return err
}

func (r *PostgresSocialGraphRepository) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	var exists bool
	err := r.db.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = $2)`, followerID, followeeID).Scan(&exists)
	return exists, err
}

func (r *PostgresSocialGraphRepository) Block(ctx context.Context, b *sgDomain.Block) error {
	query := `INSERT INTO blocks (blocker_id, blocked_id, created_at)
	          VALUES ($1, $2, $3)
	          ON CONFLICT (blocker_id, blocked_id) DO NOTHING`
	_, err := r.db.Pool.Exec(ctx, query, b.BlockerID, b.BlockedID, b.CreatedAt)
	return err
}

func (r *PostgresSocialGraphRepository) Unblock(ctx context.Context, blockerID, blockedID string) error {
	query := `DELETE FROM blocks WHERE blocker_id = $1 AND blocked_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, blockerID, blockedID)
	return err
}

func (r *PostgresSocialGraphRepository) IsBlocked(ctx context.Context, userA, userB string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM blocks WHERE (blocker_id = $1 AND blocked_id = $2) OR (blocker_id = $2 AND blocked_id = $1))`
	err := r.db.Pool.QueryRow(ctx, query, userA, userB).Scan(&exists)
	return exists, err
}

func (r *PostgresSocialGraphRepository) Mute(ctx context.Context, m *sgDomain.Mute) error {
	query := `INSERT INTO mutes (user_id, muted_user_id, created_at)
	          VALUES ($1, $2, $3)
	          ON CONFLICT (user_id, muted_user_id) DO NOTHING`
	_, err := r.db.Pool.Exec(ctx, query, m.UserID, m.MutedUserID, m.CreatedAt)
	return err
}

func (r *PostgresSocialGraphRepository) Unmute(ctx context.Context, userID, mutedUserID string) error {
	query := `DELETE FROM mutes WHERE user_id = $1 AND muted_user_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, userID, mutedUserID)
	return err
}

func (r *PostgresSocialGraphRepository) IsMuted(ctx context.Context, userID, targetUserID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM mutes WHERE user_id = $1 AND muted_user_id = $2)`
	err := r.db.Pool.QueryRow(ctx, query, userID, targetUserID).Scan(&exists)
	return exists, err
}

func (r *PostgresSocialGraphRepository) GetFollowers(ctx context.Context, userID string) ([]*domain.Profile, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT follower_id FROM follows WHERE followee_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Profile
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil && r.profileRepo != nil {
			if p, err := r.profileRepo.GetByUserID(ctx, id); err == nil && p != nil {
				list = append(list, p)
			}
		}
	}
	return list, nil
}

func (r *PostgresSocialGraphRepository) GetFollowing(ctx context.Context, userID string) ([]*domain.Profile, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT followee_id FROM follows WHERE follower_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Profile
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil && r.profileRepo != nil {
			if p, err := r.profileRepo.GetByUserID(ctx, id); err == nil && p != nil {
				list = append(list, p)
			}
		}
	}
	return list, nil
}

func (r *PostgresSocialGraphRepository) GetFollowingUserIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT followee_id FROM follows WHERE follower_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (r *PostgresSocialGraphRepository) GetMutedUserIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT muted_user_id FROM mutes WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
