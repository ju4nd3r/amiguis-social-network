package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidListName = errors.New("feed list name must be between 1 and 100 characters")
)

type FeedPreferences struct {
	UserID    string    `json:"user_id"`
	ShowText  bool      `json:"show_text"`
	ShowPhoto bool      `json:"show_photo"`
	ShowVideo bool      `json:"show_video"`
	UpdatedAt time.Time `json:"updated_at"`
}

func DefaultFeedPreferences(userID string) *FeedPreferences {
	return &FeedPreferences{
		UserID:    userID,
		ShowText:  true,
		ShowPhoto: true,
		ShowVideo: true,
		UpdatedAt: time.Now().UTC(),
	}
}

type FeedList struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	MembersCount int       `json:"members_count"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewFeedList(id, userID, name string) (*FeedList, error) {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" || len(cleanName) > 100 {
		return nil, ErrInvalidListName
	}
	return &FeedList{
		ID:           id,
		UserID:       userID,
		Name:         cleanName,
		MembersCount: 0,
		CreatedAt:    time.Now().UTC(),
	}, nil
}

type FeedFilter struct {
	Type   string // "all", "text", "photo", "video"
	ListID string // Optional circle/list filter
}
