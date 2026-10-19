package app

import (
	"context"

	"amiguis/internal/platform/events"
	"amiguis/internal/profiles/domain"
	sgDomain "amiguis/internal/socialgraph/domain"
)

type SocialGraphRepository interface {
	Follow(ctx context.Context, follow *sgDomain.Follow) error
	Unfollow(ctx context.Context, followerID, followeeID string) error
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)

	Block(ctx context.Context, block *sgDomain.Block) error
	Unblock(ctx context.Context, blockerID, blockedID string) error
	IsBlocked(ctx context.Context, userA, userB string) (bool, error)

	Mute(ctx context.Context, mute *sgDomain.Mute) error
	Unmute(ctx context.Context, userID, mutedUserID string) error
	IsMuted(ctx context.Context, userID, targetUserID string) (bool, error)

	GetFollowers(ctx context.Context, userID string) ([]*domain.Profile, error)
	GetFollowing(ctx context.Context, userID string) ([]*domain.Profile, error)
	GetFollowingUserIDs(ctx context.Context, userID string) ([]string, error)
	GetMutedUserIDs(ctx context.Context, userID string) ([]string, error)
}

type SocialGraphService struct {
	repo     SocialGraphRepository
	eventBus events.Bus
}

func NewSocialGraphService(repo SocialGraphRepository, eventBus events.Bus) *SocialGraphService {
	return &SocialGraphService{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (s *SocialGraphService) Follow(ctx context.Context, followerID, followeeID string) error {
	follow, err := sgDomain.NewFollow(followerID, followeeID)
	if err != nil {
		return err
	}

	if err := s.repo.Follow(ctx, follow); err != nil {
		return err
	}

	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.EventUserFollowed, map[string]string{
			"follower_id": followerID,
			"followee_id": followeeID,
		})
	}

	return nil
}

func (s *SocialGraphService) Unfollow(ctx context.Context, followerID, followeeID string) error {
	return s.repo.Unfollow(ctx, followerID, followeeID)
}

func (s *SocialGraphService) Block(ctx context.Context, blockerID, blockedID string) error {
	block, err := sgDomain.NewBlock(blockerID, blockedID)
	if err != nil {
		return err
	}
	_ = s.repo.Unfollow(ctx, blockerID, blockedID)
	_ = s.repo.Unfollow(ctx, blockedID, blockerID)
	return s.repo.Block(ctx, block)
}

func (s *SocialGraphService) Unblock(ctx context.Context, blockerID, blockedID string) error {
	return s.repo.Unblock(ctx, blockerID, blockedID)
}

func (s *SocialGraphService) Mute(ctx context.Context, userID, mutedUserID string) error {
	mute, err := sgDomain.NewMute(userID, mutedUserID)
	if err != nil {
		return err
	}
	return s.repo.Mute(ctx, mute)
}

func (s *SocialGraphService) Unmute(ctx context.Context, userID, mutedUserID string) error {
	return s.repo.Unmute(ctx, userID, mutedUserID)
}

func (s *SocialGraphService) GetFollowers(ctx context.Context, userID string) ([]*domain.Profile, error) {
	return s.repo.GetFollowers(ctx, userID)
}

func (s *SocialGraphService) GetFollowing(ctx context.Context, userID string) ([]*domain.Profile, error) {
	return s.repo.GetFollowing(ctx, userID)
}

func (s *SocialGraphService) GetFollowingUserIDs(ctx context.Context, userID string) ([]string, error) {
	return s.repo.GetFollowingUserIDs(ctx, userID)
}

func (s *SocialGraphService) GetMutedUserIDs(ctx context.Context, userID string) ([]string, error) {
	return s.repo.GetMutedUserIDs(ctx, userID)
}
