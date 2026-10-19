package domain_test

import (
	"strings"
	"testing"

	"amiguis/internal/engagement/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngagementRules(t *testing.T) {
	// Valid Colombian & standard reactions
	for _, rt := range []domain.ReactionType{domain.ReactionLike, domain.ReactionLove, domain.ReactionArepa, domain.ReactionParce, domain.ReactionGenius} {
		reaction, err := domain.NewReaction("post-1", "user-1", rt)
		require.NoError(t, err)
		assert.Equal(t, rt, reaction.Type)
	}

	// Invalid reaction
	_, err := domain.NewReaction("post-1", "user-1", domain.ReactionType("angry"))
	assert.ErrorIs(t, err, domain.ErrInvalidReactionType)

	// Valid comment
	comment, err := domain.NewComment("cid-1", "post-1", "author-1", nil, "¡Excelente publicación, parcero!")
	require.NoError(t, err)
	assert.Equal(t, "¡Excelente publicación, parcero!", comment.Body)

	// Empty comment
	_, err = domain.NewComment("cid-2", "post-1", "author-1", nil, "   ")
	assert.ErrorIs(t, err, domain.ErrEmptyCommentBody)

	// Comment too long (> 500 chars)
	_, err = domain.NewComment("cid-3", "post-1", "author-1", nil, strings.Repeat("x", 501))
	assert.ErrorIs(t, err, domain.ErrCommentBodyTooLong)
}
