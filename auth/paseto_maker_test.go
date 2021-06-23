package auth

import (
	"testing"
	"time"

	"github.com/maxeth/go-bank-app/library"
	"github.com/o1egl/paseto"
	"github.com/stretchr/testify/require"
)

func TestPasetoMaker(t *testing.T) {
	secret := library.RandomString(32)
	maker, err := NewPasetoMaker(secret)

	require.NoError(t, err)

	username := library.RandomString(15)
	duration := time.Duration(time.Second * 10)

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := maker.CreateToken(username, duration)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)

	require.NoError(t, err)
	require.WithinDuration(t, payload.IssuedAt, issuedAt, time.Second)
	require.WithinDuration(t, payload.ExpiredAt, expiredAt, time.Second)
	require.Equal(t, payload.Username, username)
	require.NotEmpty(t, payload.ID)
}

func TestTokenExpiry(t *testing.T) {
	maker, err := NewPasetoMaker(library.RandomString(32))
	require.NoError(t, err)

	token, err := maker.CreateToken("someUsername", -time.Second)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpireToken.Error())
	require.Nil(t, payload)
}

func TestInvalidToken(t *testing.T) {
	payload, err := NewPayload(library.RandomString(20), time.Minute)
	require.NoError(t, err)

	invalidKey := []byte(library.RandomString(32))
	invalidToken, err := paseto.NewV2().Encrypt(invalidKey, payload, nil)
	require.NoError(t, err)

	// ensure our jwt implementation rejects this invalid key
	maker, err := NewPasetoMaker(library.RandomString(32))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(invalidToken)
	require.Error(t, err)
	require.Nil(t, payload)
}
