package app_test

import (
	"context"
	"testing"

	"amiguis/internal/catalog/app"
	contentPostgres "amiguis/internal/content/adapters/postgres"
	contentApp "amiguis/internal/content/app"
	contentDomain "amiguis/internal/content/domain"
	"amiguis/internal/platform/events"
	profileDomain "amiguis/internal/profiles/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubProfileFetcher struct{}

func (s *stubProfileFetcher) GetByUserID(_ context.Context, userID string) (*profileDomain.Profile, error) {
	return &profileDomain.Profile{
		UserID:      userID,
		Username:    "camila_medallo",
		DisplayName: "Camila Restrepo",
		City:        "Medellín",
	}, nil
}

func TestCatalogSearchAndReview(t *testing.T) {
	ctx := context.Background()
	memPosts := make(map[string]*contentDomain.Post)
	postRepo := contentPostgres.NewMemoryPostRepository(memPosts)
	contentService := contentApp.NewContentService(postRepo, &stubProfileFetcher{}, events.NewMemoryBus())
	catalogService := app.NewCatalogService(contentService)

	// Search by title
	movies := catalogService.Search(ctx, "Encanto", "movie")
	require.Len(t, movies, 1)
	assert.Equal(t, "Encanto", movies[0].Title)
	assert.Equal(t, 2021, movies[0].Year)
	assert.NotEmpty(t, movies[0].Cast)
	assert.NotEmpty(t, movies[0].Producer)

	// Search books
	books := catalogService.Search(ctx, "olvido", "book")
	require.Len(t, books, 1)
	assert.Equal(t, "El olvido que seremos", books[0].Title)

	// Submit review -> creates a PostTypeReview post and updates rating count
	postDTO, updatedItem, err := catalogService.SubmitReview(ctx, app.ReviewInput{
		AuthorID: "u-1",
		ItemID:   "cat-movie-2",
		Body:     "¡Increíble película!",
		Rating:   5,
		Watched:  true,
		Liked:    true,
	})
	require.NoError(t, err)
	assert.Equal(t, contentDomain.PostTypeReview, postDTO.Type)
	require.NotNil(t, postDTO.Review)
	assert.Equal(t, "Encanto", postDTO.Review.ItemTitle)
	assert.Equal(t, 5, postDTO.Review.Rating)
	assert.True(t, postDTO.Review.Watched)
	assert.True(t, postDTO.Review.Liked)
	assert.Equal(t, 521, updatedItem.RatingsCount)
}
