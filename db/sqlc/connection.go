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

func GetOrCreate() (*TestDB, error) {
	conf, err := config.New(".")
	if err != nil {
		return nil, err
	}

	if testDB == nil || testDB.Ping() != nil {
		// not connected yet, either testDB var is nil or Ping() returned an error
		var err error
		testDB, err = sql.Open(conf.DBDriver, conf.DBString)
		if err != nil {
			return nil, err
		}
		return &TestDB{DB: testDB}, nil
	}

	// already have a connection
	return &TestDB{DB: testDB}, nil
}
