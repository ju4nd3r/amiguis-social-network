package app

import (
	"context"

	"amiguis/internal/content/domain"
	"amiguis/internal/platform/events"
	profileDomain "amiguis/internal/profiles/domain"

	"github.com/google/uuid"
)

type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	FindByID(ctx context.Context, id string) (*domain.Post, error)
	Delete(ctx context.Context, id, authorID string) error
	FindByAuthorID(ctx context.Context, authorID string, cursor string, limit int) ([]*domain.Post, string, error)
	GetPostsByIDs(ctx context.Context, ids []string) ([]*domain.Post, error)
	CountByAuthorID(ctx context.Context, authorID string) (int, error)
}

type ProfileFetcher interface {
	GetByUserID(ctx context.Context, userID string) (*profileDomain.Profile, error)
}

type CreatePostInput struct {
	AuthorID string
	Type     domain.PostType
	Body     string
	Media    []domain.PostMedia
}

type PostDTO struct {
	ID             string                 `json:"id"`
	Author         *profileDomain.Profile `json:"author"`
	Type           domain.PostType        `json:"type"`
	Body           string                 `json:"body"`
	Media          []domain.PostMedia     `json:"media"`
	ReactionsCount int                    `json:"reactions_count"`
	CommentsCount  int                    `json:"comments_count"`
	MyReaction     *string                `json:"my_reaction"`
	CreatedAt      string                 `json:"created_at"`
}

type ContentService struct {
	repo           PostRepository
	profileFetcher ProfileFetcher
	eventBus       events.Bus
}

func NewContentService(repo PostRepository, profileFetcher ProfileFetcher, eventBus events.Bus) *ContentService {
	return &ContentService{
		repo:           repo,
		profileFetcher: profileFetcher,
		eventBus:       eventBus,
	}
}

func (s *ContentService) CreatePost(ctx context.Context, in CreatePostInput) (*PostDTO, error) {
	postID := uuid.New().String()
	post, err := domain.NewPost(postID, in.AuthorID, in.Type, in.Body, in.Media)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, err
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.EventPostCreated, map[string]string{
			"post_id":   post.ID,
			"author_id": post.AuthorID,
			"type":      string(post.Type),
		})
	}

	authorProfile, _ := s.profileFetcher.GetByUserID(ctx, post.AuthorID)
	if authorProfile == nil {
		authorProfile = &profileDomain.Profile{
			UserID:      post.AuthorID,
			Username:    "amigui",
			DisplayName: "Amigui",
			City:        "Colombia",
		}
	}

	return &PostDTO{
		ID:             post.ID,
		Author:         authorProfile,
		Type:           post.Type,
		Body:           post.Body,
		Media:          post.Media,
		ReactionsCount: 0,
		CommentsCount:  0,
		CreatedAt:      post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *ContentService) GetPost(ctx context.Context, id, viewerID string) (*PostDTO, error) {
	post, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	authorProfile, _ := s.profileFetcher.GetByUserID(ctx, post.AuthorID)
	if authorProfile == nil {
		authorProfile = &profileDomain.Profile{
			UserID:      post.AuthorID,
			Username:    "amigui",
			DisplayName: "Amigui",
			City:        "Colombia",
		}
	}

	return &PostDTO{
		ID:             post.ID,
		Author:         authorProfile,
		Type:           post.Type,
		Body:           post.Body,
		Media:          post.Media,
		ReactionsCount: post.ReactionsCount,
		CommentsCount:  post.CommentsCount,
		MyReaction:     post.MyReaction,
		CreatedAt:      post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (s *ContentService) DeletePost(ctx context.Context, id, authorID string) error {
	return s.repo.Delete(ctx, id, authorID)
}

func (s *ContentService) ListAuthorPosts(ctx context.Context, authorID, cursor string, limit int) ([]*PostDTO, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	posts, nextCursor, err := s.repo.FindByAuthorID(ctx, authorID, cursor, limit)
	if err != nil {
		return nil, "", err
	}

	var dtos []*PostDTO
	authorProfile, _ := s.profileFetcher.GetByUserID(ctx, authorID)

	for _, p := range posts {
		dtos = append(dtos, &PostDTO{
			ID:             p.ID,
			Author:         authorProfile,
			Type:           p.Type,
			Body:           p.Body,
			Media:          p.Media,
			ReactionsCount: p.ReactionsCount,
			CommentsCount:  p.CommentsCount,
			MyReaction:     p.MyReaction,
			CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return dtos, nextCursor, nil
}
