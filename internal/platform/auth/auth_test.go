package auth_test

import (
	"testing"

	"amiguis/internal/platform/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgon2idPasswordHashing(t *testing.T) {
	password := "mi-clave-segura-123"

	hash, err := auth.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")

	// Valid password
	valid, err := auth.VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)

	// Invalid password
	valid, err = auth.VerifyPassword("otra-clave-incorrecta", hash)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestJWTTokens(t *testing.T) {
	secret := "test-secret-key-32-chars-length!!"
	userID := "user-abc-123"
	username := "camilo_bogota"

	tokens, err := auth.GenerateTokenPair(secret, userID, username)
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, "Bearer", tokens.TokenType)

	// Validate access token
	claims, err := auth.ValidateToken(secret, tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)

	// Validate with wrong secret fails
	_, err = auth.ValidateToken("wrong-secret", tokens.AccessToken)
	assert.Error(t, err)
}
