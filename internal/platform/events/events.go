package events

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	EventPostCreated   = "PostCreated"
	EventUserFollowed  = "UserFollowed"
	EventCommentAdded  = "CommentAdded"
	EventReactionAdded = "ReactionAdded"
	EventUserReported  = "UserReported"
)

type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   any       `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type EventHandler func(ctx context.Context, event Event) error

type Bus interface {
	Publish(ctx context.Context, eventType string, payload any) error
	Subscribe(eventType string, handler EventHandler)
}

type MemoryBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (b *MemoryBus) Subscribe(eventType string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *MemoryBus) Publish(ctx context.Context, eventType string, payload any) error {
	b.mu.RLock()
	handlers := b.handlers[eventType]
	b.mu.RUnlock()

	evt := Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}

	for _, h := range handlers {
		// Run handlers safely
		go func(handler EventHandler, e Event) {
			_ = handler(context.Background(), e)
		}(h, evt)
	}

	return nil
}

type OutboxEvent struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
	ProcessedAt *time.Time      `json:"processed_at,omitempty"`
}
