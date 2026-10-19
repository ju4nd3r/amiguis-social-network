package app

import (
	"context"

	"amiguis/internal/engagement/domain"
	"amiguis/internal/platform/events"
	profileDomain "amiguis/internal/profiles/domain"

	"github.com/google/uuid"
)

type EngagementRepository interface {
	AddReaction(ctx context.Context, reaction *domain.Reaction) error
	RemoveReaction(ctx context.Context, postID, userID string) error
	GetReactionsCount(ctx context.Context, postID string) (int, error)
	GetUserReaction(ctx context.Context, postID, userID string) (*domain.Reaction, error)

	AddComment(ctx context.Context, comment *domain.Comment) error
	GetCommentsByPostID(ctx context.Context, postID string) ([]*domain.Comment, error)
	FindCommentByID(ctx context.Context, id string) (*domain.Comment, error)
	DeleteComment(ctx context.Context, id, authorID string) error
}

type ProfileFetcher interface {
	GetByUserID(ctx context.Context, userID string) (*profileDomain.Profile, error)
}

type CommentDTO struct {
	ID        string                 `json:"id"`
	PostID    string                 `json:"post_id"`
	Author    *profileDomain.Profile `json:"author"`
	ParentID  *string                `json:"parent_id,omitempty"`
	Body      string                 `json:"body"`
	CreatedAt string                 `json:"created_at"`
	Replies   []*CommentDTO          `json:"replies,omitempty"`
}

type EngagementService struct {
	repo           EngagementRepository
	profileFetcher ProfileFetcher
	eventBus       events.Bus
}

func NewEngagementService(repo EngagementRepository, profileFetcher ProfileFetcher, eventBus events.Bus) *EngagementService {
	return &EngagementService{
		repo:           repo,
		profileFetcher: profileFetcher,
		eventBus:       eventBus,
	}
}

func (s *EngagementService) AddReaction(ctx context.Context, postID, userID string, rType domain.ReactionType) error {
	reaction, err := domain.NewReaction(postID, userID, rType)
	if err != nil {
		return err
	}

	if err := s.repo.AddReaction(ctx, reaction); err != nil {
		return err
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.EventReactionAdded, map[string]string{
			"post_id": postID,
			"user_id": userID,
			"type":    string(rType),
		})
	}

	return nil
}

func (s *EngagementService) RemoveReaction(ctx context.Context, postID, userID string) error {
	return s.repo.RemoveReaction(ctx, postID, userID)
}

func (s *EngagementService) AddComment(ctx context.Context, postID, authorID string, parentID *string, body string) (*CommentDTO, error) {
	id := uuid.New().String()
	comment, err := domain.NewComment(id, postID, authorID, parentID, body)
	if err != nil {
		return nil, err
	}

	if err := s.repo.AddComment(ctx, comment); err != nil {
		return nil, err
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.EventCommentAdded, map[string]string{
			"comment_id": comment.ID,
			"post_id":    comment.PostID,
			"author_id":  comment.AuthorID,
		})
	}

	authorProfile, _ := s.profileFetcher.GetByUserID(ctx, authorID)

	return &CommentDTO{
		ID:        comment.ID,
		PostID:    comment.PostID,
		Author:    authorProfile,
		ParentID:  comment.ParentID,
		Body:      comment.Body,
		CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *EngagementService) GetComments(ctx context.Context, postID string) ([]*CommentDTO, error) {
	comments, err := s.repo.GetCommentsByPostID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Build nested comment tree (parent -> replies)
	dtoMap := make(map[string]*CommentDTO)
	var topLevel []*CommentDTO

	for _, c := range comments {
		author, _ := s.profileFetcher.GetByUserID(ctx, c.AuthorID)
		dto := &CommentDTO{
			ID:        c.ID,
			PostID:    c.PostID,
			Author:    author,
			ParentID:  c.ParentID,
			Body:      c.Body,
			CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Replies:   []*CommentDTO{},
		}
		dtoMap[c.ID] = dto
	}

	for _, c := range comments {
		dto := dtoMap[c.ID]
		if c.ParentID != nil && *c.ParentID != "" {
			if parent, exists := dtoMap[*c.ParentID]; exists {
				parent.Replies = append(parent.Replies, dto)
				continue
			}
		}
		topLevel = append(topLevel, dto)
	}

	return topLevel, nil
}

func (s *EngagementService) DeleteComment(ctx context.Context, id, authorID string) error {
	return s.repo.DeleteComment(ctx, id, authorID)
}
