package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidPostType = errors.New("post type must be 'text', 'photo', 'video', or 'review'")
	ErrEmptyPostBody   = errors.New("post body cannot be empty")
	ErrPostBodyTooLong = errors.New("post body exceeds 1000 characters")
	ErrPostNotFound    = errors.New("post not found")
	ErrUnauthorized    = errors.New("unauthorized to modify this post")
)

type PostType string

const (
	PostTypeText   PostType = "text"
	PostTypePhoto  PostType = "photo"
	PostTypeVideo  PostType = "video"
	PostTypeReview PostType = "review"
)

type PostMedia struct {
	ID           string   `json:"id"`
	StorageKey   string   `json:"storage_key"`
	URL          string   `json:"url"`
	ThumbnailURL *string  `json:"thumbnail_url,omitempty"`
	Kind         string   `json:"kind"`
	Width        *int     `json:"width,omitempty"`
	Height       *int     `json:"height,omitempty"`
	Duration     *float64 `json:"duration,omitempty"`
	Status       string   `json:"status"`
}

type ReviewMetadata struct {
	ItemID    string `json:"item_id"`
	Category  string `json:"category"` // "movie", "series", "book"
	ItemTitle string `json:"item_title"`
	ItemYear  int    `json:"item_year"`
	PosterURL string `json:"poster_url"`
	Rating    int    `json:"rating"`  // 0 to 5 stars
	Watched   bool   `json:"watched"` // Vista / Visto
	Liked     bool   `json:"liked"`   // Like heart
	Creator   string `json:"creator,omitempty"`
	Producer  string `json:"producer,omitempty"`
}

type Post struct {
	ID             string          `json:"id"`
	AuthorID       string          `json:"author_id"`
	Type           PostType        `json:"type"`
	Body           string          `json:"body"`
	Media          []PostMedia     `json:"media,omitempty"`
	Review         *ReviewMetadata `json:"review,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty"`
	ReactionsCount int             `json:"reactions_count"`
	CommentsCount  int             `json:"comments_count"`
	MyReaction     *string         `json:"my_reaction,omitempty"`
}

func NewPost(id, authorID string, postType PostType, body string, media []PostMedia) (*Post, error) {
	if postType != PostTypeText && postType != PostTypePhoto && postType != PostTypeVideo && postType != PostTypeReview {
		return nil, ErrInvalidPostType
	}

	cleanBody := strings.TrimSpace(body)
	if cleanBody == "" && len(media) == 0 && postType != PostTypeReview {
		return nil, ErrEmptyPostBody
	}
	if len(cleanBody) > 1000 {
		return nil, ErrPostBodyTooLong
	}

	return &Post{
		ID:        id,
		AuthorID:  authorID,
		Type:      postType,
		Body:      cleanBody,
		Media:     media,
		CreatedAt: time.Now().UTC(),
	}, nil
}
