package db

import (
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/maxeth/go-bank-app/config"
)

var (
	testQueries *Queries
)

func TestMain(m *testing.M) {

	c, err := config.New("../../")
	if err != nil {
		panic("couldn't load config env file")
	}

	dbConn, err := GetOrCreate(c)
	if err != nil {
		panic("Cannot test because db connection fails")
	}

	testQueries = New(dbConn.DB)

	os.Exit(m.Run())
}
