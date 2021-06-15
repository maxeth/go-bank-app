package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	mockdb "github.com/maxeth/go-bank-app/db/mock"
	db "github.com/maxeth/go-bank-app/db/sqlc"
	rnd "github.com/maxeth/go-bank-app/db/util"
	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T) {
	// generate random acc that is then being returned by the stub
	acc := generateRandomAccount()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockRepository(ctrl)

	// build stubs
	store.EXPECT().
		GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
		Times(1).
		Return(acc, nil)
	// -> "Expect the GetAccount method to be called once with any context and the account id, and return this account with a nil err"
	// In case any of these expectations don't match exactly, the test fails

	// start server and send request
	server := NewServer(store)
	recorder := httptest.NewRecorder()

	// request finding the random user we created at the top
	url := fmt.Sprintf("/accounts/%d", acc.ID)
	req, err := http.NewRequest(http.MethodGet, url, nil)

	require.NoError(t, err)

	// serve just this one request
	server.router.ServeHTTP(recorder, req)

	// verify that the random user we created at the top was fetched
	require.Equal(t, http.StatusOK, recorder.Code)
	requireBodyAccountMatch(t, recorder.Body, acc)
}

func generateRandomAccount() db.Account {
	return db.Account{
		ID:      rnd.RandomInt(1, 100),
		Owner:   rnd.RandomOwner(),
		Balance: rnd.RandomBalance(),
	}
}

func requireBodyAccountMatch(t *testing.T, body *bytes.Buffer, got db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var have db.Account
	err = json.Unmarshal(data, &have)
	require.NoError(t, err)

	require.Equal(t, got, have)
}
