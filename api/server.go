package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/maxeth/go-bank-app/db/sqlc"
)

type Server struct {
	repository db.Repository
	router     *gin.Engine
}

func NewServer(repo db.Repository) *Server {
	server := &Server{repository: repo}
	router := gin.Default()

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
