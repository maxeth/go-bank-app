package db

import (
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var (
	testQueries *Queries
)

func TestMain(m *testing.M) {

	dbConn := GetOrCreate()

	testQueries = New(dbConn.DB)

	os.Exit(m.Run())
}
