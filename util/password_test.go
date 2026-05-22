package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	password := RandomString(8)
	hashedPassword, err := HashPassword(password)

	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	err = ValidatePassword(password, hashedPassword)
	require.NoError(t, err)
}

func TestWrongPassword(t *testing.T) {
	password := RandomString(8)
	wrongPassword := RandomString(8)
	hashedPassword, err := HashPassword(password)

	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	err = ValidatePassword(wrongPassword, hashedPassword)
	require.Error(t, err)
}
