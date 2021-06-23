package auth

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/maxeth/go-bank-app/library"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	secret := library.RandomString(50)
	maker, err := NewJWTMaker(secret)

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

func TestJWTExpiry(t *testing.T) {
	maker, err := NewJWTMaker(library.RandomString(50))
	require.NoError(t, err)

	token, err := maker.CreateToken("someUsername", -time.Second)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpireToken.Error())
	require.Nil(t, payload)
}

func TestInvalidJWT(t *testing.T) {
	payload, err := NewPayload(library.RandomString(20), time.Minute)
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload) // no signing method specified
	signedToken, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	// ensure our jwt implementation rejects this JWT token which specifies no signing algorithm
	maker, err := NewJWTMaker(library.RandomString(50))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(signedToken)
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}
