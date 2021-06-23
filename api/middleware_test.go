package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maxeth/go-bank-app/auth"
	"github.com/stretchr/testify/require"
)

// adds authorization headers to passed http request in expected format, including a token created form the passed username
func addAuthToHeader(
	t *testing.T,
	req *http.Request,
	tm auth.TokenMaker,
	dur time.Duration,
	authType string,
	username string,

) {
	token, err := tm.CreateToken(username, dur)
	require.NoError(t, err)

	authHeader := fmt.Sprintf("%v %v", authType, token)
	req.Header.Set(authHeaderKey, authHeader)
}

func TestAuthMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, req *http.Request, tm auth.TokenMaker) // for example: creates token and adds it to auth headar
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, req *http.Request, tm auth.TokenMaker) {
				addAuthToHeader(t, req, tm, time.Minute, authTypeBearer, "user")
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, req *http.Request, tm auth.TokenMaker) {
				addAuthToHeader(t, req, tm, -time.Second, authTypeBearer, "user")
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{

			name: "NoAuth",
			setupAuth: func(t *testing.T, req *http.Request, tm auth.TokenMaker) {
				// add no auth header
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
		{
			name: "WrongAuthType",
			setupAuth: func(t *testing.T, req *http.Request, tm auth.TokenMaker) {
				addAuthToHeader(t, req, tm, time.Minute, "basic", "user")
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		}, {
			name: "NoAuthType",
			setupAuth: func(t *testing.T, req *http.Request, tm auth.TokenMaker) {
				addAuthToHeader(t, req, tm, time.Minute, "", "user")
			},
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, rec.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)

			authPath := "/auth"
			server.router.GET(
				authPath,
				authMiddleware(server.tokenMaker), // apply the tested middleware
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{}) // dummy handler
				},
			)

			rec := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			// modify http header
			tc.setupAuth(t, req, server.tokenMaker)

			server.router.ServeHTTP(rec, req)

			// check the response status
			tc.checkResponse(t, rec)
		})

	}
}
