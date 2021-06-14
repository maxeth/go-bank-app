package db

import (
	"context"
	"testing"
	"time"

	db "github.com/maxeth/go-bank-app/db/util"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T, from, to Account) Transfer {
	args := CreateTransferParams{
		FromAccountID: from.ID,
		ToAccountID:   to.ID,
		Amount:        db.RandomMoney(),
	}

	trf, err := testQueries.CreateTransfer(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, trf)

	require.Equal(t, args.FromAccountID, trf.FromAccountID)
	require.Equal(t, args.ToAccountID, trf.ToAccountID)
	require.Equal(t, args.Amount, trf.Amount)

	require.NotZero(t, trf.ID)
	require.NotZero(t, trf.CreatedAt)

	return trf
}

func TestCreateTransfer(t *testing.T) {
	accA := createRandomAccount(t)
	accB := createRandomAccount(t)
	createRandomTransfer(t, accA, accB)
}

func TestGetTransfer(t *testing.T) {
	accA := createRandomAccount(t)
	accB := createRandomAccount(t)
	trf := createRandomTransfer(t, accA, accB)

	transfer, err := testQueries.GetTransfer(context.Background(), trf.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)
	require.Equal(t, trf.ID, transfer.ID)
	require.Equal(t, trf.FromAccountID, transfer.FromAccountID)
	require.Equal(t, trf.ToAccountID, transfer.ToAccountID)

	require.WithinDuration(t, trf.CreatedAt, transfer.CreatedAt, time.Second)
}

func TestListeTransfers(t *testing.T) {
	accA := createRandomAccount(t)
	accB := createRandomAccount(t)

	// create 20 random transfers to retreive
	for i := 0; i < 4; i++ {
		createRandomTransfer(t, accA, accB)
		createRandomTransfer(t, accB, accA)
	}

	args := ListTransfersParams{FromAccountID: accA.ID, ToAccountID: accB.ID, Limit: 3, Offset: 1}

	transfers, err := testQueries.ListTransfers(context.Background(), args)
	require.NoError(t, err)
	require.Len(t, transfers, 3)

	for _, trf := range transfers {
		require.NotEmpty(t, trf)
		require.True(t, trf.FromAccountID == accA.ID || trf.ToAccountID == accB.ID)
	}

	args2 := ListTransfersParams{FromAccountID: accA.ID, ToAccountID: accB.ID, Limit: 3, Offset: 4}

	empty, err := testQueries.ListTransfers(context.Background(), args2)
	require.NoError(t, err)
	require.Len(t, empty, 0) // Offet of 4, even though only 4 entires match the args, should return 0 rows
}
