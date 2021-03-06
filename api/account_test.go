package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/maxeth/go-bank-app/auth"
	mockdb "github.com/maxeth/go-bank-app/db/mock"
	db "github.com/maxeth/go-bank-app/db/sqlc"
	library "github.com/maxeth/go-bank-app/library"
	"github.com/stretchr/testify/require"
)

func TestGetAccountAPI(t *testing.T) {
	// generate random acc that is then being returned by the stub
	user, _ := randomUser(t)
	acc := generateRandomAccount(user.Username)

	// define all our different test cases to get more coverage
	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, req *http.Request, tm auth.TokenMaker)
		buildStubs    func(repo *mockdb.MockRepository)
		checkResponse func(t *testing.T, resRec *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: acc.ID,
			setupAuth: func(t *testing.T, req *http.Request, tm auth.TokenMaker) {
				// create a token witht he randomly created users username so that the requests are not being rejected
				addAuthToHeader(t, req, tm, time.Minute, authTypeBearer, user.Username)
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(acc, nil)
				// expect the mock repo GetAccount method to be called once with arguments: (context, acc.ID)
				// and let it return the acc defined above and a nil err

			},
			checkResponse: func(t *testing.T, resRec *httptest.ResponseRecorder) {
				// verify that the random user we created at the top was fetched and returned
				require.Equal(t, http.StatusOK, resRec.Code)
				requireBodyAccountMatch(t, resRec.Body, acc)
			},
		},
		{
			name:      "NotFound",
			accountID: acc.ID,
			setupAuth: func(t *testing.T, req *http.Request, tm auth.TokenMaker) {
				addAuthToHeader(t, req, tm, time.Minute, authTypeBearer, user.Username)
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
				// expect the mock repo GetAccount method to be called once with arguments: (context, acc.ID)
				// and make it return an empty account struct  and a sql.ErrNoRows error
			},
			checkResponse: func(t *testing.T, resRec *httptest.ResponseRecorder) {
				// verify that the no user is being returned
				require.Equal(t, http.StatusNotFound, resRec.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountID: acc.ID,
			setupAuth: func(t *testing.T, req *http.Request, tm auth.TokenMaker) {
				addAuthToHeader(t, req, tm, time.Minute, authTypeBearer, user.Username)
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(acc.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
				// expect the mock repo GetAccount method to be called once with arguments: (context, acc.ID)
				// and make it return an empty account struct and an error different to sql.ErrNoRows which will throw an internal error in the account handler
			},
			checkResponse: func(t *testing.T, resRec *httptest.ResponseRecorder) {
				// verify that handler returned internal server error
				require.Equal(t, http.StatusInternalServerError, resRec.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: -1337,
			setupAuth: func(t *testing.T, req *http.Request, tm auth.TokenMaker) {
				addAuthToHeader(t, req, tm, time.Minute, authTypeBearer, user.Username)
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				// expect the mock repo GetAccount method to not be called at all (any, any)  because an invalid id will abrupt instantly
				// and make it return an empty account struct and an error different to sql.ErrNoRows which will throw an internal error in the account handler
			},
			checkResponse: func(t *testing.T, resRec *httptest.ResponseRecorder) {
				// verify that handler returned internal server error
				require.Equal(t, http.StatusBadRequest, resRec.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockdb.NewMockRepository(ctrl)
			tc.buildStubs(repo)

			// start server and send request
			server := newTestServer(t, repo)
			recorder := httptest.NewRecorder()

			// request finding the random user with the id of each individual test case
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, req, server.tokenMaker)
			// serve just this one request
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)

		})
	}
}

func TestListAccountsAPI(t *testing.T) {
	user, _ := randomUser(t)

	n := 5
	accs := make([]db.Account, n)
	for i := 0; i < n; i++ {
		accs[i] = generateRandomAccount(user.Username)
	}

	type Query struct {
		Page  int
		Limit int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, req *http.Request, tm auth.TokenMaker)
		buildStubs    func(repo *mockdb.MockRepository)
		checkResponse func(t *testing.T, resRec *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				Page:  1,
				Limit: 5,
			},
			setupAuth: func(t *testing.T, req *http.Request, tm auth.TokenMaker) {
				// create a token witht he randomly created users username so that the requests are not being rejected
				addAuthToHeader(t, req, tm, time.Minute, authTypeBearer, user.Username)
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				args := db.ListAccountsParams{
					Owner:  user.Username,
					Limit:  int32(n),
					Offset: 0,
				}
				repo.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(args)).
					Times(1).
					Return(accs, nil)

			},
			checkResponse: func(t *testing.T, resRec *httptest.ResponseRecorder) {
				// verify that the random user we created at the top was fetched and returned
				require.Equal(t, http.StatusOK, resRec.Code)
				requireBodyAccountsArrayMatch(t, resRec.Body, accs)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockdb.NewMockRepository(ctrl)
			tc.buildStubs(repo)

			// start server and send request
			server := newTestServer(t, repo)
			recorder := httptest.NewRecorder()

			// request finding the random user with the id of each individual test case
			url := "/accounts"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// add query params for limit and page
			q := req.URL.Query()
			q.Add("page", fmt.Sprintf("%v", tc.query.Page))
			q.Add("limit", fmt.Sprintf("%v", tc.query.Limit))
			req.URL.RawQuery = q.Encode()

			tc.setupAuth(t, req, server.tokenMaker)

			// serve just this one request
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)

		})
	}
}

func generateRandomAccount(owner string) db.Account {
	return db.Account{
		ID:       library.RandomInt(1, 100),
		Owner:    owner,
		Balance:  library.RandomBalance(),
		Currency: library.RandomCurrency(),
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

func requireBodyAccountsArrayMatch(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}
