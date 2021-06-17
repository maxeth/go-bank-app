package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/maxeth/go-bank-app/db/mock"
	db "github.com/maxeth/go-bank-app/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestCreateTransfer(t *testing.T) {
	accA := generateRandomAccount()
	accB := generateRandomAccount()
	accC := generateRandomAccount()

	transferAmount := int64(10)

	accA.Currency = "USD"
	accB.Currency = "USD"
	accC.Currency = "CAD"

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(repo *mockdb.MockRepository)
		checkResponse func(resRec *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"fromAccountID": accA.ID,
				"toAccountID":   accB.ID,
				"amount":        transferAmount,
				"currency":      accA.Currency,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				// the GetAccount method is beind called inside the handler when checking whether both accounts have the same currency / enough balance
				// for reference: /api/transfer.go
				repo.EXPECT().GetAccount(gomock.Any(), gomock.Eq(accA.ID)).Times(1).Return(accA, nil)
				repo.EXPECT().GetAccount(gomock.Any(), gomock.Eq(accB.ID)).Times(1).Return(accB, nil)

				// the TransferTx method is expected be called with these parameters internally
				args := db.TransferTxParams{
					FromAccountID: accA.ID,
					ToAccountID:   accB.ID,
					Amount:        transferAmount,
				}
				repo.EXPECT().TransferTx(gomock.Any(), gomock.Eq(args)).Times(1)
			},
			checkResponse: func(resRec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resRec.Code)
			},
		},
		{
			name: "SenderNotFound",
			body: gin.H{
				"fromAccountID": accA.ID,
				"toAccountID":   accB.ID,
				"amount":        transferAmount,
				"currency":      accA.Currency,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accA.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)

				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accB.ID)).
					Times(0)

				repo.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(resRec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, resRec.Code)
			},
		},
		{
			name: "ReceiverNotFound",
			body: gin.H{
				"fromAccountID": accA.ID,
				"toAccountID":   accB.ID,
				"amount":        transferAmount,
				"currency":      accA.Currency,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accA.ID)).
					Times(1).
					Return(accA, nil)

				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accB.ID)).
					Times(1).Return(db.Account{}, sql.ErrNoRows)

				repo.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(resRec *httptest.ResponseRecorder) {
				fmt.Println(resRec.Code)
				require.Equal(t, http.StatusNotFound, resRec.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"fromAccountID": accA.ID,
				"toAccountID":   accB.ID,
				"amount":        transferAmount,
				"currency":      "ABCDEFG",
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accA.ID)).
					Times(0)

				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accB.ID)).
					Times(0)

				repo.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(resRec *httptest.ResponseRecorder) {
				fmt.Println(resRec.Code)
				require.Equal(t, http.StatusBadRequest, resRec.Code)
			},
		},
		{
			name: "ReceiverCurrencyMismatch",
			body: gin.H{
				"fromAccountID": accA.ID,
				"toAccountID":   accC.ID,
				"amount":        transferAmount,
				"currency":      "USD",
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accA.ID)).
					Times(1).
					Return(accA, nil)

				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accC.ID)).
					Times(1).
					Return(accC, nil)

				repo.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(resRec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, resRec.Code)
			},
		},
		{
			name: "SenderCurrencyMismatch",
			body: gin.H{
				"fromAccountID": accA.ID,
				"toAccountID":   accC.ID,
				"amount":        transferAmount,
				"currency":      "CAD",
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accA.ID)).
					Times(1).
					Return(accA, nil)

				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accC.ID)).
					Times(0)

				repo.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(resRec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, resRec.Code)
			},
		}, {
			name: "SendingNegativeAmount",
			body: gin.H{
				"fromAccountID": accA.ID,
				"toAccountID":   accB.ID,
				"amount":        int64(-transferAmount), // sending 1 more than the account actually has
				"currency":      accA.Currency,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accA.ID)).
					Times(0)

				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accB.ID)).
					Times(0)

				repo.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(resRec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, resRec.Code)
			},
		},
		{
			name: "SendingTooMuch",
			body: gin.H{
				"fromAccountID": accA.ID,
				"toAccountID":   accB.ID,
				"amount":        accA.Balance + int64(1), // sending 1 more than the account actually has
				"currency":      accA.Currency,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accA.ID)).
					Times(1).
					Return(accA, nil)

				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accB.ID)).
					Times(0)

				repo.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(resRec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, resRec.Code)
			},
		}, {
			name: "SendingEverything",
			body: gin.H{
				"fromAccountID": accA.ID,
				"toAccountID":   accB.ID,
				"amount":        accA.Balance, // sending 1 more than the account actually has
				"currency":      accA.Currency,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accA.ID)).
					Times(1).
					Return(accA, nil)

				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(accB.ID)).
					Times(1).Return(accB, nil)

				repo.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1)
			},
			checkResponse: func(resRec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resRec.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		fmt.Println(i)

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockdb.NewMockRepository(ctrl)
			tc.buildStubs(repo)

			server := NewServer(repo)
			recorder := httptest.NewRecorder()

			// prepare/marshal request data for the http POST request
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/transfers", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)

			tc.checkResponse(recorder)
		})
	}
}
