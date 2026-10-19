package domain_test

import (
	"strings"
	"testing"

	"amiguis/internal/engagement/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngagementRules(t *testing.T) {
	// Valid "interesante" reaction
	reaction, err := domain.NewReaction("post-1", "user-1", domain.ReactionInteresante)
	require.NoError(t, err)
	assert.Equal(t, domain.ReactionInteresante, reaction.Type)

	// Other reactions are rejected
	for _, invalid := range []domain.ReactionType{"like", "love", "arepa", "parce", "angry"} {
		_, err = domain.NewReaction("post-1", "user-1", invalid)
		assert.ErrorIs(t, err, domain.ErrInvalidReactionType)
	}

	// Valid comment
	comment, err := domain.NewComment("cid-1", "post-1", "author-1", nil, "¡Excelente publicación, amigui!")
	require.NoError(t, err)
	assert.Equal(t, "¡Excelente publicación, amigui!", comment.Body)

	// Empty comment
	_, err = domain.NewComment("cid-2", "post-1", "author-1", nil, "   ")
	assert.ErrorIs(t, err, domain.ErrEmptyCommentBody)

	// Comment too long (> 500 chars)
	_, err = domain.NewComment("cid-3", "post-1", "author-1", nil, strings.Repeat("x", 501))
	assert.ErrorIs(t, err, domain.ErrCommentBodyTooLong)
}
