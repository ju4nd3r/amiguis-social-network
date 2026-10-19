package domain

import (
	"errors"
	"time"
)

var (
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	ErrCannotBlockSelf  = errors.New("cannot block yourself")
	ErrCannotMuteSelf   = errors.New("cannot mute yourself")
)

type Follow struct {
	FollowerID string
	FolloweeID string
	CreatedAt  time.Time
}

type Block struct {
	BlockerID string
	BlockedID string
	CreatedAt time.Time
}

type Mute struct {
	UserID      string
	MutedUserID string
	CreatedAt   time.Time
}

func NewFollow(followerID, followeeID string) (*Follow, error) {
	if followerID == followeeID {
		return nil, ErrCannotFollowSelf
	}
	return &Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
		CreatedAt:  time.Now().UTC(),
	}, nil
}

func NewBlock(blockerID, blockedID string) (*Block, error) {
	if blockerID == blockedID {
		return nil, ErrCannotBlockSelf
	}
	return &Block{
		BlockerID: blockerID,
		BlockedID: blockedID,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func NewMute(userID, mutedUserID string) (*Mute, error) {
	if userID == mutedUserID {
		return nil, ErrCannotMuteSelf
	}
	return &Mute{
		UserID:      userID,
		MutedUserID: mutedUserID,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
