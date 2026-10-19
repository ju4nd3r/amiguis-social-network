package app

import (
	"context"
	"sync"

	"amiguis/internal/notifications/domain"
	"amiguis/internal/platform/events"
	profileDomain "amiguis/internal/profiles/domain"

	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	GetByUser(ctx context.Context, userID string, cursor string, limit int) ([]*domain.Notification, error)
	MarkAsRead(ctx context.Context, id, userID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
	GetPreferences(ctx context.Context, userID string) ([]*domain.NotificationPreference, error)
	SavePreferences(ctx context.Context, prefs []*domain.NotificationPreference) error
}

type ProfileFetcher interface {
	GetByUserID(ctx context.Context, userID string) (*profileDomain.Profile, error)
}

type NotificationDTO struct {
	ID        string                 `json:"id"`
	Type      domain.NotificationType `json:"type"`
	Actor     *profileDomain.Profile `json:"actor"`
	EntityRef *string                `json:"entity_ref,omitempty"`
	ReadAt    *string                `json:"read_at,omitempty"`
	CreatedAt string                 `json:"created_at"`
}

type SSEClient chan []byte

type SSEHub struct {
	mu      sync.RWMutex
	clients map[string]map[SSEClient]bool // userID -> set of client channels
}

func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[string]map[SSEClient]bool),
	}
}

func (h *SSEHub) Register(userID string, client SSEClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[SSEClient]bool)
	}
	h.clients[userID][client] = true
}

func (h *SSEHub) Unregister(userID string, client SSEClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.clients[userID]; ok {
		delete(set, client)
		close(client)
		if len(set) == 0 {
			delete(h.clients, userID)
		}
	}
}

func (h *SSEHub) BroadcastToUser(userID string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if set, ok := h.clients[userID]; ok {
		for client := range set {
			select {
			case client <- data:
			default:
			}
		}
	}
}

type NotificationService struct {
	repo           NotificationRepository
	profileFetcher ProfileFetcher
	sseHub         *SSEHub
}

func NewNotificationService(repo NotificationRepository, profileFetcher ProfileFetcher, sseHub *SSEHub) *NotificationService {
	return &NotificationService{
		repo:           repo,
		profileFetcher: profileFetcher,
		sseHub:         sseHub,
	}
}

func (s *NotificationService) CreateNotification(ctx context.Context, userID string, nType domain.NotificationType, actorID string, entityRef *string) error {
	if userID == actorID {
		return nil // Don't notify self
	}

	id := uuid.New().String()
	notif, err := domain.NewNotification(id, userID, nType, actorID, entityRef)
	if err != nil {
		return err
	}

	if err := s.repo.Create(ctx, notif); err != nil {
		return err
	}

	// Real-time SSE alert to online user
	if s.sseHub != nil {
		s.sseHub.BroadcastToUser(userID, []byte(`{"event":"notification","type":"`+string(nType)+`"}`))
	}

	return nil
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID, cursor string, limit int) ([]*NotificationDTO, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	items, err := s.repo.GetByUser(ctx, userID, cursor, limit)
	if err != nil {
		return nil, err
	}

	var dtos []*NotificationDTO
	for _, it := range items {
		actorProfile, _ := s.profileFetcher.GetByUserID(ctx, it.ActorID)
		var readAtStr *string
		if it.ReadAt != nil {
			t := it.ReadAt.Format("2006-01-02T15:04:05Z07:00")
			readAtStr = &t
		}
		dtos = append(dtos, &NotificationDTO{
			ID:        it.ID,
			Type:      it.Type,
			Actor:     actorProfile,
			EntityRef: it.EntityRef,
			ReadAt:    readAtStr,
			CreatedAt: it.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return dtos, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

func (s *NotificationService) GetPreferences(ctx context.Context, userID string) ([]*domain.NotificationPreference, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *NotificationService) UpdatePreferences(ctx context.Context, prefs []*domain.NotificationPreference) error {
	return s.repo.SavePreferences(ctx, prefs)
}

func (s *NotificationService) RegisterEventHandler(bus events.Bus) {
	if bus == nil {
		return
	}
	bus.Subscribe(events.EventUserFollowed, func(ctx context.Context, evt events.Event) error {
		if m, ok := evt.Payload.(map[string]string); ok {
			follower := m["follower_id"]
			followee := m["followee_id"]
			return s.CreateNotification(ctx, followee, domain.TypeFollow, follower, nil)
		}
		return nil
	})
	bus.Subscribe(events.EventReactionAdded, func(ctx context.Context, evt events.Event) error {
		if m, ok := evt.Payload.(map[string]string); ok {
			postID := m["post_id"]
			userID := m["user_id"]
			ref := postID
			return s.CreateNotification(ctx, userID, domain.TypeReaction, userID, &ref)
		}
		return nil
	})
}
