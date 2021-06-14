package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	util "github.com/maxeth/go-bank-app/db/util"
)

func createRandomAccount(t *testing.T) Account {
	arg := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomBalance(),
		Currency: util.RandomCurrency(),
	}

	acc, err := testQueries.CreateAccount(context.Background(), arg)
	require.Nil(t, err)

	require.Equal(t, arg.Owner, acc.Owner)
	require.Equal(t, arg.Balance, acc.Balance)
	require.Equal(t, arg.Currency, acc.Currency)

	require.NotZero(t, acc.ID)
	require.NotZero(t, acc.CreatedAt)

	return acc
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	acc := createRandomAccount(t)
	res, err := testQueries.GetAccount(context.Background(), acc.ID)

	require.Nil(t, err)
	require.NotEmpty(t, res)

	require.Equal(t, res.Owner, acc.Owner)
	require.Equal(t, res.Balance, acc.Balance)
	require.Equal(t, res.Currency, acc.Currency)
}

func TestUpdateAccount(t *testing.T) {
	acc := createRandomAccount(t)

	arg := UpdateAccountBalanceParams{
		ID:      acc.ID,
		Balance: util.RandomBalance(),
	}

	res, err := testQueries.UpdateAccountBalance(context.Background(), arg)

	require.Nil(t, err)
	require.NotEmpty(t, res)

	require.NotEqual(t, acc.Balance, res.Balance)
	require.Equal(t, acc.Owner, res.Owner)
	require.Equal(t, acc.Currency, res.Currency)
	require.Equal(t, acc.CreatedAt, res.CreatedAt)
	require.Equal(t, acc.ID, res.ID)
}

func TestDeleteAccount(t *testing.T) {
	acc := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), acc.ID)
	require.Nil(t, err)

	res, err := testQueries.GetAccount(context.Background(), acc.ID)

	require.Empty(t, res)
	require.NotNil(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())

}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accs, err := testQueries.ListAccounts(context.Background(), arg)

	require.Nil(t, err)
	require.Len(t, accs, 5)

	for _, acc := range accs {
		require.NotEmpty(t, acc)
	}
}
