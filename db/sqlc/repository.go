package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	// Without saying Queries *Queries, we will be able to directly call Queries methods on the Repository struct instead of having to preofix Queries
	// "Inheritcance"
	*Queries
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db:      db,
		Queries: New(db),
	}
}

func (repo *Repository) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	query := New(tx)

	// call higher order function passed as argument
	err = fn(query)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			// transaction failed, couldn't rollback
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		// transaction failed, rollback was successful
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"fromAccountID"`
	ToAccountID   int64 `json:"toAccountID"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"fromAccount"`
	ToAccount   Account  `json:"toAccount"`
	FromEntry   Entry    `json:"fromEntry"`
	ToEntry     Entry    `json:"toEntry"`
}

// Transfer creates a money Transfer from a sender to a receiver account
// More specifically, Transfer creates a transfer, from-entry and to-entry SQL record as part of a single SQL transaction
func (repo *Repository) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult
	var err error

	err = repo.execTx(ctx, func(q *Queries) error {
		// this is the anonymous higher order function that is being called inside execTx as part of the transcation

		trArgs := CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		}
		result.Transfer, err = q.CreateTransfer(ctx, trArgs)
		// note how even though this function is being called "inside" execTx as a higher order function,
		// it accesses result.Transfer which makes it a Closure.
		// https://gobyexample.com/closures
		if err != nil {
			return err
		}

		fArgs := CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    -arg.Amount,
		}
		result.FromEntry, err = q.CreateEntry(ctx, fArgs)
		if err != nil {
			return err
		}

		tArgs := CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		}

		result.ToEntry, err = q.CreateEntry(ctx, tArgs)
		if err != nil {
			return err
		}

		// arrange the order of accounts inside the sql transcations based on the account id
		// this is necessary to prevent a deadlock. all transaction operations should follow this pattern
		// of ordering by some unique key such as the id
		var args AddMoneyParams

		if arg.FromAccountID < arg.ToAccountID {
			args = AddMoneyParams{arg.ToAccountID, arg.FromAccountID, arg.Amount, -arg.Amount}
		} else {
			args = AddMoneyParams{arg.FromAccountID, arg.ToAccountID, -arg.Amount, arg.Amount}
		}

		result.ToAccount, result.FromAccount, err = updateTransferBalances(ctx, q, args)
		if err != nil {
			return err
		}
		return nil
	})

	return result, err
}

type AddMoneyParams struct {
	accAID,
	accBID,
	amountA,
	amountB int64
}

// this function dynamically updates user balances in  order to prevent a deadlock. we always pass the account
// with the smaller id as accA, and the one with the bigger id as accB
func updateTransferBalances(ctx context.Context, q *Queries, args AddMoneyParams) (accA, accB Account, err error) {
	accA, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     args.accAID,
		Amount: args.amountA,
	})
	if err != nil {
		return
	}

	accB, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     args.accBID,
		Amount: args.amountB,
	})
	if err != nil {
		return
	}

	return // no need to return any parameters because they are already being assigned above. this is a feature in go.
}
