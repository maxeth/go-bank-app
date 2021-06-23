package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/maxeth/go-bank-app/auth"
	"github.com/maxeth/go-bank-app/config"
	db "github.com/maxeth/go-bank-app/db/sqlc"
)

type Server struct {
	config     config.Config
	repository db.Repository
	router     *gin.Engine
	tokenMaker auth.TokenMaker
}

func NewServer(conf config.Config, repo db.Repository) (*Server, error) {
	tokenMaker, err := auth.NewPasetoMaker(conf.TokenSummetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannt create token maker: %w", err)
	}
	server := &Server{
		config:     conf,
		repository: repo,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.applyRoutes()

	return server, nil
}

func (server *Server) applyRoutes() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	// create a group of routes that are going to be protected
	authGroup := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authGroup.POST("/accounts", server.createAccount)
	authGroup.GET("/accounts/:id", server.getAccount)
	authGroup.GET("/accounts", server.listAccounts)

	authGroup.POST("/transfers", server.createTransfer)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
