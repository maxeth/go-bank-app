package auth

import (
	"testing"

	"github.com/maxeth/go-bank-app/library"
	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	pw := library.RandomString(6)

	hashedPw, err := HashPassword(pw)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPw)

	err = CheckPassword(hashedPw, pw)
	require.NoError(t, err)

	// invalid pw should not pass
	invalidPw := library.RandomString(6)

	err = CheckPassword(hashedPw, invalidPw)
	require.Error(t, err)

	// the same passwort should always return a different hash because of random salt generation as part of bcrypt inplementation
	hashedPw2, err := HashPassword(pw)
	require.NoError(t, err)
	require.NotEqual(t, hashedPw, hashedPw2)
}
