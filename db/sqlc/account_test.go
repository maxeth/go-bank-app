package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	library "github.com/maxeth/go-bank-app/library"
)

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
		Balance: library.RandomBalance(),
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
	var lastAcc Account
	for i := 0; i < 10; i++ {
		lastAcc = createRandomAccount(t)
	}

	arg := ListAccountsParams{
		Owner:  lastAcc.Owner,
		Limit:  5,
		Offset: 0,
	}

	accs, err := testQueries.ListAccounts(context.Background(), arg)

	require.Nil(t, err)
	require.NotEmpty(t, accs)

	for _, acc := range accs {
		require.NotEmpty(t, acc)
		require.Equal(t, lastAcc.Owner, acc.Owner)
	}
}
func createRandomAccount(t *testing.T) Account {
	user := createRandomUser(t)

	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  library.RandomBalance(),
		Currency: library.RandomCurrency(),
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
