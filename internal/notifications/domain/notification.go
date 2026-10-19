package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidNotificationType = errors.New("invalid notification type")
)

type NotificationType string

const (
	TypeFollow   NotificationType = "follow"
	TypeReaction NotificationType = "reaction"
	TypeComment  NotificationType = "comment"
	TypeReply    NotificationType = "reply"
	TypeMention  NotificationType = "mention"
)

type Notification struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`
	Type      NotificationType `json:"type"`
	ActorID   string           `json:"actor_id"`
	EntityRef *string          `json:"entity_ref,omitempty"`
	ReadAt    *time.Time       `json:"read_at,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

type NotificationPreference struct {
	UserID  string           `json:"user_id"`
	Type    NotificationType `json:"type"`
	Enabled bool             `json:"enabled"`
}

func NewNotification(id, userID string, nType NotificationType, actorID string, entityRef *string) (*Notification, error) {
	switch nType {
	case TypeFollow, TypeReaction, TypeComment, TypeReply, TypeMention:
	default:
		return nil, ErrInvalidNotificationType
	}

	return &Notification{
		ID:        id,
		UserID:    userID,
		Type:      nType,
		ActorID:   actorID,
		EntityRef: entityRef,
		CreatedAt: time.Now().UTC(),
	}, nil
}
