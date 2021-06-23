package main

import (
	"math/rand"
	"time"

	_ "github.com/lib/pq"
	"github.com/maxeth/go-bank-app/api"
	"github.com/maxeth/go-bank-app/config"
	db "github.com/maxeth/go-bank-app/db/sqlc"
)

const (
	address = "0.0.0.0:8080"
)

func main() {
	conf, err := config.New(".")
	if err != nil {
		panic("cannot connect to db")
	}

	conn, err := db.GetOrCreate(conf)
	if err != nil {
		panic("cannot connect to db")
	}

	repo := db.NewRepository(conn.DB)
	server, err := api.NewServer(conf, repo)
	if err != nil {
		panic("couldnt create new instance of a server")
	}

	err = server.Start(address)
	if err != nil {
		panic("server couldn't start")
	}

}

func init() {
	rand.Seed(time.Now().UnixNano())
}
