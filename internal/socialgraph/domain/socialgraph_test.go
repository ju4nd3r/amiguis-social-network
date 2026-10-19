package domain_test

import (
	"testing"

	"amiguis/internal/socialgraph/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSocialGraphRules(t *testing.T) {
	// Cannot follow self
	_, err := domain.NewFollow("user-1", "user-1")
	assert.ErrorIs(t, err, domain.ErrCannotFollowSelf)

	// Can follow another user
	follow, err := domain.NewFollow("user-1", "user-2")
	require.NoError(t, err)
	assert.Equal(t, "user-1", follow.FollowerID)
	assert.Equal(t, "user-2", follow.FolloweeID)

	// Cannot block self
	_, err = domain.NewBlock("user-1", "user-1")
	assert.ErrorIs(t, err, domain.ErrCannotBlockSelf)

	// Cannot mute self
	_, err = domain.NewMute("user-1", "user-1")
	assert.ErrorIs(t, err, domain.ErrCannotMuteSelf)
}
