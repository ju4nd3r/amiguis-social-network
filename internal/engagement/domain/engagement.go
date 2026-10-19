package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidReactionType = errors.New("reaction type must be: like, love, arepa, parce, or genius")
	ErrEmptyCommentBody    = errors.New("comment body cannot be empty")
	ErrCommentBodyTooLong  = errors.New("comment body exceeds 500 characters")
	ErrCommentNotFound     = errors.New("comment not found")
	ErrUnauthorizedComment = errors.New("not authorized to delete this comment")
)

type ReactionType string

const (
	ReactionLike   ReactionType = "like"
	ReactionLove   ReactionType = "love"
	ReactionArepa  ReactionType = "arepa"
	ReactionParce  ReactionType = "parce"
	ReactionGenius ReactionType = "genius"
)

type Reaction struct {
	PostID    string       `json:"post_id"`
	UserID    string       `json:"user_id"`
	Type      ReactionType `json:"type"`
	CreatedAt time.Time    `json:"created_at"`
}

type Comment struct {
	ID        string     `json:"id"`
	PostID    string     `json:"post_id"`
	AuthorID  string     `json:"author_id"`
	ParentID  *string    `json:"parent_id,omitempty"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func NewReaction(postID, userID string, rType ReactionType) (*Reaction, error) {
	switch rType {
	case ReactionLike, ReactionLove, ReactionArepa, ReactionParce, ReactionGenius:
	default:
		return nil, ErrInvalidReactionType
	}
	return &Reaction{
		PostID:    postID,
		UserID:    userID,
		Type:      rType,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func NewComment(id, postID, authorID string, parentID *string, body string) (*Comment, error) {
	clean := strings.TrimSpace(body)
	if clean == "" {
		return nil, ErrEmptyCommentBody
	}
	if len(clean) > 500 {
		return nil, ErrCommentBodyTooLong
	}
	return &Comment{
		ID:        id,
		PostID:    postID,
		AuthorID:  authorID,
		ParentID:  parentID,
		Body:      clean,
		CreatedAt: time.Now().UTC(),
	}, nil
}
