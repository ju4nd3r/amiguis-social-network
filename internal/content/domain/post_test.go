package domain_test

import (
	"strings"
	"testing"

	"amiguis/internal/content/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostValidation(t *testing.T) {
	// Valid text post
	post, err := domain.NewPost("pid-1", "author-1", domain.PostTypeText, "¡Hola Colombia! Bienvenidos a Amiguis 🇨🇴", nil)
	require.NoError(t, err)
	assert.Equal(t, "¡Hola Colombia! Bienvenidos a Amiguis 🇨🇴", post.Body)
	assert.Equal(t, domain.PostTypeText, post.Type)

	// Empty body without media
	_, err = domain.NewPost("pid-2", "author-1", domain.PostTypeText, "   ", nil)
	assert.ErrorIs(t, err, domain.ErrEmptyPostBody)

	// Body too long (> 1000 chars)
	longBody := strings.Repeat("A", 1001)
	_, err = domain.NewPost("pid-3", "author-1", domain.PostTypeText, longBody, nil)
	assert.ErrorIs(t, err, domain.ErrPostBodyTooLong)

	// Invalid post type
	_, err = domain.NewPost("pid-4", "author-1", domain.PostType("invalid_type"), "Some text", nil)
	assert.ErrorIs(t, err, domain.ErrInvalidPostType)
}
