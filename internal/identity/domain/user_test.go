package domain_test

import (
	"testing"

	"amiguis/internal/identity/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserValidation(t *testing.T) {
	// Valid user
	user, err := domain.NewUser("uid-1", "juan@amiguis.co", "juan_perez", "hashed_pwd", "Juan Pérez", "Medellín")
	require.NoError(t, err)
	assert.Equal(t, "juan@amiguis.co", user.Email)
	assert.Equal(t, "juan_perez", user.Username)
	assert.Equal(t, "Juan Pérez", user.DisplayName)
	assert.Equal(t, "Medellín", user.City)

	// Invalid email
	_, err = domain.NewUser("uid-2", "not-an-email", "valid_user", "hashed_pwd", "Test", "Cali")
	assert.ErrorIs(t, err, domain.ErrInvalidEmail)

	// Invalid username (too short)
	_, err = domain.NewUser("uid-3", "test@amiguis.co", "ab", "hashed_pwd", "Test", "Bogotá")
	assert.ErrorIs(t, err, domain.ErrInvalidUsername)

	// Invalid username (disallowed characters)
	_, err = domain.NewUser("uid-4", "test@amiguis.co", "juan perez!", "hashed_pwd", "Test", "Barranquilla")
	assert.ErrorIs(t, err, domain.ErrInvalidUsername)
}
