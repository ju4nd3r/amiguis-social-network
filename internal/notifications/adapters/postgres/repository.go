package postgres

import (
	"context"
	"sort"
	"sync"
	"time"

	"amiguis/internal/notifications/domain"
	"amiguis/internal/platform/database"
)

type MemoryNotificationRepository struct {
	mu            sync.RWMutex
	notifications map[string]*domain.Notification
	preferences   map[string]map[domain.NotificationType]bool
}

func NewMemoryNotificationRepository() *MemoryNotificationRepository {
	return &MemoryNotificationRepository{
		notifications: make(map[string]*domain.Notification),
		preferences:   make(map[string]map[domain.NotificationType]bool),
	}
}

func (r *MemoryNotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifications[n.ID] = n
	return nil
}

func (r *MemoryNotificationRepository) GetByUser(ctx context.Context, userID string, cursor string, limit int) ([]*domain.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*domain.Notification
	for _, n := range r.notifications {
		if n.UserID == userID {
			list = append(list, n)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	if len(list) > limit {
		list = list[:limit]
	}

	return list, nil
}

func (r *MemoryNotificationRepository) MarkAsRead(ctx context.Context, id, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n, ok := r.notifications[id]; ok && n.UserID == userID {
		now := time.Now().UTC()
		n.ReadAt = &now
	}
	return nil
}

func (r *MemoryNotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	for _, n := range r.notifications {
		if n.UserID == userID && n.ReadAt == nil {
			n.ReadAt = &now
		}
	}
	return nil
}

func (r *MemoryNotificationRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, n := range r.notifications {
		if n.UserID == userID && n.ReadAt == nil {
			count++
		}
	}
	return count, nil
}

func (r *MemoryNotificationRepository) GetPreferences(ctx context.Context, userID string) ([]*domain.NotificationPreference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := []domain.NotificationType{domain.TypeFollow, domain.TypeReaction, domain.TypeComment, domain.TypeReply, domain.TypeMention}
	var res []*domain.NotificationPreference
	userPrefs := r.preferences[userID]
	for _, t := range types {
		enabled := true
		if userPrefs != nil {
			if val, ok := userPrefs[t]; ok {
				enabled = val
			}
		}
		res = append(res, &domain.NotificationPreference{
			UserID:  userID,
			Type:    t,
			Enabled: enabled,
		})
	}
	return res, nil
}

func (r *MemoryNotificationRepository) SavePreferences(ctx context.Context, prefs []*domain.NotificationPreference) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range prefs {
		if r.preferences[p.UserID] == nil {
			r.preferences[p.UserID] = make(map[domain.NotificationType]bool)
		}
		r.preferences[p.UserID][p.Type] = p.Enabled
	}
	return nil
}

type PostgresNotificationRepository struct {
	db *database.DB
}

func NewPostgresNotificationRepository(db *database.DB) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

func (r *PostgresNotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	query := `INSERT INTO notifications (id, user_id, type, actor_id, entity_ref, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Pool.Exec(ctx, query, n.ID, n.UserID, string(n.Type), n.ActorID, n.EntityRef, n.CreatedAt)
	return err
}

func (r *PostgresNotificationRepository) GetByUser(ctx context.Context, userID string, cursor string, limit int) ([]*domain.Notification, error) {
	query := `SELECT id, user_id, type, actor_id, entity_ref, read_at, created_at
	          FROM notifications
	          WHERE user_id = $1
	          ORDER BY created_at DESC
	          LIMIT $2`
	rows, err := r.db.Pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Notification
	for rows.Next() {
		n := &domain.Notification{}
		var nType string
		if err := rows.Scan(&n.ID, &n.UserID, &nType, &n.ActorID, &n.EntityRef, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.Type = domain.NotificationType(nType)
		list = append(list, n)
	}
	return list, nil
}

func (r *PostgresNotificationRepository) MarkAsRead(ctx context.Context, id, userID string) error {
	query := `UPDATE notifications SET read_at = NOW() WHERE id = $1 AND user_id = $2 AND read_at IS NULL`
	_, err := r.db.Pool.Exec(ctx, query, id, userID)
	return err
}

func (r *PostgresNotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `UPDATE notifications SET read_at = NOW() WHERE user_id = $1 AND read_at IS NULL`
	_, err := r.db.Pool.Exec(ctx, query, userID)
	return err
}

func (r *PostgresNotificationRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`, userID).Scan(&count)
	return count, err
}

func (r *PostgresNotificationRepository) GetPreferences(ctx context.Context, userID string) ([]*domain.NotificationPreference, error) {
	query := `SELECT type, enabled FROM notification_preferences WHERE user_id = $1`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*domain.NotificationPreference
	for rows.Next() {
		var nType string
		var enabled bool
		if err := rows.Scan(&nType, &enabled); err == nil {
			res = append(res, &domain.NotificationPreference{
				UserID:  userID,
				Type:    domain.NotificationType(nType),
				Enabled: enabled,
			})
		}
	}
	return res, nil
}

func (r *PostgresNotificationRepository) SavePreferences(ctx context.Context, prefs []*domain.NotificationPreference) error {
	for _, p := range prefs {
		query := `INSERT INTO notification_preferences (user_id, type, enabled)
		          VALUES ($1, $2, $3)
		          ON CONFLICT (user_id, type) DO UPDATE SET enabled = EXCLUDED.enabled`
		_, _ = r.db.Pool.Exec(ctx, query, p.UserID, string(p.Type), p.Enabled)
	}
	return nil
}
