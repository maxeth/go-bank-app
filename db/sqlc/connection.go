package db

import (
	"database/sql"

	config "github.com/maxeth/go-bank-app/config"
)

type TestDB struct {
	DB *sql.DB
}

var (
	conf   config.Config = config.New()
	testDB *sql.DB
)

func GetOrCreate() *TestDB {

	if testDB == nil || testDB.Ping() != nil {
		// not connected yet, either testDB var is nil or Ping() returned an error
		var err error
		testDB, err = sql.Open(conf.Driver, conf.ConnectionString)
		if err != nil {
			panic("cannot connect to db")
		}
		return &TestDB{DB: testDB}
	}

	// already have a connection
	return &TestDB{DB: testDB}
}
