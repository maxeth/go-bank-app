package db

import (
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var (
	testQueries *Queries
)

func TestMain(m *testing.M) {

	dbConn, err := GetOrCreate()
	if err != nil {
		fmt.Println(err.Error())
		panic("Cannot test because db connection fails")
	}

	testQueries = New(dbConn.DB)

	os.Exit(m.Run())
}
