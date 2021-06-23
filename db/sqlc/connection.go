package db

import (
	"database/sql"

	config "github.com/maxeth/go-bank-app/config"
)

type TestDB struct {
	DB *sql.DB
}

var (
	testDB *sql.DB
)

func GetOrCreate(c config.Config) (*TestDB, error) {
	if testDB == nil || testDB.Ping() != nil {
		// not connected yet, either testDB var is nil or Ping() returned an error
		var err error
		testDB, err = sql.Open(c.DBDriver, c.DBString)
		if err != nil {
			return nil, err
		}
		return &TestDB{DB: testDB}, nil
	}

	// already have a connection
	return &TestDB{DB: testDB}, nil
}
