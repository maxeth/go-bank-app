package db

import (
	"context"
	"testing"
	"time"

	auth "github.com/maxeth/go-bank-app/auth"
	library "github.com/maxeth/go-bank-app/library"

	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user := createRandomUser(t)

	res, err := testQueries.GetUser(context.Background(), user.Username)

	require.NoError(t, err)

	require.Equal(t, res.Email, user.Email)
	require.Equal(t, res.Username, user.Username)
	require.Equal(t, res.HashedPassword, user.HashedPassword)
	require.Equal(t, res.FullName, user.FullName)

	// there might be a small difference, depending on the db
	require.WithinDuration(t, res.PasswordChangedAt, user.PasswordChangedAt, time.Second)
	require.WithinDuration(t, res.CreatedAt, user.CreatedAt, time.Second)
}

func createRandomUser(t *testing.T) User {
	hashedPw, err := auth.HashPassword(library.RandomString(10))
	require.NoError(t, err)

	args := CreateUserParams{
		Username:       library.RandomOwner(),
		Email:          library.RandomOwner() + ".lastname@gmail.com",
		HashedPassword: hashedPw,
		FullName:       library.RandomOwner(),
	}

	user, err := testQueries.CreateUser(context.Background(), args)
	require.Nil(t, err)

	require.Equal(t, args.Username, user.Username)
	require.Equal(t, args.Email, user.Email)
	require.Equal(t, args.HashedPassword, user.HashedPassword)
	require.Equal(t, args.FullName, user.FullName)

	require.NotZero(t, user.CreatedAt)
	require.NotZero(t, user.PasswordChangedAt)

	return user
}

//func requireSameUser(t *testing.T, userA User, userB User) {

//require.Equal(t, userA.CreatedAt, userB.CreatedAt)
//}
