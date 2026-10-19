package domain

import (
	"time"
)

type Profile struct {
	UserID             string         `json:"id"`
	Username           string         `json:"username"`
	DisplayName        string         `json:"display_name"`
	Bio                string         `json:"bio"`
	AvatarURL          string         `json:"avatar_url"`
	City               string         `json:"city"`
	Website            string         `json:"website"`
	VisibilitySettings map[string]any `json:"visibility_settings,omitempty"`
	FollowersCount     int            `json:"followers_count"`
	FollowingCount     int            `json:"following_count"`
	PostsCount         int            `json:"posts_count"`
	IsFollowing        bool           `json:"is_following"`
	IsMuted            bool           `json:"is_muted"`
	IsBlocked          bool           `json:"is_blocked"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

func NewProfile(userID, username, displayName, city string) *Profile {
	now := time.Now().UTC()
	return &Profile{
		UserID:             userID,
		Username:           username,
		DisplayName:        displayName,
		City:               city,
		VisibilitySettings: make(map[string]any),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}
