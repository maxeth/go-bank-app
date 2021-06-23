package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/maxeth/go-bank-app/auth"
	db "github.com/maxeth/go-bank-app/db/sqlc"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"` // currency is a custom validation function, defined in api/validation/currency and applied inside server.go
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// get the callers username from the auth-token he sent in the headers
	authPayload := ctx.MustGet(authPayloadKey).(*auth.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	acc, err := server.repository.CreateAccount(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			// error is a pgError
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				{
					ctx.JSON(http.StatusForbidden, pqErr.Error())
					return
				}
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, acc)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	acc, err := server.repository.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))

		} else {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}

	// only return if the fetched account has the same username as the one in the auth-token
	authPayload := ctx.MustGet(authPayloadKey).(*auth.Payload)
	if acc.Owner != authPayload.Username {
		err := errors.New("not authorized to fetch this account")
		ctx.JSON(http.StatusUnauthorized, err)
		return
	}

	ctx.JSON(http.StatusOK, acc)
}

type listAccountsRequest struct {
	PageID int32 `form:"page" binding:"required,min=1"`
	Limit  int32 `form:"limit" binding:"required,min=5,max=50"`
}

func (server *Server) listAccounts(ctx *gin.Context) {
	var req listAccountsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// only list the accounts from the owner of the query
	authPayload := ctx.MustGet(authPayloadKey).(*auth.Payload)
	args := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.Limit,
		Offset: (req.PageID - 1) * req.Limit,
	}

	acc, err := server.repository.ListAccounts(ctx, args)
	// if page will be "out of reach", acc will be an empty array, because we set emit_empty_slices to true in sqlc.yml
	// but no error will be thrown
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, acc)
}
