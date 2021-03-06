package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maxeth/go-bank-app/auth"
	db "github.com/maxeth/go-bank-app/db/sqlc"
)

type createTransferRequest struct {
	FromID   int64  `json:"fromAccountID" binding:"required,min=1"`
	ToID     int64  `json:"toAccountID" binding:"required,min=1"`
	Amount   int64  `json:"amount" binding:"required,min=0"`      // 100 is 1.00[currency], so 1 would be 1 cent in case the currency is divisable
	Currency string `json:"currency" binding:"required,currency"` // currency validation method is implemented in api/validation/currency.go
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authPayloadKey).(*auth.Payload)

	// validate that sender and receiver use the same currencies. Optionally add some conversion later on
	// check whether sender account sends the right currency and whether he has enough balance to perform the transfer
	isValidFrom, accFrom := server.checkValidAccount(ctx, req.FromID, req.Currency)
	if !isValidFrom {
		return
	}
	if accFrom.Owner != authPayload.Username {
		// sender isnt the owner of the account he is trying to send money from
		err := errors.New("not authorized to make this tansfer")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	if accFrom.Balance-req.Amount < 0 {
		err := errors.New("sender does not have enough balance to perform this transfer")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// if receiver has different currency, cancel
	isValidTo, accTo := server.checkValidAccount(ctx, req.ToID, req.Currency)
	// ensure sender isnt the same acc as receiver
	if accTo.Owner == authPayload.Username {
		err := errors.New("cannot make transfer to your own account")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if !isValidTo {
		return
	}

	arg := db.TransferTxParams{FromAccountID: req.FromID, ToAccountID: req.ToID, Amount: req.Amount}

	// execute transfer transcation repository method
	trf, err := server.repository.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, trf)
}

// function checks whether the passed account id has the passed currency as primary currency set, and returns the account
func (server *Server) checkValidAccount(ctx *gin.Context, id int64, curr string) (bool, db.Account) {
	acc, err := server.repository.GetAccount(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return false, db.Account{}
	}
	if acc.Currency != curr {
		err = fmt.Errorf("invalid currency for account [%d]: expected %s received %s", acc.ID, acc.Currency, curr)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return false, db.Account{}
	}

	return true, acc
}
