package main

import (
	_ "github.com/lib/pq"
	"github.com/maxeth/go-bank-app/api"
	db "github.com/maxeth/go-bank-app/db/sqlc"
)

const (
	address = "0.0.0.0:8080"
)

func main() {
	conn, err := db.GetOrCreate()
	if err != nil {
		panic("cannot connect to db")
	}

	repo := db.NewRepository(conn.DB)
	server := api.NewServer(repo)

	err = server.Start(address)
	if err != nil {
		panic("server couldn't start")
	}

}
