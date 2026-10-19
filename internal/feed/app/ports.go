package app

import (
	"context"

	contentDomain "amiguis/internal/content/domain"
	"amiguis/internal/feed/domain"
	profileDomain "amiguis/internal/profiles/domain"

	"github.com/google/uuid"
)

type FeedRepository interface {
	GetChronologicalFeed(ctx context.Context, authorIDs []string, filter domain.FeedFilter, cursor string, limit int) ([]*contentDomain.Post, string, error)
	GetPreferences(ctx context.Context, userID string) (*domain.FeedPreferences, error)
	SavePreferences(ctx context.Context, prefs *domain.FeedPreferences) error
	GetLists(ctx context.Context, userID string) ([]*domain.FeedList, error)
	CreateList(ctx context.Context, list *domain.FeedList) error
	AddListMember(ctx context.Context, listID, userID string) error
	RemoveListMember(ctx context.Context, listID, userID string) error
	GetListMembers(ctx context.Context, listID string) ([]string, error)
}

type SocialGraphFetcher interface {
	GetFollowingUserIDs(ctx context.Context, userID string) ([]string, error)
	GetMutedUserIDs(ctx context.Context, userID string) ([]string, error)
}

type ProfileFetcher interface {
	GetByUserID(ctx context.Context, userID string) (*profileDomain.Profile, error)
}

type PostDTO struct {
	ID             string                 `json:"id"`
	Author         *profileDomain.Profile `json:"author"`
	Type           contentDomain.PostType `json:"type"`
	Body           string                 `json:"body"`
	Media          []contentDomain.PostMedia `json:"media"`
	ReactionsCount int                    `json:"reactions_count"`
	CommentsCount  int                    `json:"comments_count"`
	MyReaction     *string                `json:"my_reaction"`
	CreatedAt      string                 `json:"created_at"`
}

type FeedResponseDTO struct {
	Items      []*PostDTO `json:"items"`
	NextCursor string     `json:"next_cursor,omitempty"`
	HasMore    bool       `json:"has_more"`
}

type FeedService struct {
	repo               FeedRepository
	socialGraphFetcher SocialGraphFetcher
	profileFetcher     ProfileFetcher
}

func NewFeedService(repo FeedRepository, sg SocialGraphFetcher, pf ProfileFetcher) *FeedService {
	return &FeedService{
		repo:               repo,
		socialGraphFetcher: sg,
		profileFetcher:     pf,
	}
}

func (s *FeedService) GetFeed(ctx context.Context, viewerID string, filter domain.FeedFilter, cursor string, limit int) (*FeedResponseDTO, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var candidateAuthorIDs []string

	if filter.ListID != "" {
		members, err := s.repo.GetListMembers(ctx, filter.ListID)
		if err == nil {
			candidateAuthorIDs = members
		}
	} else if viewerID != "" {
		// Normal chronological feed: People I follow + my own posts
		following, err := s.socialGraphFetcher.GetFollowingUserIDs(ctx, viewerID)
		if err == nil {
			candidateAuthorIDs = append(following, viewerID)
		} else {
			candidateAuthorIDs = []string{viewerID}
		}
	}

	// Filter out muted authors
	mutedMap := make(map[string]bool)
	if muted, err := s.socialGraphFetcher.GetMutedUserIDs(ctx, viewerID); err == nil {
		for _, m := range muted {
			mutedMap[m] = true
		}
	}

	var activeAuthors []string
	for _, a := range candidateAuthorIDs {
		if !mutedMap[a] {
			activeAuthors = append(activeAuthors, a)
		}
	}

	// Query strictly chronological feed: ORDER BY created_at DESC, id DESC
	posts, nextCursor, err := s.repo.GetChronologicalFeed(ctx, activeAuthors, filter, cursor, limit)
	if err != nil {
		return nil, err
	}

	items := make([]*PostDTO, 0, len(posts))
	for _, p := range posts {
		author, _ := s.profileFetcher.GetByUserID(ctx, p.AuthorID)
		if author == nil {
			author = &profileDomain.Profile{
				UserID:      p.AuthorID,
				Username:    "amigui",
				DisplayName: "Amigui",
				City:        "Colombia",
			}
		}
		items = append(items, &PostDTO{
			ID:             p.ID,
			Author:         author,
			Type:           p.Type,
			Body:           p.Body,
			Media:          p.Media,
			ReactionsCount: p.ReactionsCount,
			CommentsCount:  p.CommentsCount,
			MyReaction:     p.MyReaction,
			CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &FeedResponseDTO{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    nextCursor != "",
	}, nil
}

func (s *FeedService) GetPreferences(ctx context.Context, userID string) (*domain.FeedPreferences, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *FeedService) UpdatePreferences(ctx context.Context, prefs *domain.FeedPreferences) error {
	return s.repo.SavePreferences(ctx, prefs)
}

func (s *FeedService) GetLists(ctx context.Context, userID string) ([]*domain.FeedList, error) {
	return s.repo.GetLists(ctx, userID)
}

func (s *FeedService) CreateList(ctx context.Context, userID, name string) (*domain.FeedList, error) {
	id := uuid.New().String()
	list, err := domain.NewFeedList(id, userID, name)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateList(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

func (s *FeedService) AddListMember(ctx context.Context, listID, memberUserID string) error {
	return s.repo.AddListMember(ctx, listID, memberUserID)
}

func (s *FeedService) RemoveListMember(ctx context.Context, listID, memberUserID string) error {
	return s.repo.RemoveListMember(ctx, listID, memberUserID)
}
