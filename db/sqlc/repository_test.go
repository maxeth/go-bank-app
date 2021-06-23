package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	repo := NewRepository(testDB)
	accA := createRandomAccount(t)

	accB := createRandomAccount(t)

	// channels to check whether goroutines passed test
	errs := make(chan error)
	results := make(chan TransferTxResult)

	const txAmount int64 = 50
	const txCount int64 = 5
	for i := 0; i < int(txCount); i++ {
		go func() {
			txParams := TransferTxParams{FromAccountID: accA.ID, ToAccountID: accB.ID, Amount: txAmount}
			result, err := repo.TransferTx(context.Background(), txParams)
			errs <- err
			results <- result
		}()
	}

	for i := 0; i < int(txCount); i++ {
		//	fmt.Println(">> count: ", i)
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		// assert transfer records
		trf := result.Transfer

		require.NotEmpty(t, trf)

		require.Equal(t, accA.ID, trf.FromAccountID)
		require.Equal(t, accB.ID, trf.ToAccountID)
		require.Equal(t, txAmount, trf.Amount)

		require.NotZero(t, trf.ID, trf.CreatedAt)

		_, err = repo.GetTransfer(context.Background(), trf.ID)
		require.NoError(t, err)

		// assert from-entry records
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, accA.ID, fromEntry.AccountID)
		require.Equal(t, -txAmount, fromEntry.Amount) // amount negative because of money outflow
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = repo.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		// assert to-entry records
		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, accB.ID, toEntry.AccountID)
		require.Equal(t, txAmount, toEntry.Amount) // amount positive because of money inflow
		require.NotZero(t, toEntry.CreatedAt)

		var entr Entry
		entr, err = repo.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)
		require.NotEmpty(t, entr)
		require.Equal(t, entr.Amount, txAmount)

		// check whether account changed
		fromAcc := result.FromAccount
		toAcc := result.ToAccount
		require.NotEmpty(t, fromAcc, toAcc)
		require.Equal(t, fromAcc.ID, accA.ID)
		require.Equal(t, toAcc.ID, accB.ID)

		// check whether the toAccount received the correct amount of money
		// lets say accA with 100 bal sent 20 to accB: diffA = 100 - 80 = 20, diffB = 120 - 100 = 20
		diffA := accA.Balance - fromAcc.Balance
		diffB := toAcc.Balance - accB.Balance
		require.Equal(t, diffA, diffB)
		require.True(t, diffA > 0, diffB > 0)
	}

	// check final result
	updatedA, err := testQueries.GetAccount(context.Background(), accA.ID)
	require.NoError(t, err)

	updatedB, err := testQueries.GetAccount(context.Background(), accB.ID)
	require.NoError(t, err)

	//fmt.Println(">> ", accA.Balance, accB.Balance)

	require.Equal(t, updatedA.Balance, accA.Balance-txCount*txAmount)
	require.Equal(t, updatedB.Balance, accB.Balance+txCount*txAmount)
}
