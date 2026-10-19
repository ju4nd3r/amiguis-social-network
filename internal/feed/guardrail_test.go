package feed_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	contentDomain "amiguis/internal/content/domain"
	feedPostgres "amiguis/internal/feed/adapters/postgres"
	feedApp "amiguis/internal/feed/app"
	feedDomain "amiguis/internal/feed/domain"
	profileDomain "amiguis/internal/profiles/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSocialGraph struct {
	following []string
	muted     []string
}

func (m *mockSocialGraph) GetFollowingUserIDs(ctx context.Context, userID string) ([]string, error) {
	return m.following, nil
}

func (m *mockSocialGraph) GetMutedUserIDs(ctx context.Context, userID string) ([]string, error) {
	return m.muted, nil
}

type mockProfileFetcher struct{}

func (m *mockProfileFetcher) GetByUserID(ctx context.Context, userID string) (*profileDomain.Profile, error) {
	return &profileDomain.Profile{
		UserID:      userID,
		Username:    "user_" + userID,
		DisplayName: "User " + userID,
	}, nil
}

// TestFeedGuardrailStrictlyChronological enforces Guiding Principle 1:
// "Chronological, user-controlled feed. No ranking or recommendation code exists.
// The feed is a query: posts from people you follow, newest first, filtered by your own settings."
func TestFeedGuardrailStrictlyChronological(t *testing.T) {
	ctx := context.Background()
	viewerID := "viewer-123"
	authorA := "author-a"
	authorB := "author-b"

	sg := &mockSocialGraph{
		following: []string{authorA, authorB},
		muted:     []string{},
	}
	pf := &mockProfileFetcher{}

	postsMap := make(map[string]*contentDomain.Post)
	repo := feedPostgres.NewMemoryFeedRepository(postsMap)
	service := feedApp.NewFeedService(repo, sg, pf)

	// Create 50 posts with timestamps spaced across multiple hours/days
	baseTime := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	var createdPosts []*contentDomain.Post

	for i := 0; i < 50; i++ {
		author := authorA
		if i%2 == 0 {
			author = authorB
		}
		p := &contentDomain.Post{
			ID:             fmt.Sprintf("post-%03d", i),
			AuthorID:       author,
			Type:           contentDomain.PostTypeText,
			Body:           fmt.Sprintf("Post content %d", i),
			CreatedAt:      baseTime.Add(time.Duration(i) * time.Minute),
			ReactionsCount: rand.Intn(500), // Random engagement (likes/parceros)
			CommentsCount:  rand.Intn(100), // Random comments count
		}
		createdPosts = append(createdPosts, p)
	}

	// Shuffle the posts in random insertion order
	r := rand.New(rand.NewSource(42))
	shuffled := make([]*contentDomain.Post, len(createdPosts))
	copy(shuffled, createdPosts)
	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Insert into repo in completely scrambled order
	for _, p := range shuffled {
		repo.AddPost(p)
	}

	// Query feed
	filter := feedDomain.FeedFilter{Type: "all"}
	resp, err := service.GetFeed(ctx, viewerID, filter, "", 50)
	require.NoError(t, err)
	require.Len(t, resp.Items, 50)

	// ASSERTION: Every item MUST be strictly newer or equal to the next item
	// (created_at DESC, id DESC), independent of reactions or comments!
	for i := 0; i < len(resp.Items)-1; i++ {
		currTime, err := time.Parse("2006-01-02T15:04:05Z07:00", resp.Items[i].CreatedAt)
		require.NoError(t, err)
		nextTime, err := time.Parse("2006-01-02T15:04:05Z07:00", resp.Items[i+1].CreatedAt)
		require.NoError(t, err)

		isChronological := currTime.After(nextTime) || (currTime.Equal(nextTime) && resp.Items[i].ID >= resp.Items[i+1].ID)
		assert.True(t, isChronological,
			"Feed ordering violation at index %d: current post (%s, %s) is older than next post (%s, %s). Amiguis strictly forbids algorithmic or engagement-based reordering!",
			i, resp.Items[i].ID, currTime, resp.Items[i+1].ID, nextTime,
		)
	}

	// Verify latest created post (index 49) is indeed the very first item
	assert.Equal(t, "post-049", resp.Items[0].ID)
	// Verify oldest created post (index 0) is the very last item
	assert.Equal(t, "post-000", resp.Items[49].ID)
}

// TestFeedGuardrailMuteExclusion verifies that muted users are excluded from the feed
func TestFeedGuardrailMuteExclusion(t *testing.T) {
	ctx := context.Background()
	viewerID := "viewer-123"
	authorA := "author-a"
	authorB := "author-b" // will be muted

	sg := &mockSocialGraph{
		following: []string{authorA, authorB},
		muted:     []string{authorB}, // authorB is muted!
	}
	pf := &mockProfileFetcher{}

	postsMap := make(map[string]*contentDomain.Post)
	repo := feedPostgres.NewMemoryFeedRepository(postsMap)
	service := feedApp.NewFeedService(repo, sg, pf)

	now := time.Now().UTC()
	repo.AddPost(&contentDomain.Post{
		ID:        "post-a",
		AuthorID:  authorA,
		Type:      contentDomain.PostTypeText,
		Body:      "Post from A",
		CreatedAt: now,
	})
	repo.AddPost(&contentDomain.Post{
		ID:        "post-b",
		AuthorID:  authorB,
		Type:      contentDomain.PostTypeText,
		Body:      "Post from B should not appear",
		CreatedAt: now.Add(time.Minute),
	})

	resp, err := service.GetFeed(ctx, viewerID, feedDomain.FeedFilter{Type: "all"}, "", 20)
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "post-a", resp.Items[0].ID)
}
