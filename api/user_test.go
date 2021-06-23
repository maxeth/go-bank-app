package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	mockdb "github.com/maxeth/go-bank-app/db/mock"
	db "github.com/maxeth/go-bank-app/db/sqlc"
	library "github.com/maxeth/go-bank-app/library"
	"github.com/stretchr/testify/require"

	auth "github.com/maxeth/go-bank-app/auth"
)

type eqCreateUserParamsMatcher struct {
	args     db.CreateUserParams
	password string
}

// In order to create a custom go-mock matcher, we need a struct that implemenets the matcher interface
// which requires it to have a "Matches" and a "String" method
func (eq eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	args, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	// a CreateUserParams struct was passed. check whether the password is the same by using the checkpassword method
	err := auth.CheckPassword(args.HashedPassword, eq.password)
	if err != nil {
		return false
	}

	// set matcher struct hash possword to the passed hash password as it is valid
	// eq.args.HashPassword was previously = ""
	eq.args.HashedPassword = args.HashedPassword

	// check whether also the rest of the createUserParams matches
	return reflect.DeepEqual(eq.args, args)
}

func (eq eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", eq.args, eq.password)
}

// returns the struct that implements the  mock.Matcher  interface
// this struts "Matches" method will be called whenever a check is being performed
func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	// it has the password saved so it can be checked against the input password without having to
	// are being generated randomly directly compare the hashes which wouldnt work since the hashes
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(repo *mockdb.MockRepository)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullName": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				repo.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)). // simple equal check wouldnt work since hashed pw is always different
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireResponseBodyMatch(t, recorder.Body, user)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullName": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullName": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: "23505"}) // expecing a code 23505 error: unique_violation
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": "invalid-user#1",
				"password": password,
				"fullName": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullName": user.FullName,
				"email":    "invalid-email",
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"username": user.Username,
				"password": "123",
				"fullName": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			server := newTestServer(t, repo)
			if server == nil {
				panic("cannot connect server")
			}
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestLoginUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(repo *mockdb.MockRepository)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "WrongPassword",
			body: gin.H{
				"username": user.Username,
				"password": "123456",
			},
			buildStubs: func(repo *mockdb.MockRepository) {
				repo.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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

			server := newTestServer(t, repo)
			if server == nil {
				panic("cannot connect server")
			}
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users/login"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = library.RandomString(15)
	hashedPw, err := auth.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       library.RandomString(10),
		Email:          library.RandomString(15) + ".lastname@gmail.com",
		HashedPassword: hashedPw,
		FullName:       library.RandomString(5),
	}

	return
}

// require the response body of a server response matches a given user struct
func requireResponseBodyMatch(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User

	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)

	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)

	// the hashed pw shouldnt be returned from the server
	require.Empty(t, gotUser.HashedPassword)
}
