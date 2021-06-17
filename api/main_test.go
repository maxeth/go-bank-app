package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// set gin  to test mode so it doesnt run in debug mode during test
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
