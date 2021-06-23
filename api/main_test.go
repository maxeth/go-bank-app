package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maxeth/go-bank-app/config"
	db "github.com/maxeth/go-bank-app/db/sqlc"
	"github.com/maxeth/go-bank-app/library"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// set gin  to test mode so it doesnt run in debug mode during test
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}

func newTestServer(t *testing.T, repo db.Repository) *Server {
	conf := config.Config{
		TokenSummetricKey:   library.RandomString(32),
		AccessTokenDuration: time.Second * 15,
	}
	server, err := NewServer(conf, repo)
	require.NoError(t, err)
	require.NotNil(t, server)

	return server
}
