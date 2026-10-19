package app

import (
	"context"

	"amiguis/internal/profiles/domain"
)

type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	GetByUsername(ctx context.Context, username string) (*domain.Profile, error)
	Save(ctx context.Context, profile *domain.Profile) error
	Search(ctx context.Context, query, city string, limit int) ([]*domain.Profile, error)
	GetSocialGraphCounts(ctx context.Context, userID string) (followers int, following int, posts int, err error)
}

type UpdateProfileInput struct {
	UserID      string
	DisplayName *string
	Bio         *string
	City        *string
	Website     *string
	AvatarURL   *string
}

type ProfileService struct {
	repo ProfileRepository
}

func NewProfileService(repo ProfileRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) GetProfile(ctx context.Context, username string, viewerID string) (*domain.Profile, error) {
	profile, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	followers, following, posts, err := s.repo.GetSocialGraphCounts(ctx, profile.UserID)
	if err == nil {
		profile.FollowersCount = followers
		profile.FollowingCount = following
		profile.PostsCount = posts
	}

	return profile, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, in UpdateProfileInput) (*domain.Profile, error) {
	profile, err := s.repo.GetByUserID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	if in.DisplayName != nil {
		profile.DisplayName = *in.DisplayName
	}
	if in.Bio != nil {
		profile.Bio = *in.Bio
	}
	if in.City != nil {
		profile.City = *in.City
	}
	if in.Website != nil {
		profile.Website = *in.Website
	}
	if in.AvatarURL != nil {
		profile.AvatarURL = *in.AvatarURL
	}

	if err := s.repo.Save(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *ProfileService) SearchProfiles(ctx context.Context, query, city string, limit int) ([]*domain.Profile, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	return s.repo.Search(ctx, query, city, limit)
}
